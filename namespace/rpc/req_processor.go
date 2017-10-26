package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/api"
	"hyperchain/common"
	"reflect"
	"strings"
)

type RequestProcessor interface {
	// Start registers all the JSON-RPC API service.
	Start() error

	// Stop stops process request.(Need to be implemented later.)
	Stop() error

	// ProcessRequest checks request parameters and then executes the given request.
	ProcessRequest(request *common.RPCRequest) *common.RPCResponse
}

type JsonRpcProcessorImpl struct {
	// namespace this JsonRpcProcessorImpl will serve.
	namespace string

	// apis this namespace provides.
	apis      map[string]*api.API

	// register the apis of this namespace to this processor.
	services  serviceRegistry

	// logger implementation
	logger    *logging.Logger
}

// NewJsonRpcProcessorImpl creates a new JsonRpcProcessorImpl instance for given namespace and apis.
func NewJsonRpcProcessorImpl(namespace string, apis map[string]*api.API) *JsonRpcProcessorImpl {
	jpri := &JsonRpcProcessorImpl{
		namespace: namespace,
		apis:      apis,
		services:  make(serviceRegistry),
		logger:    common.GetLogger(namespace, "rpc"),
	}
	for key,value:= range jpri.apis{
		jpri.logger.Critical("key, value:", key, reflect.TypeOf(value.Service))
	}
	return jpri
}

// Start registers all the JSON-RPC API service.
func (jrpi *JsonRpcProcessorImpl) Start() error {
	err := jrpi.registerAllAPIService()
	if err != nil {
		jrpi.logger.Errorf("Failed to start JSON-RPC processor of namespace %s .", jrpi.namespace)
		return err
	}
	return nil
}

// Stop stops the json rpc processor.
func (jrpi *JsonRpcProcessorImpl) Stop() error {
	return nil
}

// ProcessRequest checks request parameters and then executes the given request.
func (jrpi *JsonRpcProcessorImpl) ProcessRequest(request *common.RPCRequest) *common.RPCResponse {
	sr := jrpi.checkRequestParams(request)
	return jrpi.exec(request.Ctx, sr)
}

// registerAllAPIService will register all the JSON-RPC API service. If there are
// no services offered, an error is returned.
func (jrpi *JsonRpcProcessorImpl) registerAllAPIService() error {
	if jrpi.apis == nil || len(jrpi.apis) == 0 {
		return ErrNoApis
	}
	for _, api := range jrpi.apis {
		if err := jrpi.registerAPIService(api.Svcname, api.Service); err != nil {
			jrpi.logger.Errorf("registerAPIService error: %v ", err)
			return err
		}
	}

	return nil
}

