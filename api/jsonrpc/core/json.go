//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"encoding/json"
	"hyperchain/common"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	//"crypto/ecdsa"
	//"hyperchain/core/crypto/primitives"
	"hyperchain/namespace"
	"github.com/pkg/errors"
)

const (
	JSONRPCVersion         = "2.0"
	serviceMethodSeparator = "_"
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
	Namespace string      `json:"namespace"`
	Id        interface{} `json:"id,omitempty"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Result    interface{} `json:"result,omitempty"`
}

// JSON-RPC notification payload
type jsonSubscription struct {
	Subscription string      `json:"subscription"`
	Result       interface{} `json:"result,omitempty"`
}

// JSON-RPC notification
type jsonNotification struct {
	Version string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  jsonSubscription `json:"params"`
}

// jsonCodec reads and writes JSON-RPC messages to the underlying connection. It
// also has support for parsing arguments and serializing (result) objects.
type jsonCodec struct {
	closer sync.Once          // close closed channel once
	closed chan interface{}   // closed on Close
	decMu  sync.Mutex         // guards d
	d      *json.Decoder      // decodes incoming requests
	encMu  sync.Mutex         // guards e
	e      *json.Encoder      // encodes responses
	rw     io.ReadWriteCloser // connection
	req    *http.Request
	nr     namespace.NamespaceManager
}

// NewJSONCodec creates a new RPC server codec with support for JSON-RPC 2.0
func NewJSONCodec(rwc io.ReadWriteCloser, req *http.Request, nr namespace.NamespaceManager) ServerCodec {
	d := json.NewDecoder(rwc)
	d.UseNumber()
	return &jsonCodec{
		closed: make(chan interface{}),
		d: d,
		e: json.NewEncoder(rwc),
		rw: rwc,
		req: req,
		nr: nr,
	}
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

// CheckHttpHeaders will check http header.
func (c *jsonCodec) CheckHttpHeaders(namespace string) common.RPCError {
	ns := c.nr.GetNamespaceByName(namespace)
	if ns == nil {
		return &common.NamespaceNotFound{Name: namespace}
	}
	cm := ns.GetCAManager()
	if !cm.IsCheckTCert() {
		return nil
	}
	c.decMu.Lock()
	defer c.decMu.Unlock()
	//TODO fix cert problem
	//signature := c.req.Header.Get("signature")
	//msg := common.TransportDecode(c.req.Header.Get("msg"))
	//tcertPem := common.TransportDecode(c.req.Header.Get("tcert"))
	//tcert, err := primitives.ParseCertificate(tcertPem)
	//if err != nil {
	//	log.Error("fail to parse tcert.", err)
	//	return &common.UnauthorizedError{}
	//}
	//tcertPublicKey := tcert.PublicKey
	//pubKey := tcertPublicKey.(*(ecdsa.PublicKey))

	//signB := common.Hex2Bytes(signature)
	//verifySignature, err := primitives.ECDSAVerifyTransport(pubKey, []byte(msg), signB)
	//if err != nil || !verifySignature {
	//	log.Error("Fail to verify TransportSignture!", err)
	//	return &common.UnauthorizedError{}
	//}
	//verifyTcert, err := cm.VerifyTCert(tcertPem)
	//
	//if verifyTcert == false || err != nil {
	//	log.Error("Fail to verify tcert!", err)
	//	return &common.UnauthorizedError{}
	//}
	return nil
}

// ReadRequestHeaders will read new requests without parsing the arguments. It will
// return a collection of requests, an indication if these requests are in batch
// form or an error when the incoming message could not be read/parsed.
func (c *jsonCodec) ReadRequestHeaders() ([]*common.RPCRequest, bool, common.RPCError) {
	c.decMu.Lock()
	defer c.decMu.Unlock()

	var incomingMsg json.RawMessage
	if err := c.d.Decode(&incomingMsg); err != nil {
		log.Error(err)
		return nil, false, &common.InvalidRequestError{Message: err.Error()}
	}
	if isBatch(incomingMsg) {
		return parseBatchRequest(incomingMsg)
	}

	return parseRequest(incomingMsg)
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
func parseRequest(incomingMsg json.RawMessage) ([]*common.RPCRequest, bool, common.RPCError) {
	var in JSONRequest
	if err := json.Unmarshal(incomingMsg, &in); err != nil {
		return nil, false, &common.InvalidMessageError{Message:err.Error()}
	}
	if err := checkReqId(in.Id); err != nil {
		return nil, false, &common.InvalidMessageError{Message:err.Error()}
	}

	// regular RPC call
	elems := strings.Split(in.Method, serviceMethodSeparator)
	if len(elems) != 2 {
		return nil, false, &common.MethodNotFoundError{Service:in.Method, Method:""}
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
		return nil, false, &common.InvalidMessageError{Message:err.Error()}
	}

	requests := make([]*common.RPCRequest, len(in))
	for i, r := range in {
		if err := checkReqId(r.Id); err != nil {
			return nil, false, &common.InvalidMessageError{Message:err.Error()}
		}

		id := &in[i].Id

		elems := strings.Split(r.Method, serviceMethodSeparator)
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

// Write message to client
func (c *jsonCodec) Write(res interface{}) error {
	c.encMu.Lock()
	defer c.encMu.Unlock()

	return c.e.Encode(res)
}

// Close the underlying connection
func (c *jsonCodec) Close() {
	c.closer.Do(func() {
		close(c.closed)
		c.rw.Close()
	})
}

// Closed returns a channel which will be closed when Close is called
func (c *jsonCodec) Closed() <-chan interface{} {
	return c.closed
}
