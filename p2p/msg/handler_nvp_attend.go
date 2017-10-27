package msg

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/manager/event"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/hyperchain/hyperchain/p2p/peerevent"
	"github.com/op/go-logging"
)

type NVPAttendMsgHandler struct {
	mchan   chan interface{}
	ev      *event.TypeMux
	peermgr *event.TypeMux
	log     *logging.Logger
}

func NewNVPAttendHandler(blackHole chan interface{}, ev *event.TypeMux, peermgr *event.TypeMux, log *logging.Logger) *NVPAttendMsgHandler {
	return &NVPAttendMsgHandler{
		mchan:   blackHole,
		ev:      ev,
		peermgr: peermgr,
		log:     log,
	}
}

//Process
func (h *NVPAttendMsgHandler) Process() {
	for msg := range h.mchan {
		fmt.Println("got a Attend message", string(msg.(*pb.Message).Payload))
	}
}

//Teardown
func (h *NVPAttendMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(h.mchan)
}

//Receive
func (h *NVPAttendMsgHandler) Receive() chan<- interface{} {
	return h.mchan
}

//Execute
func (h *NVPAttendMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	h.log.Infof("GOT A NVP ATTEND MSG hostname(%s), type: %s \n", msg.From.Hostname, msg.MessageType)
	var isRec bool
	if string(msg.Payload) == "True" {
		isRec = true
	}

	ev := peerevent.S_NVPConnect{
		Namespace:   string(msg.From.Field),
		Hostname:    string(msg.From.Hostname),
		Hash:        common.ToHex(msg.From.UUID),
		IsReconnect: isRec,
	}
	h.peermgr.Post(ev)
	rsp := &pb.Message{
		MessageType: pb.MsgType_RESPONSE,
	}
	return rsp, nil
}
