package network

import "hyperchain/p2p/hts"

import (
	"google.golang.org/grpc"
	pb "hyperchain/p2p/message"
	"fmt"
	"golang.org/x/net/context"
	"time"
	"github.com/pkg/errors"
	"github.com/terasum/pool"
	"github.com/looplab/fsm"
)

type Client struct {
	addr string
	sec *Sec
	connPool pool.Pool
	MsgChan chan *pb.Message
	hts hts.HTS
	stateMachine  *fsm.FSM


	// configurations
	keepAliveDuration time.Duration
	keepAliveFailTimes int

	pendingDuration time.Duration
	pendingFailTimes int

	//connection pool init capacity
	connInitCap int
	// connection pool upper limit
	connUpperlimit int

}


//connCreator implements the Hyper Transport Layer security
func connCreator(addr string,options []grpc.DialOption)(interface{},error){
	return grpc.Dial(addr,options...)
}

func connCloser(v interface{}) error{
	return v.(*grpc.ClientConn).Close()
}

func NewClient(addr string,sec *Sec) (*Client,error){
	//connCreator := func(endpoint string,options []grpc.DialOption) (interface{}, error) { return grpc.Dial(endpoint,options)}
	//connCloser  := func(v interface{}) error { return v.(*grpc.ClientConn).Close() }
	poolConfig := &pool.PoolConfig{
		InitialCap: 2,
		MaxCap:     10,
		Factory:    connCreator,
		Close:      connCloser,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 15 * time.Second,
		EndPoint:addr,
		Options:sec.GetGrpcClientOpts(),
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil,err
	}
	c := &Client{
		MsgChan: make(chan *pb.Message,100000),
		addr: addr,
		connPool:p,
		sec: sec,
		//todo those configuration sould be read from configuration
		connInitCap:2,
		connUpperlimit:10,
		keepAliveDuration:time.Second*3,
		pendingDuration:time.Second*5,
		pendingFailTimes:3,
		keepAliveFailTimes:3,
	}
	// start fsm
	c.initState()
	return c,nil
}

func(c *Client)Close(){
	c.connPool.Release()
}
// Chat chat remote peer as bidi stream
func(c *Client)Chat() (error){
	connv,err :=c.connPool.Get()
	if err !=  nil{
		logger.Warningf(" cannot get the conn from connection pool (%v) ",c.addr)
		return errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ",c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	stream,err := client.Chat(context.Background())
	if err != nil{
		logger.Warningf("cannot create stream! %v " ,err)
		return err
	}
	for msg := range c.MsgChan{
		logger.Debugf("actual send", string(msg.Payload), time.Now().UnixNano())
		err := stream.Send(msg)
		if err != nil{
			fmt.Errorf(err.Error())
		}
	}
	return nil
}

// Greeting doube arrow greeting message transfer
func(c *Client)Greeting(in *pb.Message) (*pb.Message, error){
	connv,err :=c.connPool.Get()
	if err !=  nil{
		logger.Warningf(" cannot get the conn from connection pool (%v) ",c.addr)
		return nil,errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ",c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	return client.Greeting(context.Background(),in)
}

// Whisper Transfer the the high level information
func(c *Client)Whisper(in *pb.Message) (*pb.Message, error){
	// get client from conn pool
	connv,err :=c.connPool.Get()
	if err !=  nil{
		logger.Warningf(" cannot get the conn from connection pool (%v) ",c.addr)
		return nil,errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ",c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	return client.Whisper(context.Background(),in)
}

// Discuss Transfer the the node health information
func(c *Client)Discuss(in *pb.Package) (*pb.Package, error){
	// get client from conn pool
	connv,err :=c.connPool.Get()
	if err !=  nil{
		logger.Warningf(" cannot get the conn from connection pool (%v) ",c.addr)
		return nil,errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ",c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	return client.Discuss(context.Background(),in)
}


