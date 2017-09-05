package jvm

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	pb "hyperchain/core/vm/jcee/protos"
	"fmt"
)

type Client struct {
	port   int
	host   string
	conn   *grpc.ClientConn
	stream pb.Contract_RegisterClient
	CC     pb.ContractClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		conn: conn,
	}
}

func (c *Client) Connect() error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.host, c.port), grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.conn = conn
	c.CC = pb.NewContractClient(c.conn)
	stream, err := c.CC.Register(context.Background())
	if err != nil {
		return err
	}
	c.stream = stream
	return nil
}

func (c *Client) sendMsg(msg *pb.Message) error {
	return c.stream.Send(msg)
}

//Execute execute transaction in syn way.
func (c *Client) Execute(msg *pb.Message) (*pb.Message, error) {
	c.stream.Send(msg)
	resp, err := c.stream.Recv()
	return resp, err
}

//AsyncExecute execute Request in asynchronous way.
func (c *Client) AsyncExecute(tx *pb.Request) error {
	msg := &pb.Message{
		Type: pb.Message_TRANSACTION,
	}
	payload, err := proto.Marshal(tx)
	if err != nil {
		return err
	}
	msg.Payload = payload
	return c.stream.Send(msg)
}

//BatchExecute execute requests by batch
func (c *Client) BatchExecute(requests []*pb.Request) ([]*pb.Response, error) {
	len := len(requests)
	for i := 0; i < len; i++ {
		err := c.AsyncExecute(requests[i])
		return nil, err
	}
	responses := make([]*pb.Response, len)
	count := 0
	for count < len { //TODO: add timeout jude
		msg, err := c.stream.Recv()
		if err != nil {
			return nil, err
		}
		//TODO: add tx hash judge
		rsp := &pb.Response{}
		proto.Unmarshal(msg.Payload, rsp)
		responses[count] = rsp
		count++
	}
	return responses, nil
}
