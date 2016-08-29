package pbft


import (
	"testing"
	"fmt"
	"hyperchain-alpha/event"
	"github.com/golang/protobuf/proto"
	"hyperchain-alpha/protos"
	"hyperchain-alpha/consensus/helper"
)

func TestEvent(t *testing.T){
	msgQ :=new(event.TypeMux)
	h:=helper.NewHelper(msgQ)
	c:=GetPlugin(3, h)
	msg:=protos.Message{
		Type      :0 ,
		Timestamp :233333,
		Payload   : []byte {'a', 'b', 'c', 'd'},
		Id        :22222,
	}
	b,err:=proto.Marshal(msg)
	if err==nil{
		fmt.Println("recvMsg")
		c.RecvMsg(b)
	}

}
