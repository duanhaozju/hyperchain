package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"reflect"
	"strings"
	"context"
	"hyperchain/api"
)

const (
	UnsubscribeMethodSuffix  = "_unsubscribe"
)

type RequestProcessor interface {
	Start() error
	Stop() error
	ProcessRequest(request *common.RPCRequest) *common.RPCResponse
}

type JsonRpcProcessorImpl struct {
	namespace string
	apis      map[string]*api.API
	services  serviceRegistry // map hpc to methods of hpc
	log       *logging.Logger
}

//NewJsonRpcProcessorImpl new an instance of JsonRpcProcessorImpl with namespace and apis
func NewJsonRpcProcessorImpl(namespace string, apis map[string]*api.API) *JsonRpcProcessorImpl {
	jpri := &JsonRpcProcessorImpl{
		namespace: namespace,
		apis:      apis,
		services:  make(serviceRegistry),
		log:       common.GetLogger(namespace, "rpc"),
	}
	return jpri
}

func (jrpi *JsonRpcProcessorImpl) ProcessRequest(request *common.RPCRequest) *common.RPCResponse {
	sr := jrpi.checkRequestParams(request)
	return jrpi.exec(request.Ctx, sr)
}

//Start starts an instance of JsonRpcProcessorImpl.
func (jrpi *JsonRpcProcessorImpl) Start() error {
	err := jrpi.registerAllName()
	if err != nil {
		jrpi.log.Errorf("Failed to start RPC Manager of namespace %s!!!", jrpi.namespace)
		return err
	}
	return nil
}

func (jrpi *JsonRpcProcessorImpl) Stop() error {
	return nil
}

//RegisterAllName registers all namespace of given JsonRpcProcessorImpl.
func (jrpi *JsonRpcProcessorImpl) registerAllName() error {
	if jrpi.apis == nil || len(jrpi.apis) == 0 {
		return errors.New("no api methods registered")
	}
	for _, api := range jrpi.apis {
		if err := jrpi.registerName(api.Srvname, api.Service); err != nil {
			jrpi.log.Errorf("registerName error: %v ", err)
			return err
		}
	}

	return nil
}

// registerName will create an service for the given rcvr type under the given name. When no methods on the given rcvr
// match the criteria to be either a RPC method or a subscription, then an error is returned. Otherwise a new service is
// created and added to the service collection this server instance serves.
func (jrpi *JsonRpcProcessorImpl) registerName(name string, rcvr interface{}) error {
	svc := new(service)
	svc.typ = reflect.TypeOf(rcvr)
	rcvrVal := reflect.ValueOf(rcvr)

	if name == "" {
		return fmt.Errorf("no service name for type %s", svc.typ.String())
	}

	if !isExported(reflect.Indirect(rcvrVal).Type().Name()) {
		return fmt.Errorf("%s is not exported", reflect.Indirect(rcvrVal).Type().Name())
	}

	// already a previous service register under given sname, merge methods/subscriptions
	if regsvc, present := jrpi.services[name]; present {
		methods, subscriptions := suitableCallbacks(rcvrVal, svc.typ)
		if len(methods) == 0 && len(subscriptions) == 0 {
			return fmt.Errorf("Service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
		}

		for _, m := range methods {
			regsvc.callbacks[formatName(m.method.Name)] = m
		}
		for _, s := range subscriptions {
			regsvc.subscriptions[formatName(s.method.Name)] = s
		}

		return nil
	}

	svc.name = name
	svc.callbacks, svc.subscriptions = suitableCallbacks(rcvrVal, svc.typ)

	if len(svc.callbacks) == 0 && len(svc.subscriptions) == 0 {
		return fmt.Errorf("Service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
	}

	jrpi.services[svc.name] = svc

	return nil
}

// checkRequestParams requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (jrpi *JsonRpcProcessorImpl) checkRequestParams(req *common.RPCRequest) *serverRequest {
	var sr *serverRequest
	var ok bool
	var svc *service

	if req.IsPubSub && strings.HasSuffix(req.Method, UnsubscribeMethodSuffix) {
		sr = &serverRequest{id: req.Id, isUnsubscribe: true}
		argTypes := []reflect.Type{reflect.TypeOf("")} // expect subscription id as first arg
		if args, err := jrpi.ParseRequestArguments(argTypes, req.Params); err == nil {
			sr.args = args
		} else {
			sr.err = &common.InvalidParamsError{Message: err.Error()}
		}
		return sr
	}

	if svc, ok = jrpi.services[req.Service]; !ok { // rpc method isn't available
		sr = &serverRequest{id: req.Id, err: &common.MethodNotFoundError{Service: req.Service, Method: req.Method}}
		return sr
	}

	if req.IsPubSub { // eth_subscribe, r.method contains the subscription method name
		if callb, ok := svc.subscriptions[req.Method]; ok {
			sr = &serverRequest{id: req.Id, svcname: svc.name, callb: callb}
			if req.Params != nil && len(callb.argTypes) > 0 {
				argTypes := []reflect.Type{reflect.TypeOf("")}
				argTypes = append(argTypes, callb.argTypes...)
				if args, err := jrpi.ParseRequestArguments(argTypes, req.Params); err == nil {
					sr.args = args[1:] // first one is service.method name which isn't an actual argument
				} else {
					sr.err = &common.InvalidParamsError{Message: err.Error()}
				}
			}
		} else {
			 sr = &serverRequest{id: req.Id, err: &common.MethodNotFoundError{Service: req.Service, Method: req.Method}}
		}
		return sr
	}

	if callb, ok := svc.callbacks[req.Method]; ok { // lookup RPC method
		sr = &serverRequest{id: req.Id, svcname: svc.name, callb: callb}
		if req.Params != nil && len(callb.argTypes) > 0 {
			if args, err := jrpi.ParseRequestArguments(callb.argTypes, req.Params); err == nil {
				sr.args = args
			} else {
				sr.err = &common.InvalidParamsError{Message: err.Error()}
			}
		}
		return sr
	}
	return &serverRequest{
		id: req.Id,
		err: &common.MethodNotFoundError{
			Service: req.Service,
			Method:  req.Method},
	}

}

// ParseRequestArguments tries to parse the given params (json.RawMessage) with the given types. It returns the parsed
// values or an error when the parsing failed.
func (jrpi *JsonRpcProcessorImpl) ParseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, error) {
	if args, ok := params.(json.RawMessage); !ok {
		return nil, &common.InvalidParamsError{
			Message: "Invalid params supplied",
		}
	} else {
		return jrpi.parsePositionalArguments(args, argTypes)
	}
}

