package rpc

import (
	"reflect"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"encoding/json"
	"golang.org/x/net/context"
)

var log *logging.Logger // package-level logger

func init() {
	log = logging.MustGetLogger("rpc")
}

type RPCProcesser interface {
	Start()
	RegisterAllName() RPCError
	registerName(name string, rcvr interface{}) RPCError
	DelRegisterName()
	AddRegisterName()
	ProcessRequest(RPCRequest []common.RPCRequest, singleShot bool, batch bool) common.RPCResponse
	readRequset(reqs []common.RPCRequest) RPCError
	parsePositionalArguments(args json.RawMessage, callbackArgs []reflect.Type) ([]reflect.Value, RPCError)
}

type RPCProcesserImpl struct {
	namespace string
	apis      []API
	services  serviceRegistry	// map hpc to methods of hpc
}

//NewRPCManagerImpl new an instance of RPCManager with namespace and apis
func NewRPCManagerImpl(namespace string, apis []API) *RPCProcesserImpl {
	rpcproc := &RPCProcesserImpl{
		namespace: namespace,
		apis:      apis,
		services:  make(serviceRegistry),
	}

	return rpcproc
}

//Start starts an instance of RPCManager
func (rpcproc *RPCProcesserImpl) Start() {
	err := rpcproc.RegisterAllName()
	if err != nil {
		log.Errorf("Failed to start RPC Manager of namespace %s!!!", rpcproc.namespace)
		return nil
	}
}

//RegisterAllName registers all namespace of given RPCManager
func (rpcproc *RPCProcesserImpl) RegisterAllName() RPCError {
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
func (rpcproc *RPCProcesserImpl) registerName(name string, rcvr interface{}) RPCError {
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

// todo 相当于 handle()
func (rpcproc *RPCProcesserImpl) ProcessRequest(RPCRequest []common.RPCRequest, singleShot bool, batch bool) common.RPCResponse{
	reqs, err := rpcproc.readRequset(RPCRequest)

	// check if server is ordered to shutdown and return an error
	// telling the client that his request failed.

	//TODO how to arrange services.run
	//if atomic.LoadInt32(&rpcproc.services.run) != 1 {
	//	err := &shutdownError{}
	//	if batch {
	//		resps := make([]interface{}, len(reqs))
	//		for i, r := range reqs {
	//			resps[i] = codec.CreateErrorResponse(&r.id, err)
	//		}
	//		codec.Write(resps)
	//	} else {
	//		codec.Write(codec.CreateErrorResponse(&reqs[0].id, err))
	//	}
	//	return nil
	//}

	if singleShot && batch {
		rpcproc.execBatch(ctx, codec, reqs)
		return nil
	} else if singleShot && !batch {
		rpcproc.exec(ctx, codec, reqs[0])
		return nil
	} else if !singleShot && batch {
		go rpcproc.execBatch(ctx, codec, reqs)
	} else {
		go rpcproc.exec(ctx, codec, reqs[0])
	}
	return nil
}

// readRequest requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (rpcproc *RPCProcesserImpl) readRequset(reqs []common.RPCRequest) ([]*serverRequest, RPCError) {

	requests := make([]*serverRequest, len(reqs))

	// verify requests
	for i, r := range reqs {
		var ok bool
		var svc *service

		if svc, ok = rpcproc.services[r.Service]; !ok { // rpc method isn't available
			requests[i] = &serverRequest{id: r.Id, err: &methodNotFoundError{r.Service, r.Method}}
			continue
		}

		if callb, ok := svc.callbacks[r.Method]; ok { // lookup RPC method
			requests[i] = &serverRequest{id: r.Id, svcname: svc.name, callb: callb}
			if r.Params != nil && len(callb.argTypes) > 0 {
				if args, err := rpcproc.ParseRequestArguments(callb.argTypes, r.Params); err == nil {
					requests[i].args = args
				} else {
					requests[i].err = &invalidParamsError{err.Error()}
				}
			}
			continue
		}
		requests[i] = &serverRequest{id: r.Id, err: &methodNotFoundError{r.Service, r.Method}}
	}
	return requests, nil
}

// ParseRequestArguments tries to parse the given params (json.RawMessage) with the given types. It returns the parsed
// values or an error when the parsing failed.
func (rpcproc *RPCProcesserImpl) ParseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, RPCError) {
	//log.Info("==================enter ParseRequestArguments()==================")
	if args, ok := params.(json.RawMessage); !ok {
		return nil, &invalidParamsError{"Invalid params supplied"}
	} else {
		return rpcproc.parsePositionalArguments(args, argTypes)
	}
}

