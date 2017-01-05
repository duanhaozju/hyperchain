//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"strconv"
)

const (
	JSONRPCVersion         = "2.0"
	serviceMethodSeparator = "_"
)

// JSON-RPC request
type JSONRequest struct {
	Method  string          `json:"method"`
	Version string          `json:"jsonrpc"`
	Id      json.RawMessage `json:"id,omitempty"`
	Payload json.RawMessage `json:"params,omitempty"`
}

// JSON-RPC response
type JSONSuccessResponse struct {
	Version string      `json:"jsonrpc"`
	Id      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result"`
}

// JSON-RPC error object
type JSONError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON-RPC error response
type JSONErrResponse struct {
	Version string      `json:"jsonrpc"`
	Id      interface{} `json:"id,omitempty"`
	Error   JSONError   `json:"error"`
}

// Hyperchain Release1.2
type JSONResponse struct {
	Version string 		`json:"jsonrpc"`
	Id 	interface{} 	`json:"id,omitempty"`
	Code 	int 		`json:"code"`
	Message string 		`json:"message"`
	Result 	interface{} 	`json:"result"`
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
}

// NewJSONCodec creates a new RPC server codec with support for JSON-RPC 2.0
func NewJSONCodec(rwc io.ReadWriteCloser) ServerCodec {
	d := json.NewDecoder(rwc)
	d.UseNumber()
	return &jsonCodec{closed: make(chan interface{}), d: d, e: json.NewEncoder(rwc), rw: rwc}
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

// ReadRequestHeaders will read new requests without parsing the arguments. It will
// return a collection of requests, an indication if these requests are in batch
// form or an error when the incoming message could not be read/parsed.
func (c *jsonCodec) ReadRequestHeaders() ([]rpcRequest, bool, RPCError) {
	c.decMu.Lock()
	defer c.decMu.Unlock()

	var incomingMsg json.RawMessage
	if err := c.d.Decode(&incomingMsg); err != nil {
		return nil, false, &invalidRequestError{err.Error()}
	}
	//log.Info(string(incomingMsg))
	if isBatch(incomingMsg) {
		return parseBatchRequest(incomingMsg)
	}

	return parseRequest(incomingMsg)
}

// checkReqId returns an error when the given reqId isn't valid for RPC method calls.
// valid id's are strings, numbers or null
func checkReqId(reqId json.RawMessage) error {
	if len(reqId) == 0 {
		return fmt.Errorf("missing request id")
	}
	if _, err := strconv.ParseFloat(string(reqId), 64); err == nil {
		return nil
	}
	var str string
	if err := json.Unmarshal(reqId, &str); err == nil {
		return nil
	}
	return fmt.Errorf("invalid request id")
}

// parseRequest will parse a single request from the given RawMessage. It will return
// the parsed request, an indication if the request was a batch or an error when
// the request could not be parsed.
func parseRequest(incomingMsg json.RawMessage) ([]rpcRequest, bool, RPCError) {
	var in JSONRequest
	if err := json.Unmarshal(incomingMsg, &in); err != nil {
		return nil, false, &invalidMessageError{err.Error()}
	}
	//log.Info(in)
	if err := checkReqId(in.Id); err != nil {
		return nil, false, &invalidMessageError{err.Error()}
	}

	// regular RPC call
	elems := strings.Split(in.Method, serviceMethodSeparator)
	if len(elems) != 2 {
		return nil, false, &methodNotFoundError{in.Method, ""}
	}

	if len(in.Payload) == 0 {
		return []rpcRequest{rpcRequest{service: elems[0], method: elems[1], id: &in.Id}}, false, nil
	}

	return []rpcRequest{rpcRequest{service: elems[0], method: elems[1], id: &in.Id, params: in.Payload}}, false, nil
}

// parseBatchRequest will parse a batch request into a collection of requests from the given RawMessage, an indication
// if the request was a batch or an error when the request could not be read.
func parseBatchRequest(incomingMsg json.RawMessage) ([]rpcRequest, bool, RPCError) {
	var in []JSONRequest
	if err := json.Unmarshal(incomingMsg, &in); err != nil {
		return nil, false, &invalidMessageError{err.Error()}
	}

	requests := make([]rpcRequest, len(in))
	for i, r := range in {
		if err := checkReqId(r.Id); err != nil {
			return nil, false, &invalidMessageError{err.Error()}
		}

		id := &in[i].Id

		elems := strings.Split(r.Method, serviceMethodSeparator)
		if len(elems) != 2 {
			return nil, true, &methodNotFoundError{r.Method, ""}
		}

		if len(r.Payload) == 0 {
			requests[i] = rpcRequest{service: elems[0], method: elems[1], id: id, params: nil}
		} else {
			requests[i] = rpcRequest{service: elems[0], method: elems[1], id: id, params: r.Payload}
		}
	}

	return requests, true, nil
}

// ParseRequestArguments tries to parse the given params (json.RawMessage) with the given types. It returns the parsed
// values or an error when the parsing failed.
func (c *jsonCodec) ParseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, RPCError) {
	//log.Info("==================enter ParseRequestArguments()==================")
	if args, ok := params.(json.RawMessage); !ok {
		return nil, &invalidParamsError{"Invalid params supplied"}
	} else {
		return parsePositionalArguments(args, argTypes)
	}
}

