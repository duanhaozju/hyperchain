//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"runtime"
	"sync/atomic"
	"time"

	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"golang.org/x/net/context"
	"gopkg.in/fatih/set.v0"
	"hyperchain/common"
	"hyperchain/namespace"
	"strings"
	"reflect"
	"fmt"
	"sync"
)

const (
	stopPendingRequestTimeout             = 3 * time.Second // give pending requests stopPendingRequestTimeout the time to finish when the server is stopped
	OptionMethodInvocation    CodecOption = 1 << iota       // OptionMethodInvocation is an indication that the codec supports RPC method calls
	// OptionSubscriptions is an indication that the codec suports RPC notifications
	OptionSubscriptions = 1 << iota // support pub sub
	adminService                          = "admin"
)

// CodecOption specifies which type of messages this codec supports
type CodecOption int

// NewServer will create a new server instance with no registered handlers.
func NewServer(nr namespace.NamespaceManager, stopHyperchain chan bool, restartHp chan bool) *Server {
	server := &Server{
		codecs:       set.New(),
		run:          1,
		namespaceMgr: nr,
		requestMgr:   make(map[string]*requestManager),
	}
	server.admin = &Administrator{
		NsMgr:         server.namespaceMgr,
		StopServer:    stopHyperchain,
		RestartServer: restartHp,
	}
	server.admin.Init()
	return server
}

// RPCService gives meta information about the server.
// e.g. gives information about the loaded modules.
type RPCService struct {
	server *Server
}

