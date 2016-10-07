// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:48
// last Modified Author: chenquan
// change log:  1.modified the param of the event register
//		2.add a english-chinese comment
//		3.
//
package peerEventManager

import (
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/peerComm"
	"errors"
	eventHandler "hyperchain/p2p/peerEventHandler"
	"sync"
	"github.com/op/go-logging"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/peerEventManager")
}


// PeerEventManager is a peer event manager which can handle the broadcast event etc.
type PeerEventManager struct {
	peerEventChain chan pb.Message
	// if need event queue, this will be used
	eventQueue *peerComm.Queue
	eventListener map[pb.Message_MsgType] eventHandler.PeerEventHandler
	syncMux sync.Mutex
}

// NewPeerEventManager offer a Event Manager instance
func NewPeerEventManager() *PeerEventManager{
	var peereventManager PeerEventManager
	peereventManager.peerEventChain = make(chan pb.Message)
	peereventManager.eventQueue = peerComm.NewQueueBySize(200)
	peereventManager.eventListener = make(map[pb.Message_MsgType]eventHandler.PeerEventHandler)
	return &peereventManager
}

//RegisterEvent register the event handler
func (this *PeerEventManager)RegisterEvent(msgType pb.Message_MsgType,eHandler eventHandler.PeerEventHandler)error{
	if _,ok := this.eventListener[msgType];ok{
		return errors.New("This event type already has been registered!(duplicate)")
	}else{
		this.eventListener[msgType] = eHandler
		return nil
	}
}

// PostEvent post the event into listen thread
func (this *PeerEventManager) PostEvent(msgType pb.Message_MsgType,message pb.Message)error{
	this.syncMux.Lock()
	defer this.syncMux.Unlock()
	if _,ok := this.eventListener[msgType];ok{
		this.peerEventChain <- message
		return nil
	}else{
		return errors.New("This event type hasn't been registered!")

	}
}

// Start start the listen thread
func(this *PeerEventManager)Start(){
	go this.eventLoop()
}

func (this *PeerEventManager)eventLoop(){
	//if the send peer close the channel, this loop will be break
	for msg := range this.peerEventChain {
		if handler, ok := this.eventListener[msg.MessageType];ok {
			handler.ProcessEvent(&msg)
		} else {
			log.Fatal("错误,该事件未绑定/ERROR,the Event hasn't register")
		}
	}
}