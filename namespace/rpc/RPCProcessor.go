package rpc

import (
	"reflect"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"encoding/json"
	"golang.org/x/net/context"
	"hyperchain/api"
	"hyperchain/core/db_utils"
	"math/big"
)

var log *logging.Logger // package-level logger

func init() {
	log = logging.MustGetLogger("rpc")
}

type RequestProcesser interface {
	Start() error
	ProcessRequest(request *common.RPCRequest) *common.RPCResponse
}

type JsonRpcProcesserImpl struct {
	namespace string
	apis      []hpc.API
	services  serviceRegistry	// map hpc to methods of hpc
}

//NewRPCProcessorImpl new an instance of RPCManager with namespace and apis
func NewRPCProcessorImpl(namespace string, apis []hpc.API) *JsonRpcProcesserImpl {
	rpcproc := &JsonRpcProcesserImpl{
		namespace: namespace,
		apis:      apis,
		services:  make(serviceRegistry),
	}

	return rpcproc
}

//Start starts an instance of RPCManager
func (rpcproc *JsonRpcProcesserImpl) Start() error {
	err := rpcproc.registerAllName()
	if err != nil {
		log.Errorf("Failed to start RPC Manager of namespace %s!!!", rpcproc.namespace)
		return err
	}
	return nil
}

//RegisterAllName registers all namespace of given RPCManager
func (rpcproc *JsonRpcProcesserImpl) registerAllName() error {
	for _, api := range rpcproc.apis {
		if err := rpcproc.registerName(api.Srvname, api.Service); err != nil {
			log.Errorf("registerName error: %v ", err)
			return err
		}
	}

	return nil
}

// registerName will create an service for the given rcvr type under the given name. When no methods on the given rcvr
// match the criteria to be either a RPC method or a subscription, then an error is returned. Otherwise a new service is
// created and added to the service collection this server instance serves.
func (rpcproc *JsonRpcProcesserImpl) registerName(name string, rcvr interface{}) error {
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
	if regsvc, present := rpcproc.services[name]; present {
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
	svc.callbacks, svc.subscriptions = suitableCallbacks(rcvrVal, svc.typ) // 这里的callbacks就是api.go中的方法集合

	if len(svc.callbacks) == 0 && len(svc.subscriptions) == 0 {
		return fmt.Errorf("Service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
	}

	rpcproc.services[svc.name] = svc

	return nil
}

func (rpcproc *JsonRpcProcesserImpl) ProcessRequest(request *common.RPCRequest) *common.RPCResponse {
	sr := rpcproc.checkRequestParams(request)
	return rpcproc.exec(request.Ctx, sr)
}

// checkRequestParams requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (rpcproc *JsonRpcProcesserImpl) checkRequestParams(req *common.RPCRequest) (*serverRequest) {
		var sr *serverRequest
		var ok bool
		var svc *service

		if svc, ok = rpcproc.services[req.Service]; !ok { // rpc method isn't available
			sr = &serverRequest{id: req.Id, err: &common.MethodNotFoundError{req.Service, req.Method}}
			return sr
		}

		if callb, ok := svc.callbacks[req.Method]; ok { // lookup RPC method
			sr = &serverRequest{id: req.Id, svcname: svc.name, callb: callb}
			if req.Params != nil && len(callb.argTypes) > 0 {
				if args, err := rpcproc.ParseRequestArguments(callb.argTypes, req.Params); err == nil {
					sr.args = args
				} else {
					sr.err = &common.InvalidParamsError{err.Error()}
				}
			}
			return sr
		}
		return  &serverRequest{id: req.Id, err: &common.MethodNotFoundError{req.Service, req.Method}}

}

// ParseRequestArguments tries to parse the given params (json.RawMessage) with the given types. It returns the parsed
// values or an error when the parsing failed.
func (rpcproc *JsonRpcProcesserImpl) ParseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, error) {
	//log.Info("==================enter ParseRequestArguments()==================")
	if args, ok := params.(json.RawMessage); !ok {
		return nil, &common.InvalidParamsError{"Invalid params supplied"}
	} else {
		return rpcproc.parsePositionalArguments(args, argTypes)
	}
}

