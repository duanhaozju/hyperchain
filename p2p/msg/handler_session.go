package msg

import "fmt"
import (
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
	"hyperchain/p2p/hts"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)


type SessionMsgHandler struct {
	mchan chan interface{}
	evmux *event.TypeMux
	shts *hts.ServerHTS
	logger *logging.Logger
}

func NewSessionHandler(blackHole chan interface{},eventHub *event.TypeMux,shts *hts.ServerHTS,logger *logging.Logger)*SessionMsgHandler{
	return &SessionMsgHandler{
		mchan:blackHole,
		evmux:eventHub,
		shts:shts,
		logger:logger,
	}
}

func (session  *SessionMsgHandler) Process() {
	for msg := range session.mchan {
		 fmt.Println("got a hello message", string(msg.(pb.Message).Payload))
		}
}

func (session  *SessionMsgHandler)  Teardown() {
	close(session.mchan)
}

func (session  *SessionMsgHandler) Receive() chan<- interface{}{
	return session.mchan
}

func (session  *SessionMsgHandler) Execute(msg *pb.Message) (*pb.Message,error){
	session.logger.Debugf("GOT a SESSION Message From: %s, Type: %s \n",msg.From.Hostname,msg.MessageType.String())
	session.logger.Debugf("DECRYPTED FOR %s\n",string(msg.From.UUID))
	decPayload:= session.shts.Decrypt(string(msg.From.UUID),msg.Payload)
	if decPayload == nil{
		session.logger.Errorf("SESSION PAYLOAD DECRYPT FAILED. msg from %s, namespace %s",msg.From.Hostname)
		return nil,errors.New("cannot decrypt the Msssge response")
	}
	go session.evmux.Post(event.SessionEvent{
		Message:decPayload,
	})
	rsp  := &pb.Message{
		MessageType:pb.MsgType_RESPONSE,
	}
	return rsp,nil
}
