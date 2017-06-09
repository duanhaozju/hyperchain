package network

import (
	"google.golang.org/grpc"
	"hyperchain/p2p/message"
	"fmt"
	"golang.org/x/net/context"
	"time"
)

type Client struct {
	conn *grpc.ClientConn
	client ChatClient
	MsgChan chan *message.Message
}

func NewClient() *Client{
	return &Client{
		MsgChan: make(chan *message.Message,100000),
	}
}

func(c *Client)Connect(){
	conn, err := grpc.Dial("127.0.0.1:50012",grpc.WithInsecure())
	if err != nil {
		fmt.Printf("err: %v",err)
		if conn != nil{
			conn.Close()

		}
		return
	}
	c.conn = conn
	c.client = NewChatClient(conn)
}
var Close chan struct{}
func init(){
	Close=make(chan struct{} )
}
func(c *Client)Chat() (error){
	if c.client == nil{
		fmt.Printf("the client is nil %v \n",c.client)
		return nil
	}
	stream,err := c.client.Chat(context.Background())
	if err != nil{
		fmt.Printf("cannot create stream! %v \n" ,err)
		return err
	}
	go func(){
		for msg := range c.MsgChan{
			fmt.Println("actual send", string(msg.Payload), time.Now().UnixNano())
			err := stream.Send(msg)
			if err != nil{
				fmt.Errorf(err.Error())
			}
		}
	}()
	for msg := range c.MsgChan{
		fmt.Println("actual send", string(msg.Payload), time.Now().UnixNano())
		err := stream.Send(msg)
		if err != nil{
			fmt.Errorf(err.Error())
		}
	}

	fmt.Println("client chat closed")
	close(Close)
return nil
}
// Greeting doube arrow greeting message transfer
func(c *Client)Greeting(ctx context.Context, in *message.Message, opts ...grpc.CallOption) (*message.Message, error){
	return nil,nil
}
// Wisper Transfer the the node health infomation
func(c *Client)Wisper(ctx context.Context, in *message.Message, opts ...grpc.CallOption) (*message.Message, error){
	return nil,nil
}