// parsePositionalArguments tries to parse the given args to an array of values with the given types.
// It returns the parsed values or an error when the args could not be parsed. Missing optional arguments
// are returned as reflect.Zero values.
func (rpcproc *JsonRpcProcesserImpl) parsePositionalArguments(args json.RawMessage, callbackArgs []reflect.Type) ([]reflect.Value, error) {
	//log.Info("===================enter parsePositionalArguments()====================")

	msg, msgLen, err := splitRawMessage(args)
	if err != nil {
		return nil, err
	}

	if msgLen < len(callbackArgs) {
		return nil, &common.InvalidParamsError{fmt.Sprintf("missing value for params")}
	} else if msgLen > len(callbackArgs) {
		return nil, &common.InvalidParamsError{fmt.Sprintf("too many params, want %d got %d", len(callbackArgs), msgLen)}
	}

	params := make([]interface{}, 0, len(callbackArgs))

	for i, t := range callbackArgs {
		if (t.Name() == "BlockNumber" || t.Name() == "*BlockNumber") {
			if chain := db_utils.GetChainCopy(rpcproc.namespace); chain != nil {
				height := chain.Height
				in := new(big.Int)

				if height == 0 {
					return nil, &common.InvalidParamsError{fmt.Sprintf("there is no block generated")}
				}

				if msg[i] == "\"latest\""{
					heightInt := in.SetUint64(height)
					msg[i] = heightInt.String()
				} else if blkNum, ok := in.SetString(msg[i], 0); ok && blkNum.Uint64() > height {
					return nil, &common.InvalidParamsError{
						fmt.Sprintf("block number is out of range, and now latest block number is %d", height)}
				}
			}
		}
		params = append(params, reflect.New(t).Interface()) // Interface()转换为原来的类型
	}

	args = joinRawMessage(msg)
	//log.Info(string(args)) // [{"from":"0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd","to":"0x0000000000000000000000000000000000000003","value":"0x9184e72a"}]
	//log.Info(params)	// [0xc8201437a0]
	if err := json.Unmarshal(args, &params); err != nil {
		log.Info(err)
		return nil, &common.InvalidParamsError{err.Error()}
	}

	if len(params) > len(callbackArgs) {
		return nil, &common.InvalidParamsError{fmt.Sprintf("too many params, want %d got %d", len(callbackArgs), len(params))}
	}

	// assume missing params are null values
	for i := len(params); i < len(callbackArgs); i++ {
		params = append(params, nil)
	}

	argValues := make([]reflect.Value, len(params))
	for i, p := range params {
		// verify that JSON null values are only supplied for optional arguments (ptr types)
		if p == nil && callbackArgs[i].Kind() != reflect.Ptr {
			return nil, &common.InvalidParamsError{fmt.Sprintf("invalid or missing value for params[%d]", i)}
		}
		if p == nil {
			argValues[i] = reflect.Zero(callbackArgs[i])
		} else { // deref pointers values creates previously with reflect.New
			argValues[i] = reflect.ValueOf(p).Elem()

		}
	}

	return argValues, nil
}

// exec executes the given request and writes the result back using the codec.
func (rpcproc *JsonRpcProcesserImpl) exec(ctx context.Context, req *serverRequest) *common.RPCResponse {

	response, callback := rpcproc.handle(ctx, req)

	// when request was a subscribe request this allows these subscriptions to be actived
	if callback != nil {
		callback()
	}

	return response
}

//func (rpcproc *JsonRpcProcesserImpl) execBatch(ctx context.Context, codec ServerCodec, requests []*serverRequest) {
//	responses := make([]interface{}, len(requests))
//	var callbacks []func()
//	for i, req := range requests {
//		fmt.Println("got a request", req.svcname)
//		if req.err != nil {
//			responses[i] = codec.CreateErrorResponse(&req.id, req.err)
//		} else {
//			var callback func()
//			if responses[i], callback = rpcproc.handle(ctx, codec, req); callback != nil {
//				callbacks = append(callbacks, callback)
//			}
//		}
//	}
//
//	if err := codec.Write(responses); err != nil {
//		log.Errorf("%v\n", err)
//		codec.Close()
//	}
//
//	// when request holds one of more subscribe requests this allows these subscriptions to be actived
//	for _, c := range callbacks {
//		c()
//	}
//}

// handle executes a request and returns the response from the callback.
func (rpcproc *JsonRpcProcesserImpl) handle(ctx context.Context, req *serverRequest) (*common.RPCResponse, func()) {
	if req.err != nil {
		return rpcproc.CreateErrorResponse(&req.id, req.err), nil
	}

	// regular RPC call, prepare arguments
	if len(req.args) != len(req.callb.argTypes) {
		rpcErr := &common.InvalidParamsError{
			fmt.Sprintf("%s%s%s expects %d parameters, got %d", req.svcname, "_", req.callb.method.Name,
			len(req.callb.argTypes), len(req.args))}
		return rpcproc.CreateErrorResponse(&req.id, rpcErr), nil
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
		return rpcproc.CreateResponse(req.id, nil), nil
	}

	if req.callb.errPos >= 0 { // test if method returned an error
		if !reply[req.callb.errPos].IsNil() {
			//e := reply[req.callb.errPos].Interface().(error)
			//res := codec.CreateErrorResponse(&req.id, &callbackError{e.Error()})
			e := reply[req.callb.errPos].Interface().(common.RPCError)
			res := rpcproc.CreateErrorResponse(&req.id, e)
			return res, nil
		}
	}
	return rpcproc.CreateResponse(req.id, reply[0].Interface()), nil
}

func (rpcproc *JsonRpcProcesserImpl) CreateResponse(id interface{}, reply interface{}) *common.RPCResponse {
	return &common.RPCResponse{
		Namespace: rpcproc.namespace,
		Id:    id,
		Reply: reply,
		Error: nil,
	}
}

func (rpcproc *JsonRpcProcesserImpl) CreateErrorResponse(id interface{}, err common.RPCError) *common.RPCResponse {
	return &common.RPCResponse{
		Namespace: rpcproc.namespace,
		Id:    id,
		Reply: nil,
		Error: err,
	}
}

func (rpcproc *JsonRpcProcesserImpl) CreateErrorResponseWithInfo(id interface{}, err common.RPCError, info interface{}) *common.RPCResponse {
	return &common.RPCResponse{
		Namespace: rpcproc.namespace,
		Id:    id,
		Reply: info,
		Error: err,
	}
}