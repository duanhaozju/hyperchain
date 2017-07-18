package msg

import (
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
	"fmt"
	"hyperchain/p2p/peerevent"
	"github.com/pkg/errors"
)

type NVPDeleteMsgHandler struct {
	mchan chan  interface{}
	ev *event.TypeMux
	mgrev *event.TypeMux
}

func NewNVPDeleteHandler(blackHole chan interface{},ev *event.TypeMux,mgrev *event.TypeMux)*NVPDeleteMsgHandler{
	return &NVPDeleteMsgHandler{
		mchan:blackHole,
		ev:ev,
		mgrev:mgrev,
	}
}

//Process
func (h  *NVPDeleteMsgHandler) Process() {
	for msg := range h.mchan {
		 fmt.Println("got a Attend message", string(msg.(*pb.Message).Payload))
	}
}

//Teardown
func (h  *NVPDeleteMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(h.mchan)
}

//Receive
func (h *NVPDeleteMsgHandler)Receive() chan<- interface{}{
	return h.mchan
}

//Execute
func (h *NVPDeleteMsgHandler)Execute(msg *pb.Message) (*pb.Message,error){
	fmt.Printf("GOT A NVP DELETE MSG hostname(%s), type: %s \n",msg.From.Hostname,msg.MessageType)
	rsp  := &pb.Message{
		MessageType:pb.MsgType_RESPONSE,
	}
	if msg != nil && msg.Payload != nil{
		VPHash := string(msg.Payload)
		ev := peerevent.EV_DELETE_VP{
			Hash:VPHash,
		}
		go h.mgrev.Post(ev)
		return rsp,nil
	}else{
		return nil,errors.New("in message is nil,")
	}

}
