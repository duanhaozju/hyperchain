//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"hyperchain/common"
	"hyperchain/crypto/primitives"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"hyperchain/common/interface"
)

const (
	JSONRPCVersion = "2.0"
)

// JSON-RPC request
type JSONRequest struct {
	Method    string          `json:"method"`
	Version   string          `json:"jsonrpc"`
	Namespace string          `json:"namespace"`
	Id        json.RawMessage `json:"id,omitempty"`
	Payload   json.RawMessage `json:"params,omitempty"`
}

// JSON-RPC response
type JSONResponse struct {
	Version   string      `json:"jsonrpc"`
	Namespace string      `json:"namespace,omitempty"`
	Id        interface{} `json:"id,omitempty"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Result    interface{} `json:"result,omitempty"`
	Info      interface{} `json:"info,omitempty"`
}

// JSON-RPC notification payload
type jsonSubscription struct {
	Event        string      `json:"event"`
	Subscription string      `json:"subscription"`
	Data         interface{} `json:"data,omitempty"`
}

// JSON-RPC notification
type jsonNotification struct {
	Version   string           `json:"jsonrpc"`
	Namespace string           `json:"namespace"`
	Result    jsonSubscription `json:"result"`
}

// jsonCodec reads and writes JSON-RPC messages to the underlying connection. It
// also has support for parsing arguments and serializing (result) objects.
type jsonCodecImpl struct {
	closer sync.Once          // close closed channel once
	closed chan interface{}   // closed on Close
	decMu  sync.Mutex         // guards d
	d      *json.Decoder      // decodes incoming requests
	encMu  sync.Mutex         // guards e
	e      *json.Encoder      // encodes responses
	rw     io.ReadWriteCloser // connection
	req    *http.Request
	nsMgrProcessor intfc.NsMgrProcessor
	conn   *websocket.Conn
}

// NewJSONCodec creates a new RPC server codec with support for JSON-RPC 2.0
func NewJSONCodec(rwc io.ReadWriteCloser, req *http.Request, nsMgrProcessor intfc.NsMgrProcessor, conn *websocket.Conn) ServerCodec {
	d := json.NewDecoder(rwc)
	d.UseNumber()
	return &jsonCodecImpl{
		closed: 	make(chan interface{}),
		d:     		d,
		e:      	json.NewEncoder(rwc),
		rw:     	rwc,
		req:    	req,
		nsMgrProcessor: nsMgrProcessor,
		conn:   	conn,
	}
}

// CheckHttpHeaders will check http header. If it is verified, client has access to interact with the server,
// otherwise, unauthorized error will be returned.
func (c *jsonCodecImpl) CheckHttpHeaders(namespace string, method string) common.RPCError {
	ns := c.nsMgrProcessor.GetNamespaceProcessorName(namespace)
	if ns == nil {
		return &common.NamespaceNotFound{Name: namespace}
	}

	cm := ns.GetCAManager()
	if !cm.IsCheckTCert() {
		return nil
	}

	c.decMu.Lock()
	defer c.decMu.Unlock()

	// parse certificate
	tcertPem := common.TransportDecode(c.req.Header.Get("tcert"))
	tcert, err := primitives.ParseCertificate([]byte(tcertPem))
	if err != nil {
		log.Error("fail to parse tcert.", err)
		return &common.UnauthorizedError{}
	}

	// verify signature
	pubKey := tcert.PublicKey.(*(ecdsa.PublicKey))
	signature := c.req.Header.Get("signature")
	msg := common.TransportDecode(c.req.Header.Get("msg"))
	signB := common.Hex2Bytes(signature)

	verifySignature, err := primitives.ECDSAVerifyTransport(pubKey, []byte(msg), signB)
	if err != nil || !verifySignature {
		log.Error("Fail to verify Transport Signture!", err)
		return &common.UnauthorizedError{}
	}

	//// verfiy tcert
	//verifyTcert, err := cm.VerifyTCert(tcertPem, method)
	//if verifyTcert == false || err != nil {
	//	log.Error("Fail to verify tcert!", err)
	//	return &common.UnauthorizedError{}
	//}
	return nil
}

// ReadRawRequest will read new requests without parsing the arguments. It will
// return a collection of requests, an indication if these requests are in batch
// form or an error when the incoming message could not be read/parsed.
func (c *jsonCodecImpl) ReadRawRequest(options CodecOption) ([]*common.RPCRequest, bool, common.RPCError) {
	c.decMu.Lock()
	defer c.decMu.Unlock()

	var incomingMsg json.RawMessage
	if err := c.d.Decode(&incomingMsg); err != nil {
		return nil, false, &common.InvalidRequestError{Message: err.Error()}
	}
	if isBatch(incomingMsg) {
		return parseBatchRequest(incomingMsg)
	}

	return parseRequest(incomingMsg, options)
}

// GatAuthInfo read authentication info (token and method) from http header.
func (c *jsonCodecImpl) GetAuthInfo() (string, string) {
	token := c.req.Header.Get("Authorization")
	method := c.req.Header.Get("Method")
	return token, method
}

// isBatch returns true when the first non-whitespace characters is '['
func isBatch(msg json.RawMessage) bool {
	for _, c := range msg {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		if c == 0x20 || c == 0x09 || c == 0x0a || c == 0x0d {
			continue
		}
		return c == '['
	}
	return false
}

// checkReqId returns an error when the given reqId isn't valid for RPC method calls.
// valid id's are strings, numbers or null
func checkReqId(reqId json.RawMessage) error {
	if len(reqId) == 0 {
		return errors.New("missing request id")
	}
	if _, err := strconv.ParseFloat(string(reqId), 64); err == nil {
		return nil
	}
	var str string
	if err := json.Unmarshal(reqId, &str); err == nil {
		return nil
	}
	return errors.New("invalid request id")
}

// parseRequest will parse a single request from the given RawMessage. It will return
// the parsed request, an indication if the request was a batch or an error when
// the request could not be parsed.
func parseRequest(incomingMsg json.RawMessage, options CodecOption) ([]*common.RPCRequest, bool, common.RPCError) {
	var in JSONRequest
	if err := json.Unmarshal(incomingMsg, &in); err != nil {
		return nil, false, &common.InvalidMessageError{Message: err.Error()}
	}
	if err := checkReqId(in.Id); err != nil {
		return nil, false, &common.InvalidMessageError{Message: err.Error()}
	}

	// subscribe are special, they will always use `subscribeMethod` as first param in the payload
	if strings.HasSuffix(in.Method, common.SubscribeMethodSuffix) {
		if options == OptionMethodInvocation {
			return nil, false, &common.CallbackError{Message: ErrNotificationsUnsupported.Error()}
		}
		reqs := []*common.RPCRequest{{Id: &in.Id, IsPubSub: true}}
		if len(in.Payload) > 0 {
			// first param must be subscription name
			var subscribeMethod [1]string
			if err := json.Unmarshal(in.Payload, &subscribeMethod); err != nil {
				log.Debug(fmt.Sprintf("Unable to parse subscription method: %v\n", err))
				return nil, false, &common.InvalidRequestError{Message: "Unable to parse subscription request"}
			}
			if subscribeMethod[0] == "" {
				return nil, false, &common.InvalidParamsError{Message: "Please give a subscription name as the first param"}
			}

			reqs[0].Service, reqs[0].Method = strings.TrimSuffix(in.Method, common.SubscribeMethodSuffix), subscribeMethod[0]
			reqs[0].Params = in.Payload
			return reqs, false, nil
		}
		return nil, false, &common.InvalidRequestError{Message: "Unable to parse subscription request"}
	}

	if strings.HasSuffix(in.Method, common.UnsubscribeMethodSuffix) {
		return []*common.RPCRequest{{Id: &in.Id, IsPubSub: true,
			Method: in.Method, Params: in.Payload}}, false, nil
	}

	// regular RPC call
	elems := strings.Split(in.Method, common.ServiceMethodSeparator)
	if len(elems) != 2 {
		return nil, false, &common.MethodNotFoundError{Service: in.Method, Method: ""}
	}

	if len(in.Payload) == 0 {
		return []*common.RPCRequest{{Service: elems[0], Method: elems[1], Namespace: in.Namespace, Id: &in.Id}}, false, nil
	}

	return []*common.RPCRequest{{Service: elems[0], Method: elems[1], Namespace: in.Namespace, Id: &in.Id, Params: in.Payload}}, false, nil
}

// parseBatchRequest will parse a batch request into a collection of requests from the given RawMessage, an indication
// if the request was a batch or an error when the request could not be read.
func parseBatchRequest(incomingMsg json.RawMessage) ([]*common.RPCRequest, bool, common.RPCError) {
	var in []JSONRequest
	if err := json.Unmarshal(incomingMsg, &in); err != nil {
		return nil, false, &common.InvalidMessageError{Message: err.Error()}
	}

	requests := make([]*common.RPCRequest, len(in))
	for i, r := range in {
		if err := checkReqId(r.Id); err != nil {
			return nil, false, &common.InvalidMessageError{Message: err.Error()}
		}

		id := &in[i].Id

		elems := strings.Split(r.Method, common.ServiceMethodSeparator)
		if len(elems) != 2 {
			return nil, true, &common.MethodNotFoundError{Service: r.Method, Method: ""}
		}

		if len(r.Payload) == 0 {
			requests[i] = &common.RPCRequest{Service: elems[0], Method: elems[1], Id: id, Params: nil}
		} else {
			requests[i] = &common.RPCRequest{Service: elems[0], Method: elems[1], Id: id, Params: r.Payload}
		}
	}

	return requests, true, nil
}

// CreateResponse will create a JSON-RPC success response with the given id and reply as result.
func (c *jsonCodecImpl) CreateResponse(id interface{}, namespace string, reply interface{}) interface{} {
	if IsHexNum(reflect.TypeOf(reply)) {
		return &JSONResponse{Version: JSONRPCVersion, Namespace: namespace, Id: id, Code: 0, Message: "SUCCESS", Result: fmt.Sprintf(`%#x`, reply)}
	}
	return &JSONResponse{Version: JSONRPCVersion, Namespace: namespace, Id: id, Code: 0, Message: "SUCCESS", Result: reply}
}