// parsePositionalArguments tries to parse the given args to an array of values with the given types.
// It returns the parsed values or an error when the args could not be parsed. Missing optional arguments
// are returned as reflect.Zero values.
func (jrpi *JsonRpcProcessorImpl) parsePositionalArguments(args json.RawMessage, callbackArgs []reflect.Type) ([]reflect.Value, error) {
	params := make([]interface{}, 0, len(callbackArgs))

	for _, t := range callbackArgs {
		params = append(params, reflect.New(t).Interface()) // Interface()转换为原来的类型
	}

	if err := json.Unmarshal(args, &params); err != nil {
		jrpi.log.Info(err)
		return nil, &common.InvalidParamsError{Message: err.Error()}
	}

	if len(params) > len(callbackArgs) {
		return nil, &common.InvalidParamsError{
			Message: fmt.Sprintf("too many params, want %d got %d", len(callbackArgs), len(params)),
		}
	}

	// assume missing params are null values
	for i := len(params); i < len(callbackArgs); i++ {
		params = append(params, nil)
	}

	argValues := make([]reflect.Value, len(params))
	for i, p := range params {
		// verify that JSON null values are only supplied for optional arguments (ptr types)
		if p == nil && callbackArgs[i].Kind() != reflect.Ptr {
			return nil, &common.InvalidParamsError{
				Message: fmt.Sprintf("invalid or missing value for params[%d]", i),
			}
		}
		if p == nil {
			argValues[i] = reflect.Zero(callbackArgs[i])
		} else {
			// deref pointers values creates previously with reflect.New
			argValues[i] = reflect.ValueOf(p).Elem()
		}
	}
	return argValues, nil
}

// exec executes the given request and writes the result back using the codec.
func (jrpi *JsonRpcProcessorImpl) exec(ctx context.Context, req *serverRequest) *common.RPCResponse {

	response, callback := jrpi.handle(ctx, req)

	// when request was a subscribe request this allows these subscriptions to be actived
	if callback != nil {
		callback()
	}
	return response
}

