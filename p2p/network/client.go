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
	addr         string
	hostname     string
	sec          *Sec
	connPool     pool.Pool
	MsgChan      chan *pb.Message
	hts          hts.HTS
	stateMachine *fsm.FSM
	//configurations
	cconf        *clientConf
}


//connCreator implements the Hyper Transport Layer security
func connCreator(addr string, options []grpc.DialOption) (interface{}, error) {
	return grpc.Dial(addr, options...)
}

func connCloser(v interface{}) error {
	return v.(*grpc.ClientConn).Close()
}

func NewClient(hostname, addr string, sec *Sec, cconf *clientConf) (*Client, error) {
	//connCreator := func(endpoint string,options []grpc.DialOption) (interface{}, error) { return grpc.Dial(endpoint,options)}
	//connCloser  := func(v interface{}) error { return v.(*grpc.ClientConn).Close() }
	poolConfig := &pool.PoolConfig{
		InitialCap: cconf.connInitCap,
		MaxCap:     cconf.connUpperlimit,
		Factory:    connCreator,
		Close:      connCloser,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: cconf.connIdleTime,
		EndPoint:addr,
		Options:sec.GetGrpcClientOpts(),
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}
	c := &Client{
		MsgChan: make(chan *pb.Message, 100000),
		addr: addr,
		hostname: hostname,
		connPool:p,
		sec: sec,
		//todo those configuration sould be read from configuration
		cconf:cconf,
	}
	// start fsm
	c.initState()
	c.stateMachine.Event(c_EventConnect)
	return c, nil
}

func (c *Client)Close() {
	c.connPool.Release()
}
// Chat chat remote peer as bidi stream
func (c *Client)Chat() (error) {
	connv, err := c.connPool.Get()
	if err != nil {
		logger.Warningf(" cannot get the conn from connection pool (%v) ", c.addr)
		return errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ", c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		logger.Warningf("cannot create stream! %v ", err)
		return err
	}
	for msg := range c.MsgChan {
		logger.Debugf("actual send", string(msg.Payload), time.Now().UnixNano())
		err := stream.Send(msg)
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}
	return nil
}

// Greeting doube arrow greeting message transfer
func (c *Client)Greeting(in *pb.Message) (*pb.Message, error) {
	if c.stateMachine.Current() != c_StatWorking {
		logger.Warningf("This client's stat. is not working, ignore messge send.(stat %s, addr %s,hostname: %s)", c.stateMachine.Current(), c.addr, c.hostname)
		return nil, errors.New(fmt.Sprintf("This client's stat. is not working, ignore messge send.(stat %s, addr %s,hostname: %s)", c.stateMachine.Current(), c.addr, c.hostname))
	}
	connv, err := c.connPool.Get()
	if err != nil {
		logger.Warningf(" cannot get the conn from connection pool (%v) ", c.addr)
		return nil, errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ", c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	resp,err :=  client.Greeting(context.Background(), in)
	if err != nil{
		logger.Warningf("Greeting failed, client stat change to pending, close this client error info: %s",err.Error())
		c.stateMachine.Event(c_EventError)
	}
	return resp,err
}

// Whisper Transfer the the high level information
func (c *Client)Whisper(in *pb.Message) (*pb.Message, error) {
	if c.stateMachine.Current() != c_StatWorking {
		logger.Warningf("This client's stat. is not working, ignore messge send.(stat %s, addr %s,hostname: %s)", c.stateMachine.Current(), c.addr, c.hostname)
		return nil, errors.New(fmt.Sprintf("This client's stat. is not working, ignore messge send.(stat %s, addr %s,hostname: %s)", c.stateMachine.Current(), c.addr, c.hostname))
	}
	// get client from conn pool
	connv, err := c.connPool.Get()
	if err != nil {
		logger.Warningf(" cannot get the conn from connection pool (%v) ", c.addr)
		return nil, errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ", c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	resp,err:= client.Whisper(context.Background(), in)
	if err != nil{
		logger.Warningf("Whisper failed, client stat change to pending, close this client error info: %s",err.Error())
		c.stateMachine.Event(c_EventError)
	}
	return resp,err
}

// Discuss Transfer the the node health information
func (c *Client)Discuss(in *pb.Package) (*pb.Package, error) {
	// get client from conn pool
	connv, err := c.connPool.Get()
	if err != nil {
		logger.Warningf(" cannot get the conn from connection pool (%v) ", c.addr)
		return nil, errors.New(fmt.Sprintf("cannot get the conn from connection pool (%v) ", c.addr))
	}
	conn := connv.(*grpc.ClientConn)
	client := NewChatClient(conn)
	//put back the conn into the pool
	defer c.connPool.Put(conn)
	resp,err := client.Discuss(context.Background(), in)
	if err != nil{
		logger.Warningf("Discuss failed, client stat change to pending, close this client error info: %s",err.Error())
		c.stateMachine.Event(c_EventError)
	}
	return resp,err
}