// registerAPIService will create a service for the given rcvr type under the given svcname.
// When no methods on the given rcvr match the criteria to be either a RPC method or a
// subscription, then an error is returned. Otherwise a new service is created and added to
// the service collection this server instance serves.
func (jrpi *JsonRpcProcessorImpl) registerAPIService(svcname string, rcvr interface{}) error {
	svc := new(service)
	svc.typ = reflect.TypeOf(rcvr)
	rcvrVal := reflect.ValueOf(rcvr)

	if svcname == "" {
		jrpi.logger.Errorf("no service name for type %s", svc.typ.String())
		return ErrNoServiceName
	}

	if !isExported(reflect.Indirect(rcvrVal).Type().Name()) {
		jrpi.logger.Errorf("%s is not exported", reflect.Indirect(rcvrVal).Type().Name())
		return ErrNotExported
	}

	svc.name = svcname
	callbacks, subscriptions := suitableCallbacks(rcvrVal, svc.typ)
	if len(callbacks) == 0 && len(subscriptions) == 0 {
		jrpi.logger.Errorf("service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
		return ErrNoSuitable
	}

	// If there already existed a previous service registered under the given service name,
	// merge the methods/subscriptions, else use the service newed before.
	if regsvc, present := jrpi.services[svcname]; present {
		jrpi.logger.Debugf("already exist service with the same name: %s, merge it.", svcname)
		for _, m := range callbacks {
			regsvc.callbacks[formatName(m.method.Name)] = m
		}
		for _, s := range subscriptions {
			regsvc.subscriptions[formatName(s.method.Name)] = s
		}
		return nil
	}

	svc.callbacks, svc.subscriptions = callbacks, subscriptions
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

	if req.IsPubSub && strings.HasSuffix(req.Method, common.UnsubscribeMethodSuffix) {
		sr = &serverRequest{id: req.Id, isUnsubscribe: true}
		// expect subscription id as first arg
		argTypes := []reflect.Type{reflect.TypeOf("")}
		if args, err := jrpi.parseRequestArguments(argTypes, req.Params); err == nil {
			sr.args = args
		} else {
			sr.err = &common.InvalidParamsError{Message: err.Error()}
		}
		return sr
	}

	// If the given rpc method isn't available, return error.
	if svc, ok = jrpi.services[req.Service]; !ok {
		jrpi.logger.Debugf("No service named %s was found.", req.Service)
		sr = &serverRequest{id: req.Id, err: &common.MethodNotFoundError{Service: req.Service, Method: req.Method}}
		return sr
	}

	// For sub_subscribe, req.method contains the subscription method name.
	if req.IsPubSub {
		if callb, ok := svc.subscriptions[req.Method]; ok {
			sr = &serverRequest{id: req.Id, svcname: svc.name, callb: callb}
			if req.Params != nil && len(callb.argTypes) > 0 {
				argTypes := []reflect.Type{reflect.TypeOf("")}
				argTypes = append(argTypes, callb.argTypes...)
				if args, err := jrpi.parseRequestArguments(argTypes, req.Params); err == nil {
					// first arg is service.method name which isn't an actual argument
					sr.args = args[1:]
				} else {
					sr.err = &common.InvalidParamsError{Message: err.Error()}
				}
			}
		} else {
			jrpi.logger.Debugf("No subscription method %s_%s was found.", req.Service, req.Method)
			sr = &serverRequest{id: req.Id, err: &common.MethodNotFoundError{Service: req.Service, Method: req.Method}}
		}
		return sr
	}

	// For callbacks, req.method contains the callback method name, lookup RPC method.
	if callb, ok := svc.callbacks[req.Method]; ok {
		sr = &serverRequest{id: req.Id, svcname: svc.name, callb: callb}
		if req.Params != nil && len(callb.argTypes) > 0 {
			if args, err := jrpi.parseRequestArguments(callb.argTypes, req.Params); err == nil {
				sr.args = args
			} else {
				sr.err = &common.InvalidParamsError{Message: err.Error()}
			}
		}
		return sr
	}

	jrpi.logger.Debugf("No method named %s_%s was found.", req.Service, req.Method)
	return &serverRequest{
		id: req.Id,
		err: &common.MethodNotFoundError{
			Service: req.Service,
			Method:  req.Method},
	}

}

// parseRequestArguments tries to parse the given params (json.RawMessage) with the given
// types. It returns the parsed values or an error when the parsing failed.
func (jrpi *JsonRpcProcessorImpl) parseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, error) {
	if args, ok := params.(json.RawMessage); !ok {
		return nil, &common.InvalidParamsError{
			Message: "Invalid params supplied",
		}
	} else {
		return jrpi.parsePositionalArguments(args, argTypes)
	}
}

