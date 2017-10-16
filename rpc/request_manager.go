package jsonrpc

import "hyperchain/common"

type requestManager struct {
	namespace string
	receiver  receiver
	codec     ServerCodec
	requests  chan *common.RPCRequest
	response  chan interface{}
	exit      chan interface{}
}

// NewRequestManager creates and returns a new requestManager instance for given namespace name.
func NewRequestManager(namespace string, s *Server, codec ServerCodec) *requestManager {
	return &requestManager{
		namespace: namespace,
		receiver:  s,
		requests:  make(chan *common.RPCRequest),
		response:  make(chan interface{}),
		exit:      make(chan interface{}),
		codec:     codec,
	}
}

// Start creates a goroutine to wait for a new RPC request in the namespace.
func (rm *requestManager) Start() {
	log.Debug("start a new jsonrpc request manager")
	go rm.loop()
}

// Stop will break for loops, requestManager will not receive new RPC request.
func (rm *requestManager) Stop() {
	close(rm.exit)
}

// ProcessRequest will process RPC request.
func (rm *requestManager) ProcessRequest(request *common.RPCRequest) {
	rm.response <- rm.receiver.handleChannelReq(rm.codec, request) //TODO: add timeout detect
}

func (rm *requestManager) loop() {
	for {
		select {
		case request := <-rm.requests:
			rm.ProcessRequest(request)
		case <-rm.exit:
			log.Debugf("Close jsonrpc request queue of namespace: %v", rm.namespace)
			return
		}
	}
}