// serveRequest will reads requests from the codec, calls the RPC callback and
// writes the response to the given codec.
//
// If singleShot is true it will process a single request, otherwise it will handle
// requests until the codec returns an error when reading a request (in most cases
// an EOF). It executes requests in parallel when singleShot is false.
func (s *Server) serveRequest(codec ServerCodec, singleShot bool, options CodecOption) error {

	var pend sync.WaitGroup

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

	// if the codec supports notification include a notifier that callbacks can use
	// to send notification to clients. It is thight to the codec/connection. If the
	// connection is closed the notifier will stop and cancels all active subscriptions.
	if options&OptionSubscriptions == OptionSubscriptions {
		//ctx = context.WithValue(ctx, common.NotifierKey{}, common.NewNotifier(codec))
		ctx = context.WithValue(ctx, NotifierKey{}, NewNotifier())
	}

	s.codecsMu.Lock()
	if atomic.LoadInt32(&s.run) != 1 { // server stopped
		s.codecsMu.Unlock()
		return &common.ShutdownError{}
	}
	s.codecs.Add(codec)
	s.codecsMu.Unlock()

	// test if the server is ordered to stop
	for atomic.LoadInt32(&s.run) == 1 {
		reqs, batch, err := s.readRequest(codec)
		// If a parsing error occurred, send an error
		if err != nil {
			// If a parsing error occurred, send an error
			if err.Error() != "EOF" {
				log.Debug(fmt.Sprintf("read error %v\n", err))
				codec.Write(s.CreateErrorResponse(nil, "", err))
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
					resps[i] = s.CreateErrorResponse(r.Id, r.Namespace, err)
				}
				codec.Write(resps)
			} else {
				codec.Write(s.CreateErrorResponse(reqs[0].Id, reqs[0].Namespace, err))
			}
			return nil
		}
		if reqs[0].Service == adminService {
			response := s.handleCMD(reqs[0])
			if response.Error != nil {
				codec.Write(s.CreateErrorResponse(response.Id, response.Namespace, response.Error))
			} else if response.Reply != nil {
				if err := codec.Write(s.CreateResponse(response.Id, response.Namespace, response.Reply)); err != nil {
					log.Errorf("%v\n", err)
					codec.Close()
				}
			} else {
				codec.Write(s.CreateResponse(response.Id, response.Namespace, nil))
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

func splitRawMessage(args json.RawMessage) ([]string, error) {
	str := string(args[:])
	if len(str) < 4 {
		return nil, errors.New("invalid args")
	}
	str = str[2 : len(str)-2]
	splitstr := strings.Split(str, ",")
	return splitstr, nil
}

func (s *Server) handleCMD(req *common.RPCRequest) *common.RPCResponse {
	if args, ok := req.Params.(json.RawMessage); !ok {
		log.Critical("wrong type not json type")
		return &common.RPCResponse{Reply: "invalid args"}
	} else {
		args, err := splitRawMessage(args)
		if err != nil {
			return &common.RPCResponse{Reply: "invalid cmd"}
		}
		cmd := &Command{
			MethodName: req.Method,
			Args:       args,
		}
		rs := s.admin.CmdExecutor[req.Method](cmd)
		return &common.RPCResponse{Reply: rs}
	}

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

func (s *Server) Start() {
	if atomic.CompareAndSwapInt32(&s.run, 0, 1) {
		log.Notice("RPC Server start initiatied")
	}
}

// readRequest requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (s *Server) readRequest(codec ServerCodec) ([]*common.RPCRequest, bool, common.RPCError) {
	reqs, batch, err := codec.ReadRequestHeaders()
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
	result := make(chan interface{}, number)

	for _, req := range reqs {
		req.Ctx = ctx

		go func(s *Server, request *common.RPCRequest, codec ServerCodec, result chan interface{}){
			name := request.Namespace
			if err := codec.CheckHttpHeaders(name); err != nil {
				log.Errorf("CheckHttpHeaders error: %v", err)
				result <- s.CreateErrorResponse(request.Id, request.Namespace, &common.CertError{Message: err.Error()})
				return
			}
			var rm *requestManager
			if _, ok := s.requestMgr[name]; !ok {
				rm = NewRequestManager(name, s)
				rm.Start()
			}else {
				rm = s.requestMgr[name]
			}
			rm.requests <- request
			result <- (<- rm.response)
			return
		}(s, req, codec, result)
	}

	for i := 0; i < number; i++ {
		response[i] = <- result
	}

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

// handleChannelReq will implement an interface to handle request in channel and return jsonrpc response
func (s *Server) handleChannelReq(req *common.RPCRequest) interface{} {
	r := s.namespaceMgr.ProcessRequest(req.Namespace, req)
	if r == nil {
		log.Debug("No process result")
		return s.CreateErrorResponse(req.Id, req.Namespace, &common.CallbackError{Message:"no process result"})
	}

	if response, ok := r.(*common.RPCResponse); ok {
		if response.Error != nil {
			return s.CreateErrorResponse(response.Id, response.Namespace, response.Error)
		} else if response.Reply != nil {
			return s.CreateResponse(response.Id, response.Namespace, response.Reply)
		} else {
			return s.CreateResponse(response.Id, response.Namespace, nil)
		}
	} else if response, ok := r.(*common.RPCNotification); ok{
		return s.CreateNotification(response.SubId, response.Service, response.Namespace, nil)
	} else {
		log.Errorf("response type invalid, resp: %v\n")
		return s.CreateErrorResponse(req.Id, req.Namespace, &common.CallbackError{Message:"response type invalid!"})
	}
}

// CreateResponse will create a JSON-RPC success response with the given id and reply as result.
func (s *Server) CreateResponse(id interface{}, name string, reply interface{}) interface{} {
	if isHexNum(reflect.TypeOf(reply)) {
		return &JSONResponse{Version: JSONRPCVersion, Namespace: name, Id: id, Code: 0, Message: "SUCCESS", Result: fmt.Sprintf(`%#x`, reply)}
	}
	return &JSONResponse{Version: JSONRPCVersion, Namespace: name, Id: id, Code: 0, Message: "SUCCESS", Result: reply}
}

// CreateErrorResponse will create a JSON-RPC error response with the given id and error.
func (s *Server) CreateErrorResponse(id interface{}, name string, err common.RPCError) interface{} {
	return &JSONResponse{Version: JSONRPCVersion, Namespace: name, Id: id, Code: err.Code(), Message: err.Error()}
}

// CreateErrorResponseWithInfo will create a JSON-RPC error response with the given id and error.
// info is optional and contains additional information about the error. When an empty string is passed it is ignored.
func (s *Server) CreateErrorResponseWithInfo(id interface{}, name string, err common.RPCError, info interface{}) interface{} {
	return &JSONResponse{Version: JSONRPCVersion, Namespace: name, Id: id, Code: err.Code(), Message: err.Error(), Result: info}
}

// CreateNotification will create a JSON-RPC notification with the given subscription id and event as params.
func (s *Server) CreateNotification(subid, service, namespace string, event interface{}) interface{} {
	if isHexNum(reflect.TypeOf(event)) {
		return &jsonNotification{Version: JSONRPCVersion, Namespace: namespace, Method: service + NotificationMethodSuffix,
			Params: jsonSubscription{Subscription: subid, Result: fmt.Sprintf(`%#x`, event)}}
	}

	return &jsonNotification{Version: JSONRPCVersion, Method: namespace + NotificationMethodSuffix,
		Params: jsonSubscription{Subscription: subid, Result: event}}
}