// parsePositionalArguments tries to parse the given args to an array of values with the
// given types. It returns the parsed values or an error when the args could not be parsed.
// Missing optional arguments are returned as reflect.Zero values.
func (jrpi *JsonRpcProcessorImpl) parsePositionalArguments(args json.RawMessage, callbackArgs []reflect.Type) ([]reflect.Value, error) {
	params := make([]interface{}, 0, len(callbackArgs))

	for _, t := range callbackArgs {
		params = append(params, reflect.New(t).Interface())
	}

	if err := json.Unmarshal(args, &params); err != nil {
		jrpi.logger.Info(err)
		return nil, &common.InvalidParamsError{Message: err.Error()}
	}

	if len(params) > len(callbackArgs) {
		errMsg := fmt.Sprintf("too many params, want %d got %d", len(callbackArgs), len(params))
		jrpi.logger.Info(errMsg)
		return nil, &common.InvalidParamsError{Message: errMsg}
	}

	// assume missing params are null values
	for i := len(params); i < len(callbackArgs); i++ {
		params = append(params, nil)
	}

	argValues := make([]reflect.Value, len(params))
	for i, p := range params {
		// verify that JSON null values are only supplied for optional arguments (ptr types)
		if p == nil && callbackArgs[i].Kind() != reflect.Ptr {
			errMsg := fmt.Sprintf("invalid or missing value for params[%d], %v---%v", i, p, callbackArgs[i].Kind())
			jrpi.logger.Info(errMsg)
			return nil, &common.InvalidParamsError{Message: errMsg}
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

// exec executes the given request and returns the response to upper layer.
func (jrpi *JsonRpcProcessorImpl) exec(ctx context.Context, req *serverRequest) *common.RPCResponse {

	response, callback := jrpi.handle(ctx, req)

	// when request was a subscribe request this allows these subscriptions to be active.
	if callback != nil {
		callback()
	}
	return response
}

// handle executes a request and returns the response from the method callback.
func (jrpi *JsonRpcProcessorImpl) handle(ctx context.Context, req *serverRequest) (*common.RPCResponse, func()) {
	if req.err != nil {
		return jrpi.CreateErrorResponse(&req.id, req.err), nil
	}

	// cancel subscription, first param must be the subscription id
	if req.isUnsubscribe {
		if len(req.args) >= 1 && req.args[0].Kind() == reflect.String {
			cres := jrpi.CreateResponse(req.id, common.ID(req.args[0].String()))
			cres.IsUnsub = true

			return cres, nil
		}
		return jrpi.CreateErrorResponse(&req.id, &common.InvalidParamsError{Message:
			"Expected subscription id as first argument"}), nil
	}

	if req.callb.isSubscribe {
		subid, err := jrpi.createSubscription(ctx, req)
		if err != nil {
			jrpi.logger.Error(err)
			return jrpi.CreateErrorResponse(&req.id, &common.CallbackError{Message: err.Error()}), nil
		}

		return jrpi.CreateResponse(req.id, subid), nil
	}

	// regular RPC call, prepare arguments
	if len(req.args) != len(req.callb.argTypes) {
		errMsg := fmt.Sprintf("%s%s%s expects %d parameters, got %d",
			req.svcname, "_", req.callb.method.Name, len(req.callb.argTypes), len(req.args))
		jrpi.logger.Error(errMsg)
		err := &common.InvalidParamsError{Message: errMsg}
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

	// test if method returned an error
	if req.callb.errPos >= 0 {
		if !reply[req.callb.errPos].IsNil() {
			e := reply[req.callb.errPos].Interface().(common.RPCError)
			if !isEmpty(reply[0]) {
				return jrpi.CreateErrorResponseWithInfo(&req.id, e, reply[0].Interface()), nil
			}
			return jrpi.CreateErrorResponse(&req.id, e), nil
		}
	}
	return jrpi.CreateResponse(req.id, reply[0].Interface()), nil
}

// CreateResponse will create regular RPCResponse.
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

// CreateErrorResponse will create an error RPCResponse.
func (jrpi *JsonRpcProcessorImpl) CreateErrorResponse(id interface{}, err common.RPCError) *common.RPCResponse {
	return &common.RPCResponse{
		Namespace: jrpi.namespace,
		Id:        id,
		Reply:     nil,
		Error:     err,
	}
}

//CreateErrorResponseWithInfo will create an error RPCResponse with given info.
func (jrpi *JsonRpcProcessorImpl) CreateErrorResponseWithInfo(id interface{}, err common.RPCError, info interface{}) *common.RPCResponse {
	return &common.RPCResponse{
		Namespace: jrpi.namespace,
		Id:        id,
		Reply:     info,
		Error:     err,
	}
}

// CreateNotification will create a JSON-RPC notification with the given subscription id and event as params.
func (jrpi *JsonRpcProcessorImpl) CreateNotification(subid common.ID, service, namespace string, event interface{}) *common.RPCNotification {
	return &common.RPCNotification{
		Namespace: namespace,
		Service:   service,
		SubId:     subid,
		Result:    event,
	}
}

// createSubscription will call the subscription callback and returns the subscription id or error.
func (jrpi *JsonRpcProcessorImpl) createSubscription(ctx context.Context, req *serverRequest) (common.ID, error) {
	// subscription have as first argument the context following optional arguments
	args := []reflect.Value{req.callb.rcvr, reflect.ValueOf(ctx)}
	args = append(args, req.args...)
	reply := req.callb.method.Func.Call(args)

	if len(reply) < 2 {
		return "", ErrInvalidReply
	}
	if !reply[1].IsNil() { // subscription creation failed
		return "", reply[1].Interface().(error)
	}

	return reply[0].Interface().(common.ID), nil
}
