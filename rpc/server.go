//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"gopkg.in/fatih/set.v0"
	admin "hyperchain/api/admin"
	"hyperchain/common"
	"hyperchain/namespace"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// CodecOption specifies which type of messages this codec supports
type CodecOption int

const (
	// OptionMethodInvocation is an indication that the codec supports RPC method calls
	OptionMethodInvocation CodecOption = 1 << iota

	// OptionSubscriptions is an indication that the codec suports RPC notifications
	OptionSubscriptions
)

const (
	stopPendingRequestTimeout = 3 * time.Second // give pending requests stopPendingRequestTimeout the time to finish when the server is stopped
	adminService              = "admin"
)

// Server represents a RPC server
type Server struct {
	run          int32
	codecsMu     sync.Mutex
	codecs       *set.Set
	namespaceMgr namespace.NamespaceManager
	admin        *admin.Administrator
	requestMgrMu sync.Mutex
	requestMgr   map[string]*RequestManager
}

// NewServer will create a new server instance with no registered handlers.
func NewServer(nr namespace.NamespaceManager, config *common.Config) *Server {
	server := &Server{
		codecs:       set.New(),
		run:          1,
		namespaceMgr: nr,
		requestMgr:   make(map[string]*RequestManager),
	}
	server.admin = admin.NewAdministrator(nr, config)
	return server
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes the
// response back using the given codec. It will block until the codec is closed or the server is
// stopped. In either case the codec is closed.
func (s *Server) ServeCodec(codec ServerCodec, options CodecOption, ctx context.Context) error {
	defer codec.Close()
	return s.serveRequest(codec, false, options, ctx)
}

// ServeSingleRequest reads and processes a single RPC request from the given codec. It will not
// close the codec unless a non-recoverable error has occurred. Note, this method will return after
// a single request has been processed!
func (s *Server) ServeSingleRequest(codec ServerCodec, options CodecOption) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return s.serveRequest(codec, true, options, ctx)
}

// serveRequest will reads requests from the codec, calls the RPC callback and
// writes the response to the given codec.
//
// If singleShot is true it will process a single request, otherwise it will handle
// requests until the codec returns an error when reading a request (in most cases
// an EOF). It executes requests in parallel when singleShot is false.
func (s *Server) serveRequest(codec ServerCodec, singleShot bool, options CodecOption, ctx context.Context) error {

	var pend sync.WaitGroup

	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Errorf("panic message: %v", err)
			log.Errorf(string(buf))
		}

		s.codecsMu.Lock()
		s.codecs.Remove(codec)
		s.codecsMu.Unlock()

		return
	}()

	s.codecsMu.Lock()
	if atomic.LoadInt32(&s.run) != 1 { // server stopped
		s.codecsMu.Unlock()
		return &common.ShutdownError{}
	}
	s.codecs.Add(codec)
	s.codecsMu.Unlock()

	// test if the server is ordered to stop
	for atomic.LoadInt32(&s.run) == 1 {
		reqs, batch, err := s.readRequest(codec, options)
		// If a parsing error occurred, send an error
		if err != nil {
			// If a parsing error occurred, send an error
			if err.Error() != "EOF" {
				log.Debug(fmt.Sprintf("read error: %v\n", err))
				codec.Write(codec.CreateErrorResponse(nil, "", err))
			}
			// Error or end of stream, wait for requests and tear down
			pend.Wait()
			return nil
		}

		// check if server is ordered to shutdown and return an error
		// telling the client that his request failed.
		if atomic.LoadInt32(&s.run) != 1 {
			err := &common.ShutdownError{}
			if batch {
				resps := make([]interface{}, len(reqs))
				for i, r := range reqs {
					resps[i] = codec.CreateErrorResponse(r.Id, r.Namespace, err)
				}
				codec.Write(resps)
			} else {
				codec.Write(codec.CreateErrorResponse(reqs[0].Id, reqs[0].Namespace, err))
			}
			return nil
		}
		if reqs[0].Service == adminService {
			response := s.handleCMD(reqs[0], codec)
			if response.Error != nil {
				codec.Write(codec.CreateErrorResponse(response.Id, response.Namespace, response.Error))
			} else if response.Reply != nil {
				if err := codec.Write(codec.CreateResponse(response.Id, response.Namespace, response.Reply)); err != nil {
					log.Errorf("%v\n", err)
					codec.Close()
				}
			} else {
				codec.Write(codec.CreateResponse(response.Id, response.Namespace, nil))
			}
			return nil
		}

		if singleShot {
			s.handleReqs(ctx, codec, reqs)
			return nil
		}

		// For multi-shot connections, start a goroutine to serve and loop back
		pend.Add(1)

		go func() {
			defer pend.Done()
			s.handleReqs(ctx, codec, reqs)
		}()

	}
	return nil
}

// Stop will stop reading new requests, wait for stopPendingRequestTimeout to allow pending requests to finish,
// close all codecs which will cancels pending requests/subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log.Notice("RPC Server shutdown initiatied")
		time.AfterFunc(stopPendingRequestTimeout, func() {
			s.codecsMu.Lock()
			defer s.codecsMu.Unlock()
			s.codecs.Each(func(c interface{}) bool {
				c.(ServerCodec).Close()
				return true
			})
			log.Notice("RPC Server shutdown")
		})
	}
}

// readRequest requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (s *Server) readRequest(codec ServerCodec, options CodecOption) ([]*common.RPCRequest, bool, common.RPCError) {
	reqs, batch, err := codec.ReadRawRequest(options)
	if err != nil {
		return nil, batch, err
	}

	if len(reqs) == 0 {
		log.Errorf("no request found.")
		return nil, false, &common.InvalidRequestError{Message: "no request found"}
	}
	reqLen := len(reqs)
	for i := 0; i < reqLen; i += 1 {
		if reqs[i].Namespace == "" {
			reqs[i].Namespace = namespace.DEFAULT_NAMESPACE
		}
	}

	return reqs, batch, nil
}

