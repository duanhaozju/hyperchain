package msg

import "fmt"
import (
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/hts"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

type SessionMsgHandler struct {
	mchan  chan interface{}
	evmux  *event.TypeMux
	hub    *event.TypeMux
	shts   *hts.ServerHTS
	logger *logging.Logger
}

func NewSessionHandler(blackHole chan interface{}, eventHub *event.TypeMux, peerhub *event.TypeMux, shts *hts.ServerHTS, logger *logging.Logger) *SessionMsgHandler {
	return &SessionMsgHandler{
		mchan:  blackHole,
		evmux:  eventHub,
		hub:    peerhub,
		shts:   shts,
		logger: logger,
	}
}

func (session *SessionMsgHandler) Process() {
	for msg := range session.mchan {
		fmt.Println("got a hello message", string(msg.(pb.Message).Payload))
	}
}

func (session *SessionMsgHandler) Teardown() {
	close(session.mchan)
}

func (session *SessionMsgHandler) Receive() chan<- interface{} {
	return session.mchan
}

func (session *SessionMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	session.logger.Debugf("Got a SESSION Message From: %s, Type: %s", msg.From.Hostname, msg.MessageType.String())
	session.logger.Debugf("Decrypt message for %s", string(msg.From.UUID))
	decPayload, err := session.shts.Decrypt(string(msg.From.UUID), msg.Payload)
	if err != nil {
		session.logger.Warningf("SESSION message payload decrypt failed. msg from %s, namespace %s, type: %s, err %s", msg.From.Hostname, msg.From.Field, msg.MessageType.String(), err.Error())
		return nil, errors.New(fmt.Sprintf("SESSION message payload decrypt failed. msg from %s, namespace %s, type: %s(response) err: %s", msg.From.Hostname, msg.From.Field, msg.MessageType.String(), err.Error()))
	}
	go session.evmux.Post(event.SessionEvent{
		Message: decPayload,
	})
	rsp := &pb.Message{
		MessageType: pb.MsgType_RESPONSE,
	}
	return rsp, nil
}
