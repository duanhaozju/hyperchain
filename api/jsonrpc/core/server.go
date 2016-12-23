//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/fatih/set.v0"
	"github.com/op/go-logging"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("jsonrpc")
}
const (
	stopPendingRequestTimeout = 3 * time.Second // give pending requests stopPendingRequestTimeout the time to finish when the server is stopped

	MetadataApi     = "rpc"
	DefaultHTTPApis = "tx,node,block,account"
)

// CodecOption specifies which type of messages this codec supports
type CodecOption int

const (
	// OptionMethodInvocation is an indication that the codec supports RPC method calls
	OptionMethodInvocation CodecOption = 1 << iota
)

// NewServer will create a new server instance with no registered handlers.
func NewServer() *Server {
	server := &Server{
		services:      make(serviceRegistry),
		subscriptions: make(subscriptionRegistry),
		codecs:        set.New(),
		run:           1,
	}

	// register a default service which will provide meta information about the RPC service such as the services and
	// methods it offers.
	rpcService := &RPCService{server}
	server.RegisterName(MetadataApi, rpcService)

	return server
}

// RPCService gives meta information about the server.
// e.g. gives information about the loaded modules.
type RPCService struct {
	server *Server
}

// Modules returns the list of RPC services with their version number
func (s *RPCService) Modules() map[string]string {
	modules := make(map[string]string)
	for name := range s.server.services {
		modules[name] = "1.0"
	}
	return modules
}

// RegisterName will create an service for the given rcvr type under the given name. When no methods on the given rcvr
// match the criteria to be either a RPC method or a subscription an error is returned. Otherwise a new service is
// created and added to the service collection this server instance serves.
// 注册服务
func (s *Server) RegisterName(name string, rcvr interface{}) error {
	if s.services == nil {
		s.services = make(serviceRegistry)
	}

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
	// 如果给定name的服务已经注册
	if regsvc, present := s.services[name]; present {
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

	s.services[svc.name] = svc

	return nil
}

// hasOption returns true if option is included in options, otherwise false
func hasOption(option CodecOption, options []CodecOption) bool {
	for _, o := range options {
		if option == o {
			return true
		}
	}
	return false
}

// serveRequest will reads requests from the codec, calls the RPC callback and
// writes the response to the given codec.
//
// If singleShot is true it will process a single request, otherwise it will handle
// requests until the codec returns an error when reading a request (in most cases
// an EOF). It executes requests in parallel when singleShot is false.
func (s *Server) serveRequest(codec ServerCodec, singleShot bool, options CodecOption) error {
	//log.Info("============enter serveRequest()=================")
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Errorf(string(buf))
		}

		s.codecsMu.Lock()
		s.codecs.Remove(codec)
		s.codecsMu.Unlock()

		return
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.codecsMu.Lock()
	if atomic.LoadInt32(&s.run) != 1 { // server stopped
		s.codecsMu.Unlock()
		return &shutdownError{}
	}
	s.codecs.Add(codec)
	s.codecsMu.Unlock()

	// test if the server is ordered to stop
	for atomic.LoadInt32(&s.run) == 1 {
		reqs, batch, err := s.readRequest(codec)
		if err != nil {
			log.Debugf("%v\n", err)
			codec.Write(codec.CreateErrorResponse(nil, err))
			return nil
		}

		// check if server is ordered to shutdown and return an error
		// telling the client that his request failed.
		if atomic.LoadInt32(&s.run) != 1 {
			err = &shutdownError{}
			if batch {
				resps := make([]interface{}, len(reqs))
				for i, r := range reqs {
					resps[i] = codec.CreateErrorResponse(&r.id, err)
				}
				codec.Write(resps)
			} else {
				codec.Write(codec.CreateErrorResponse(&reqs[0].id, err))
			}
			return nil
		}

		if singleShot && batch {
			s.execBatch(ctx, codec, reqs)
			return nil
		} else if singleShot && !batch {
			s.exec(ctx, codec, reqs[0])
			return nil
		} else if !singleShot && batch {
			go s.execBatch(ctx, codec, reqs)
		} else {
			go s.exec(ctx, codec, reqs[0])
		}
	}

	return nil
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes the
// response back using the given codec. It will block until the codec is closed or the server is
// stopped. In either case the codec is closed.
func (s *Server) ServeCodec(codec ServerCodec, options CodecOption) {
	defer codec.Close()
	s.serveRequest(codec, false, options)
}

// ServeSingleRequest reads and processes a single RPC request from the given codec. It will not
// close the codec unless a non-recoverable error has occurred. Note, this method will return after
// a single request has been processed!
func (s *Server) ServeSingleRequest(codec ServerCodec, options CodecOption) {
	s.serveRequest(codec, true, options)
}

// Stop will stop reading new requests, wait for stopPendingRequestTimeout to allow pending requests to finish,
// close all codecs which will cancels pending requests/subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log.Debug("RPC Server shutdown initiatied")
		time.AfterFunc(stopPendingRequestTimeout, func() {
			s.codecsMu.Lock()
			defer s.codecsMu.Unlock()
			s.codecs.Each(func(c interface{}) bool {
				c.(ServerCodec).Close()
				return true
			})
		})
	}
}

// handle executes a request and returns the response from the callback.
func (s *Server) handle(ctx context.Context, codec ServerCodec, req *serverRequest) (interface{}, func()) {
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

// exec executes the given request and writes the result back using the codec.
func (s *Server) exec(ctx context.Context, codec ServerCodec, req *serverRequest) {
	//log.Info("=============enter exec()=================")
	var response interface{}
	var callback func()

	if req.err != nil {
		response = codec.CreateErrorResponse(&req.id, req.err)
	} else {
		response, callback = s.handle(ctx, codec, req)
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

// execBatch executes the given requests and writes the result back using the codec.
// It will only write the response back when the last request is processed.
func (s *Server) execBatch(ctx context.Context, codec ServerCodec, requests []*serverRequest) {
	responses := make([]interface{}, len(requests))
	var callbacks []func()
	for i, req := range requests {
		if req.err != nil {
			responses[i] = codec.CreateErrorResponse(&req.id, req.err)
		} else {
			var callback func()
			if responses[i], callback = s.handle(ctx, codec, req); callback != nil {
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

// readRequest requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (s *Server) readRequest(codec ServerCodec) ([]*serverRequest, bool, RPCError) {
	//log.Info("============enter readRequest()=================")
	reqs, batch, err := codec.ReadRequestHeaders()
	if err != nil {
		return nil, batch, err
	}
	requests := make([]*serverRequest, len(reqs))

	// verify requests
	for i, r := range reqs {
		var ok bool
		var svc *service

		if svc, ok = s.services[r.service]; !ok { // rpc method isn't available
			requests[i] = &serverRequest{id: r.id, err: &methodNotFoundError{r.service, r.method}}
			continue
		}

		if callb, ok := svc.callbacks[r.method]; ok { // lookup RPC method
			requests[i] = &serverRequest{id: r.id, svcname: svc.name, callb: callb}
			if r.params != nil && len(callb.argTypes) > 0 {
				if args, err := codec.ParseRequestArguments(callb.argTypes, r.params); err == nil {
					requests[i].args = args
				} else {
					requests[i].err = &invalidParamsError{err.Error()}
				}
			}
			continue
		}

		requests[i] = &serverRequest{id: r.id, err: &methodNotFoundError{r.service, r.method}}
	}
	return requests, batch, nil
}
