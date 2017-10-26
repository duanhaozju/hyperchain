package msg

import (
	"fmt"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/hts"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/hyperchain/hyperchain/p2p/peerevent"
	"github.com/pkg/errors"
)

type NVPExitMsgHandler struct {
	mchan chan interface{}
	mgrev *event.TypeMux
	shts  *hts.ServerHTS
}

func NewNVPExitHandler(blackHole chan interface{}, mgrev *event.TypeMux, shts *hts.ServerHTS) *NVPExitMsgHandler {
	return &NVPExitMsgHandler{
		mchan: blackHole,
		mgrev: mgrev,
		shts:  shts,
	}
}

//Process
func (h *NVPExitMsgHandler) Process() {
	for msg := range h.mchan {
		fmt.Println("got a Attend message", string(msg.(*pb.Message).Payload))
	}
}

//Teardown
func (h *NVPExitMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(h.mchan)
}

//Receive
func (h *NVPExitMsgHandler) Receive() chan<- interface{} {
	return h.mchan
}

//Execute
func (h *NVPExitMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	rsp := &pb.Message{
		MessageType: pb.MsgType_RESPONSE,
	}
	if msg == nil || msg.Payload == nil {
		return nil, errors.New("message is nil, invalid message")
	}
	payload, err := h.shts.Decrypt(string(msg.From.UUID), msg.Payload)
	if err != nil {
		return nil, err
	}
	NVPHash := string(payload)
	ev := peerevent.S_NVP_EXIT{
		NVPHash: NVPHash,
	}
	go h.mgrev.Post(ev)
	return rsp, nil

}