// handle executes a request and returns the response from the callback.
func (jrpi *JsonRpcProcessorImpl) handle(ctx context.Context, req *serverRequest) (*common.RPCResponse, func()) {
	if req.err != nil {
		return jrpi.CreateErrorResponse(&req.id, req.err), nil
	}

	//if req.isUnsubscribe { // cancel subscription, first param must be the subscription id
	//	if len(req.args) >= 1 && req.args[0].Kind() == reflect.String {
	//		notifier, supported := jsonrpc.NotifierFromContext(ctx)
	//		if !supported { // interface doesn't support subscriptions (e.g. http)
	//			return jrpi.CreateErrorResponse(&req.id, &common.CallbackError{Message: jsonrpc.ErrNotificationsUnsupported.Error()}), nil
	//		}
	//
	//		subid := common.ID(req.args[0].String())
	//		if err := notifier.Unsubscribe(subid); err != nil {
	//			return jrpi.CreateErrorResponse(&req.id, &common.CallbackError{Message: err.Error()}), nil
	//		}
	//
	//		return jrpi.CreateResponse(req.id, true), nil
	//	}
	//	return jrpi.CreateErrorResponse(&req.id, &common.InvalidParamsError{Message: "Expected subscription id as first argument"}), nil
	//}

	if req.callb.isSubscribe {
		subid, err := jrpi.createSubscription(ctx, req)
		if err != nil {
			return jrpi.CreateErrorResponse(&req.id, &common.CallbackError{Message: err.Error()}), nil
		}

		// active the subscription after the sub id was successfully sent to the client
		//activateSub := func() {
		//	notifier, _ := jsonrpc.NotifierFromContext(ctx)
		//	notifier.Activate(subid, req.svcname, jrpi.namespace)
		//}
		//
		//return jrpi.CreateResponse(req.id, subid), activateSub

		return jrpi.CreateResponse(req.id, subid), nil
		//return jrpi.CreateNotification(subid, req.svcname, jrpi.namespace, subid), nil
	}

	// regular RPC call, prepare arguments
	if len(req.args) != len(req.callb.argTypes) {
		err := &common.InvalidParamsError{
			Message: fmt.Sprintf("%s%s%s expects %d parameters, got %d",
				req.svcname, "_", req.callb.method.Name, len(req.callb.argTypes), len(req.args))}
		return jrpi.CreateErrorResponse(&req.id, err), nil
	}

	arguments := []reflect.Value{req.callb.rcvr}
	if req.callb.hasCtx {
		arguments = append(arguments, reflect.ValueOf(ctx))
	}
	if len(req.args) > 0 {
		arguments = append(arguments, req.args...)
	}

	// execute RPC method and return result
	reply := req.callb.method.Func.Call(arguments)
	if len(reply) == 0 {
		return jrpi.CreateResponse(req.id, nil), nil
	}

	if req.callb.errPos >= 0 { // test if method returned an error
		if !reply[req.callb.errPos].IsNil() {
			e := reply[req.callb.errPos].Interface().(common.RPCError)
			res := jrpi.CreateErrorResponse(&req.id, e)
			return res, nil
		}
	}
	return jrpi.CreateResponse(req.id, reply[0].Interface()), nil
}

func (jrpi *JsonRpcProcessorImpl) CreateResponse(id interface{}, reply interface{}) *common.RPCResponse {

	if _, ok := reply.(common.ID); ok {
		return &common.RPCResponse{
			Namespace: jrpi.namespace,
			Id:        id,
			Reply:     reply,
			Error:     nil,
			IsPubSub:  true,
		}
	}

	return &common.RPCResponse{
		Namespace: jrpi.namespace,
		Id:        id,
		Reply:     reply,
		Error:     nil,
		IsPubSub:  false,
	}
}

// CreateNotification will create a JSON-RPC notification with the given subscription id and event as params.
func (jrpi *JsonRpcProcessorImpl) CreateNotification(subid common.ID, service, namespace string, event interface{}) *common.RPCNotification {
	//if isHexNum(reflect.TypeOf(event)) {
	//	return &common.RPCNotification{Version: jsonrpcVersion, Method: namespace + common.NotificationMethodSuffix,
	//		Params: jsonSubscription{Subscription: subid, Result: fmt.Sprintf(`%#x`, event)}}
	//}
	//
	//return &jsonNotification{Version: jsonrpcVersion, Method: namespace + notificationMethodSuffix,
	//	Params: jsonSubscription{Subscription: subid, Result: event}}
	return &common.RPCNotification{
		Namespace: namespace,
		Service: service,
		SubId: subid,
		Result: event,

	}
}

func (jrpi *JsonRpcProcessorImpl) CreateErrorResponse(id interface{}, err common.RPCError) *common.RPCResponse {
	return &common.RPCResponse{
		Namespace: jrpi.namespace,
		Id:        id,
		Reply:     nil,
		Error:     err,
	}
}

func (jrpi *JsonRpcProcessorImpl) CreateErrorResponseWithInfo(id interface{}, err common.RPCError, info interface{}) *common.RPCResponse {
	return &common.RPCResponse{
		Namespace: jrpi.namespace,
		Id:        id,
		Reply:     info,
		Error:     err,
	}
}

// createSubscription will call the subscription callback and returns the subscription id or error.
func (jrpi *JsonRpcProcessorImpl) createSubscription(ctx context.Context, req *serverRequest) (common.ID, error) {
	// subscription have as first argument the context following optional arguments
	args := []reflect.Value{req.callb.rcvr, reflect.ValueOf(ctx)}
	args = append(args, req.args...)
	reply := req.callb.method.Func.Call(args)

	if !reply[1].IsNil() { // subscription creation failed
		return "", reply[1].Interface().(error)
	}

	//return reply[0].Interface().(*common.Subscription).ID, nil
	return reply[0].Interface().(common.ID), nil
}
