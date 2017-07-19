package jsonrpc

import "hyperchain/common"

type Receiver interface {
	handleChannelReq (codec ServerCodec, rq *common.RPCRequest) interface{}
}

type requestManager struct {
	namespace string
	receiver  Receiver
	codec 	  ServerCodec
	requests  chan *common.RPCRequest
	response  chan interface{}
	exit      chan interface{}
}

func NewRequestManager(namespace string, s* Server, codec ServerCodec) *requestManager {
	return &requestManager{
		namespace: namespace,
		receiver:  s,
		requests:  make(chan *common.RPCRequest),
		response:  make(chan interface{}),
		exit:      make(chan interface{}),
		codec:	   codec,
	}
}

func (rm *requestManager)Start() {
	log.Debug("start a new jsonrpc request manager")
	go rm.Loop()
}

func (rm *requestManager)Stop() {
	close(rm.exit)
}

func (rm *requestManager)ProcessRequest(request *common.RPCRequest) {
	rm.response <- rm.receiver.handleChannelReq(rm.codec, request)//TODO: add timeout detect
}

func (rm *requestManager)Loop(){
	for {
		select {
		case request := <- rm.requests:
			rm.ProcessRequest(request)
		case <-rm.exit:
			log.Debugf("Close jsonrpc request queue of namespace: %v", rm.namespace)
			return
		}
	}
}