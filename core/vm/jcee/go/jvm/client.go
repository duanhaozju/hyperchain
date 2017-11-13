package jvm

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	pb "github.com/hyperchain/hyperchain/core/vm/jcee/protos"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"time"
)

type Client struct {
	config *common.Config
	conn   *grpc.ClientConn
	stream pb.Contract_RegisterClient
	CC     pb.ContractClient
	addr   string
	logger *logging.Logger
}

func NewClient(conf *common.Config) *Client {
	return &Client{
		config: conf,
		logger: common.GetLogger(conf.GetString(common.NAMESPACE), "jvmclient"),
	}
}

func (c *Client) Connect() error {
	retryTime := 10
	for ; c.stream == nil && retryTime > 0; retryTime-- {
		c.addr = fmt.Sprintf("localhost:%s", c.config.GetString(common.JVM_PORT))
		c.logger.Debugf("%d times try to connect to %s", retryTime, c.addr)
		conn, err := grpc.Dial(c.addr, grpc.WithInsecure())
		if err != nil {
			continue
		}
		c.conn = conn
		c.CC = pb.NewContractClient(c.conn)
		stream, err := c.CC.Register(context.Background())
		if err != nil {
			continue
		}
		c.stream = stream
		c.logger.Criticalf("jvm client connect to %s successful", c.addr)
		if c.stream == nil {
			time.Sleep(1 * time.Second)
		}
		return nil
	}
	if c.stream == nil {
		return nil
	} else {
		c.stream = nil
		return fmt.Errorf("jvm client connect to server failed")
	}
}

func (c *Client) sendMsg(msg *pb.Message) error {
	return c.stream.Send(msg)
}

//Execute execute transaction in syn way.
func (c *Client) SyncExecute(req *pb.Request) (*pb.Response, error) {
	if c.stream == nil {
		err := c.Connect()
		if err != nil {
			return &pb.Response{
				Ok:     false,
				Result: []byte(err.Error()),
			}, err
		}
	}

	msg := &pb.Message{
		Type: pb.Message_TRANSACTION,
	}
	payload, err := proto.Marshal(req)
	msg.Payload = payload
	c.stream.Send(msg)
	resp, err := c.stream.Recv()
	if err != nil {
		return &pb.Response{
			Ok:     false,
			Result: []byte(err.Error()),
		}, err
	}
	rsp := &pb.Response{}
	err = proto.Unmarshal(resp.Payload, rsp)
	if err != nil {
		return &pb.Response{
			Ok:     false,
			Result: []byte(err.Error()),
		}, err
	}
	return rsp, err
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

func (c *Client) HeartBeat() (*pb.Response, error) {
	return c.CC.HeartBeat(context.Background(), &pb.Request{}, grpc.FailFast(true))
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

func (c *Client) Addr() string {
	return c.addr
}

func (c *Client) Close() error {
	return c.conn.Close()
}
