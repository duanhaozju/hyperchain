package network

import (
	"testing"
	"hyperchain/p2p/message"
	"fmt"
	"time"
	"strconv"
)

func TestClient_Chat(t *testing.T) {
	cl := NewClient("127.0.0.1:8081")
	cl.Connect()
	go cl.Chat()
	go func(){
		for i:=0; i<10000; i++{
			msg := &message.Message{
				MessageType:message.Message_HELLO,
				Payload:[]byte("client msg_1_"+strconv.Itoa(i)),
			}
			cl.MsgChan <- msg
			fmt.Println("push data",time.Now().UnixNano(),i)
		}
	}()
	go func(){
		for i:=0; i<10000; i++{
			msg := &message.Message{
				MessageType:message.Message_KEEPALIVE,
				Payload:[]byte("client msg_2_"+strconv.Itoa(i)),
			}
			cl.MsgChan <- msg
			fmt.Println("push data",time.Now().UnixNano(),i)
		}
	}()
	fmt.Println("-------------------")
	//for t := range time.Tick(time.Second){
	//	fmt.Println(" tick",t.UnixNano())
	//}
}
