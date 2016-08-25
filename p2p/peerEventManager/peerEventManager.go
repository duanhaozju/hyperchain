// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerEventManager

import (
	pb "hyperchain-alpha/p2p/peermessage"
	"hyperchain-alpha/p2p/peerComm"
	"errors"
	"log"
	eventHandler "hyperchain-alpha/p2p/peerEventHandler"
)
//the message queue


type PeerEventManager struct {
	peerEventChain chan pb.Message
	eventQueue *peerComm.Queue
	eventListener map[pb.Message_MsgType] eventHandler.PeerEventHandler
}

//提供一个事件管理器实例
func NewPeerEventManager() *PeerEventManager{
	var peereventManager PeerEventManager
	peereventManager.peerEventChain = make(chan pb.Message)
	peereventManager.eventQueue = peerComm.NewQueueBySize(200)
	peereventManager.eventListener = make(map[pb.Message_MsgType]eventHandler.PeerEventHandler)
	return &peereventManager
}

//注册事件监听器
func (pem *PeerEventManager)RegisterEvent(msgType pb.Message_MsgType,eHandler eventHandler.PeerEventHandler)error{
	this := pem
	if _,ok := this.eventListener[msgType];ok{
		return errors.New("This event type already has been registered!")
	}else{
		this.eventListener[msgType] = eHandler
		return nil
	}
}

// PostEvent 将事件发送到监听线程
func (pem *PeerEventManager) PostEvent(msgType pb.Message_MsgType,message pb.Message)error{
	this := pem
	if _,ok := this.eventListener[msgType];ok{
		this.peerEventChain <- message
		return nil
	}else{
		return errors.New("This event type hasn't been registered!")

	}
}

// Start 开启事件监听
func(pem *PeerEventManager)Start(){
	go pem.eventLoop()
}

func (pem *PeerEventManager)eventLoop(){
	this := pem
	//如果发送方关闭将无法range
	for msg := range this.peerEventChain {
		if handler, ok := this.eventListener[msg.MessageType];ok {
			handler.ProcessEvent(&msg)
		} else {
			log.Fatalln("错误,该事件未绑定,Error,the Event hasn't register")
		}
	}
}