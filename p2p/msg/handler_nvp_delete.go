package msg

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/hts"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/hyperchain/hyperchain/p2p/peerevent"
	"github.com/pkg/errors"
)

type NVPDeleteMsgHandler struct {
	mchan chan interface{}
	ev    *event.TypeMux
	mgrev *event.TypeMux
	shts  *hts.ServerHTS
}

func NewNVPDeleteHandler(blackHole chan interface{}, ev *event.TypeMux, mgrev *event.TypeMux, shts *hts.ServerHTS) *NVPDeleteMsgHandler {
	return &NVPDeleteMsgHandler{
		mchan: blackHole,
		ev:    ev,
		mgrev: mgrev,
		shts:  shts,
	}
}

//Process
func (h *NVPDeleteMsgHandler) Process() {
	for msg := range h.mchan {
		fmt.Println("got a Attend message", string(msg.(*pb.Message).Payload))
	}
}

//Teardown
func (h *NVPDeleteMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(h.mchan)
}

//Receive
func (h *NVPDeleteMsgHandler) Receive() chan<- interface{} {
	return h.mchan
}

//Execute
func (h *NVPDeleteMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
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
	VPHash := common.Bytes2Hex(payload)
	ev := peerevent.S_DELETE_VP{
		Hash: VPHash,
	}
	go h.mgrev.Post(ev)
	return rsp, nil

}
