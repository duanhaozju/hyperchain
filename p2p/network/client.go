package network

import "hyperchain/p2p/hts"

import (
	"fmt"
	"github.com/looplab/fsm"
	"github.com/pkg/errors"
	"github.com/terasum/pool"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "hyperchain/p2p/message"
	"regexp"
	"time"
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
	cconf *clientConf
}

func connCreator(addr string, options []grpc.DialOption) (interface{}, error) {
	return grpc.Dial(addr, options...)
}

func connCloser(v interface{}) error {
	return v.(*grpc.ClientConn).Close()
}

// NewClient creates and returns a new hypernet client instance.
func NewClient(hostname, addr string, sec *Sec, cconf *clientConf) (*Client, error) {

	poolConfig := &pool.PoolConfig{
		InitialCap:  cconf.connInitCap,    // the minimum number of connections in the connection pool
		MaxCap:      cconf.connUpperlimit, // the maximum number of connections in the connection pool
		Factory:     connCreator,
		Close:       connCloser,
		IdleTimeout: cconf.connIdleTime, // the maximum idle time for the connection.
		// The connection will be closed if this time is exceeded.
		EndPoint: addr, // the address to connect
		Options:  sec.GetGrpcClientOpts(),
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}
	c := &Client{
		MsgChan:  make(chan *pb.Message, 100000),
		addr:     addr,
		hostname: hostname,
		connPool: p,
		sec:      sec,
		//todo those configuration sould be read from configuration
		cconf: cconf,
	}

	// initialize client state machine
	c.initState()
	c.stateMachine.Event(c_EventConnect)
	return c, nil
}

func (c *Client) Close() {
	c.connPool.Release()
}

// Chat chats with remote peer by bi-directional streaming.
func (c *Client) Chat() error {
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

// Greeting sends a message to remote peer and returns response message.
func (c *Client) Greeting(in *pb.Message) (*pb.Message, error) {
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
	resp, err := client.Greeting(context.Background(), in)
	if err != nil {
		logger.Warningf("Greeting failed, client stat change to pending, close this client error info: %s", err.Error())
		if ok, _ := regexp.MatchString(".+closing", err.Error()); ok {
			c.stateMachine.Event(c_EventError)
		}
	}
	return resp, err
}

// Whisper sends a secret message to remote peer.
// Generally, it is used for communication with upper level module.
func (c *Client) Whisper(in *pb.Message) (*pb.Message, error) {
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
	resp, err := client.Whisper(context.Background(), in)
	if err != nil {
		logger.Warningf("Whisper failed, client stat change to pending, close this client error info: %s", err.Error())
		if ok, _ := regexp.MatchString(".+closing", err.Error()); ok {
			c.stateMachine.Event(c_EventError)
		}
	}
	return resp, err
}

// Discuss transfers the the node health information.
func (c *Client) Discuss(in *pb.Package) (*pb.Package, error) {
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
	resp, err := client.Discuss(context.Background(), in)
	if err != nil {
		logger.Warningf("Discuss failed, client stat change to pending, close this client error info: %s", err.Error())
		if ok, _ := regexp.MatchString(".+closing", err.Error()); ok {
			c.stateMachine.Event(c_EventError)
		}
	}
	return resp, err
}