// handleReqs will handle RPC request array and write result then send to client
func (s *Server) handleReqs(ctx context.Context, codec ServerCodec, reqs []*common.RPCRequest) {
	//log.Error("-----------enter handle batch req---------------")
	number := len(reqs)
	response := make([]interface{}, number)
	//result := make(chan interface{}, number)

	i := 0
	for _, req := range reqs {
		req.Ctx = ctx

		//go func(s *Server, request *common.RPCRequest, codec ServerCodec, result chan interface{}) {
		//	name := request.Namespace
		//	if err := codec.CheckHttpHeaders(name, request.Method); err != nil {
		//		log.Errorf("CheckHttpHeaders error: %v", err)
		//		result <- codec.CreateErrorResponse(request.Id, request.Namespace, &common.CertError{Message: err.Error()})
		//		return
		//	}
		//	var rm *requestManager
		//
		//	s.requestMgrMu.Lock()
		//	if _, ok := s.requestMgr[name]; !ok {
		//		rm = NewRequestManager(name, s, codec)
		//		s.requestMgr[name] = rm
		//		rm.Start()
		//	} else {
		//		s.requestMgr[name].codec = codec
		//		rm = s.requestMgr[name]
		//	}
		//	s.requestMgrMu.Unlock()
		//
		//	rm.requests <- request
		//	result <- <-rm.response
		//	return
		//}(s, req, codec, result)
		if err := codec.CheckHttpHeaders(req.Namespace, req.Method); err != nil {
			log.Errorf("CheckHttpHeaders error: %v", err)
			response[i] = codec.CreateErrorResponse(req.Id, req.Namespace, &common.CertError{Message: err.Error()})
			break
		}
		response[i] = s.handleChannelReq(codec, req)
		i++
	}

	//for i := 0; i < number; i++ {
	//	response[i] = <-result
	//}

	if number == 1 {
		if err := codec.Write(response[0]); err != nil {
			log.Errorf("%v\n", err)
			codec.Close()
		}
	} else {
		if err := codec.Write(response); err != nil {
			log.Errorf("%v\n", err)
			codec.Close()
		}
	}
}

// handleChannelReq implements receiver.handleChannelReq interface to handle request in channel and return jsonrpc response.
func (s *Server) handleChannelReq(codec ServerCodec, req *common.RPCRequest) interface{} {
	r := s.namespaceMgr.ProcessRequest(req.Namespace, req)
	if r == nil {
		log.Debug("No process result")
		return codec.CreateErrorResponse(req.Id, req.Namespace, &common.CallbackError{Message: "no process result"})
	}

	if response, ok := r.(*common.RPCResponse); ok {
		if response.Error != nil && response.Reply == nil {
			return codec.CreateErrorResponse(response.Id, response.Namespace, response.Error)
		} else if response.Error != nil && response.Reply != nil {
			return codec.CreateErrorResponseWithInfo(response.Id, response.Namespace, response.Error, response.Reply)
		} else if response.Reply != nil {
			if response.IsPubSub {
				notifier, supported := NotifierFromContext(req.Ctx)
				if !supported { // interface doesn't support subscriptions (e.g. http)
					return codec.CreateErrorResponse(response.Id, response.Namespace, &common.CallbackError{Message: ErrNotificationsUnsupported.Error()})
				}

				if response.IsUnsub {
					subid := response.Reply.(common.ID)
					if err := notifier.Unsubscribe(subid); err != nil {
						return codec.CreateErrorResponse(response.Id, response.Namespace, &common.SubNotExistError{Message: err.Error()})
					}
					return codec.CreateResponse(response.Id, response.Namespace, true)
				} else {
					// active the subscription after the sub id was successfully sent to the client
					notifier.Activate(response.Reply.(common.ID), req.Service, req.Method, req.Namespace)
				}
			}
			return codec.CreateResponse(response.Id, response.Namespace, response.Reply)

		} else {
			return codec.CreateResponse(response.Id, response.Namespace, nil)
		}
		//} else if response, ok := r.(*common.RPCNotification); ok{
		//	return s.CreateNotification(response.SubId, response.Service, response.Namespace, nil)
	} else {
		log.Errorf("response type invalid, resp: %v\n")
		return codec.CreateErrorResponse(req.Id, req.Namespace, &common.CallbackError{Message: "response type invalid!"})
	}
}

func (s *Server) handleCMD(req *common.RPCRequest, codec ServerCodec) *common.RPCResponse {
	token, method := codec.GetAuthInfo()
	err := s.admin.PreHandle(token, method)
	if err != nil {
		return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Error: &common.InvalidTokenError{Message: err.Error()}}
	}

	cmd := &admin.Command{MethodName: req.Method}
	if args, ok := req.Params.(json.RawMessage); !ok {
		log.Notice("nil parms in json")
		cmd.Args = nil
	} else {
		args, err := SplitRawMessage(args)
		if err != nil {
			return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Error: &common.InvalidParamsError{Message: err.Error()}}
		}
		cmd.Args = args
	}
	if _, ok := s.admin.CmdExecutor[req.Method]; !ok {
		return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Error: &common.MethodNotFoundError{Service: req.Service, Method: req.Method}}
	}
	rs := s.admin.CmdExecutor[req.Method](cmd)
	if rs.Ok == false {
		return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Error: rs.Error}
	}
	return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Reply: rs.Result}

}
