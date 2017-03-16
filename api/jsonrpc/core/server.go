//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/op/go-logging"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"golang.org/x/net/context"
	"gopkg.in/fatih/set.v0"
	"hyperchain/common"
	"hyperchain/namespace"
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
func NewServer(nr namespace.NamespaceManager) *Server {
	server := &Server{
		codecs:       set.New(),
		run:          1,
		namespaceMgr: nr,
	}

	return server
}

// RPCService gives meta information about the server.
// e.g. gives information about the loaded modules.
type RPCService struct {
	server *Server
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
		return &common.ShutdownError{}
	}
	s.codecs.Add(codec)
	s.codecsMu.Unlock()

	// test if the server is ordered to stop
	for atomic.LoadInt32(&s.run) == 1 {
		reqs, batch, err := s.readRequest(codec)
		if err != nil {
			log.Debugf("%v\n", err)
			codec.Write(codec.CreateErrorResponse(nil, "", err))
			return nil
		}
		//TODO: Warn do not support bath request now
		if len(reqs) == 0 {
			log.Errorf("no request found.")
			return errors.New("no request found")
		}
		if rpcErr := codec.CheckHttpHeaders(reqs[0].Namespace); rpcErr != nil {
			log.Errorf("CheckHttpHeaders error %v", rpcErr)
			return nil
		}

		if atomic.LoadInt32(&s.run) != 1 {
			err := &common.ShutdownError{}
			if batch {
				resps := make([]interface{}, len(reqs))
				for i, r := range reqs {
					resps[i] = codec.CreateErrorResponse(&r.Id, r.Namespace, err)
				}
				codec.Write(resps)
			} else {
				codec.Write(codec.CreateErrorResponse(&reqs[0].Id, reqs[0].Namespace, err))
			}
			return nil
		}
		//TODO: Warn do not support bath request now
		reqs[0].Ctx = ctx
		r := s.namespaceMgr.ProcessRequest(reqs[0].Namespace, &reqs[0])
		if r == nil {
			return errors.New("No process result")
		}
		if response, ok := r.(*common.RPCResponse); ok {
			if response.Error != nil {
				codec.Write(codec.CreateErrorResponse(&response.Id, response.Namespace, response.Error))
			} else if response.Reply != nil {
				if err := codec.Write(codec.CreateResponse(&response.Id, response.Namespace, response.Reply)); err != nil {
					log.Errorf("%v\n", err)
					codec.Close()
				}
			} else {
				codec.Write(codec.CreateResponse(&response.Id, response.Namespace, nil))
			}
		} else {
			log.Errorf("response type invalid, resp: %v", response)
			return errors.New(response.Error.Error())
		}
		return nil

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

// execBatch executes the given requests and writes the result back using the codec.
// It will only write the response back when the last request is processed.
//func (s *Server) execBatch(ctx context.Context, codec ServerCodec, requests []*serverRequest) {
//	responses := make([]interface{}, len(requests))
//	var callbacks []func()
//	for i, req := range requests {
//		fmt.Println("got a request",req.svcname)
//		if req.err != nil {
//			responses[i] = codec.CreateErrorResponse(&req.id, req.err)
//		} else {
//			var callback func()
//			if responses[i], callback = s.handle(ctx, codec, req); callback != nil {
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

// readRequest requests the next (batch) request from the codec. It will return the collection
// of requests, an indication if the request was a batch, the invalid request identifier and an
// error when the request could not be read/parsed.
func (s *Server) readRequest(codec ServerCodec) ([]common.RPCRequest, bool, common.RPCError) {
	reqs, batch, err := codec.ReadRequestHeaders()
	if err != nil {
		return nil, batch, err
	} else {
		reqLen := len(reqs)
		for i := 0; i < reqLen; i += 1 {
			if reqs[i].Namespace == "" {
				reqs[i].Namespace = namespace.DEFAULT_NAMESPACE
			}
		}
		return reqs, batch, nil
	}
}