// CreateErrorResponse will create a JSON-RPC error response with the given id and error.
func (c *jsonCodecImpl) CreateErrorResponse(id interface{}, namespace string, err common.RPCError) interface{} {
	return &JSONResponse{Version: JSONRPCVersion, Namespace: namespace, Id: id, Code: err.Code(), Message: err.Error()}
}

// CreateErrorResponseWithInfo will create a JSON-RPC error response with the given id and error.
// info is optional and contains additional information about the error. When an empty string is passed it is ignored.
func (c *jsonCodecImpl) CreateErrorResponseWithInfo(id interface{}, namespace string, err common.RPCError, info interface{}) interface{} {
	return &JSONResponse{Version: JSONRPCVersion, Namespace: namespace, Id: id, Code: err.Code(), Message: err.Error(), Info: info}
}

// CreateNotification will create a JSON-RPC notification with the given subscription id and event as params.
func (s *jsonCodecImpl) CreateNotification(subid common.ID, service, method, namespace string, event interface{}) interface{} {
	if IsHexNum(reflect.TypeOf(event)) {
		return &jsonNotification{Version: JSONRPCVersion, Namespace: namespace,
			Result: jsonSubscription{Subscription: fmt.Sprintf(`%s`, subid), Data: fmt.Sprintf(`%#x`, event)}}
	}

	return &jsonNotification{Version: JSONRPCVersion, Namespace: namespace,
		Result: jsonSubscription{Event: method, Subscription: fmt.Sprintf(`%s`, subid), Data: event}}
}

// Write will write response to client.
func (c *jsonCodecImpl) Write(res interface{}) error {
	c.encMu.Lock()
	defer c.encMu.Unlock()

	return c.e.Encode(res)
}

// WriteNotify will write websocket message to client.
func (c *jsonCodecImpl) WriteNotify(res interface{}) error {
	c.encMu.Lock()
	defer c.encMu.Unlock()

	nw, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Error(err)
		return err
	}

	if b, err := json.Marshal(res); err != nil {
		log.Error(err)
		return err
	} else {
		if _, err = nw.Write(b); err != nil {
			log.Error(err)
			return err
		}
	}

	if err := nw.Close(); err != nil {
		log.Error(err)
		return err
	}
	log.Debug("** finish writting notification to client **")
	return nil
}

// Close will close the underlying connection.
func (c *jsonCodecImpl) Close() {
	c.closer.Do(func() {
		close(c.closed)
		c.rw.Close()
	})
}

// Closed returns a channel which will be closed when Close is called.
func (c *jsonCodecImpl) Closed() <-chan interface{} {
	return c.closed
}