// parsePositionalArguments tries to parse the given args to an array of values with the given types.
// It returns the parsed values or an error when the args could not be parsed. Missing optional arguments
// are returned as reflect.Zero values.
func parsePositionalArguments(args json.RawMessage, callbackArgs []reflect.Type) ([]reflect.Value, RPCError) {
	//log.Info("===================enter parsePositionalArguments()====================")
	params := make([]interface{}, 0, len(callbackArgs))
	for _, t := range callbackArgs {
		params = append(params, reflect.New(t).Interface()) // Interface()转换为原来的类型
	}
	//log.Info(string(args)) // [{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":"0x9184e72a"}]
	//log.Info(params)	// [0xc8201437a0]
	if err := json.Unmarshal(args, &params); err != nil {
		log.Info(err)
		return nil, &invalidParamsError{err.Error()}
	}

	if len(params) > len(callbackArgs) {
		return nil, &invalidParamsError{fmt.Sprintf("too many params, want %d got %d", len(callbackArgs), len(params))}
	}

	// assume missing params are null values
	for i := len(params); i < len(callbackArgs); i++ {
		params = append(params, nil)
	}

	argValues := make([]reflect.Value, len(params))
	for i, p := range params {
		// verify that JSON null values are only supplied for optional arguments (ptr types)
		if p == nil && callbackArgs[i].Kind() != reflect.Ptr {
			return nil, &invalidParamsError{fmt.Sprintf("invalid or missing value for params[%d]", i)}
		}
		if p == nil {
			argValues[i] = reflect.Zero(callbackArgs[i])
		} else { // deref pointers values creates previously with reflect.New
			argValues[i] = reflect.ValueOf(p).Elem()
			//log.Infof("%#v",argValues[i])  // hpc.SendTxArgs{From:common.Address{0x0, 0xf, 0x1a, 0x7a, 0x8, 0xcc, 0xc4, 0x8e, 0x5d, 0x30, 0xf8, 0x8, 0x50, 0xcf, 0x1c, 0xf2, 0x83, 0xaa, 0x3a, 0xbd}, To:"0x0000000000000000000000000000000000000003", Gas:"", GasPrice:"", Value:"0x9184e72a", Payload:""}
		}
	}

	return argValues, nil
}

// CreateResponse will create a JSON-RPC success response with the given id and reply as result.
func (c *jsonCodec) CreateResponse(id interface{}, reply interface{}) interface{} {
	if isHexNum(reflect.TypeOf(reply)) {
		//return &JSONSuccessResponse{Version: JSONRPCVersion, Id: id, Result: fmt.Sprintf(`%#x`, reply)}
		return &JSONResponse{Version: JSONRPCVersion, Id: id, Code: 0, Message: "SUCCESS", Result: fmt.Sprintf(`%#x`, reply)}
	}
	//return &JSONSuccessResponse{Version: JSONRPCVersion, Id: id, Result: reply}
	return &JSONResponse{Version: JSONRPCVersion, Id: id, Code: 0, Message: "SUCCESS", Result: reply}
}

// CreateErrorResponse will create a JSON-RPC error response with the given id and error.
func (c *jsonCodec) CreateErrorResponse(id interface{}, err RPCError) interface{} {
	//return &JSONErrResponse{Version: JSONRPCVersion, Id: id, Error: JSONError{Code: err.Code(), Message: err.Error()}}
	return &JSONResponse{Version: JSONRPCVersion, Id: id, Code: err.Code(), Message: err.Error() }
}

// CreateErrorResponseWithInfo will create a JSON-RPC error response with the given id and error.
// info is optional and contains additional information about the error. When an empty string is passed it is ignored.
func (c *jsonCodec) CreateErrorResponseWithInfo(id interface{}, err RPCError, info interface{}) interface{} {
	//return &JSONErrResponse{Version: JSONRPCVersion, Id: id,
	//	Error: JSONError{Code: err.Code(), Message: err.Error(), Data: info}}
	return &JSONResponse{Version: JSONRPCVersion, Id: id,Code: err.Code(), Message: err.Error(), Result: info}
}

// CreateNotification will create a JSON-RPC notification with the given subscription id and event as params.
//func (c *jsonCodec) CreateNotification(subid string, event interface{}) interface{} {
//	if isHexNum(reflect.TypeOf(event)) {
//		return &jsonNotification{Version: JSONRPCVersion, Method: notificationMethod,
//			Params: jsonSubscription{Subscription: subid, Result: fmt.Sprintf(`%#x`, event)}}
//	}
//
//	return &jsonNotification{Version: JSONRPCVersion, Method: notificationMethod,
//		Params: jsonSubscription{Subscription: subid, Result: event}}
//}

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