// parsePositionalArguments tries to parse the given args to an array of values with the given types.
// It returns the parsed values or an error when the args could not be parsed. Missing optional arguments
// are returned as reflect.Zero values.
func (rpcproc *RPCProcesserImpl) parsePositionalArguments(args json.RawMessage, callbackArgs []reflect.Type) ([]reflect.Value, RPCError) {
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

// exec executes the given request and writes the result back using the codec.
func (rpcproc *RPCProcesserImpl) exec(ctx context.Context, codec ServerCodec, req *serverRequest) {
	//log.Info("=============enter exec()=================")
	var response interface{}
	var callback func()

	if req.err != nil {
		response = codec.CreateErrorResponse(&req.id, req.err)
	} else {
		response, callback = rpcproc.handle(ctx, codec, req)
	}

	if err := codec.Write(response); err != nil {
		log.Errorf("%v\n", err)
		codec.Close()
	}

	// when request was a subscribe request this allows these subscriptions to be actived
	if callback != nil {
		callback()
	}
}

func (rpcproc *RPCProcesserImpl) execBatch(ctx context.Context, codec ServerCodec, requests []*serverRequest) {
	responses := make([]interface{}, len(requests))
	var callbacks []func()
	for i, req := range requests {
		fmt.Println("got a request", req.svcname)
		if req.err != nil {
			responses[i] = codec.CreateErrorResponse(&req.id, req.err)
		} else {
			var callback func()
			if responses[i], callback = rpcproc.handle(ctx, codec, req); callback != nil {
				callbacks = append(callbacks, callback)
			}
		}
	}

	if err := codec.Write(responses); err != nil {
		log.Errorf("%v\n", err)
		codec.Close()
	}

	// when request holds one of more subscribe requests this allows these subscriptions to be actived
	for _, c := range callbacks {
		c()
	}
}

// handle executes a request and returns the response from the callback.
func (rpcproc *RPCProcesserImpl) handle(ctx context.Context, codec ServerCodec, req *serverRequest) (interface{}, func()) {
	//log.Info("=========================enter handle()============================")
	if req.err != nil {
		return codec.CreateErrorResponse(&req.id, req.err), nil
	}

	// regular RPC call, prepare arguments
	if len(req.args) != len(req.callb.argTypes) {
		rpcErr := &invalidParamsError{fmt.Sprintf("%s%s%s expects %d parameters, got %d",
			req.svcname, serviceMethodSeparator, req.callb.method.Name,
			len(req.callb.argTypes), len(req.args))}
		return codec.CreateErrorResponse(&req.id, rpcErr), nil
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
		return codec.CreateResponse(req.id, nil), nil
	}

	if req.callb.errPos >= 0 { // test if method returned an error
		if !reply[req.callb.errPos].IsNil() {
			//e := reply[req.callb.errPos].Interface().(error)
			//res := codec.CreateErrorResponse(&req.id, &callbackError{e.Error()})
			e := reply[req.callb.errPos].Interface().(RPCError)
			res := codec.CreateErrorResponse(&req.id, e)
			return res, nil
		}
	}
	return codec.CreateResponse(req.id, reply[0].Interface()), nil
}