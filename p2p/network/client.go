package network

import (
	"google.golang.org/grpc"
	pb "hyperchain/p2p/message"
	"fmt"
	"golang.org/x/net/context"
	"time"
	"hyperchain/p2p/hts"
	"github.com/pkg/errors"
	"github.com/silenceper/pool"
	"hyperchain/p2p/utils"
)

type Client struct {
	addr string
	conn *grpc.ClientConn
	connPool pool.Pool
	client ChatClient
	MsgChan chan *pb.Message
	hts hts.HTS
	connCreator func() (interface{}, error)
	connCloser  func(v interface{}) error
}


//connCreator implements the Hyper Transport Layer security
func connCreator(addr string)(interface{},error){
	/*
	General key agree process
	client                         server
	       -- client hello     --> query client session don't get the session key
	      <-- server hello     --
	       -- client cert      -->
	      <-- server cert      --
	       -- client random    --> generate session key and save the session
	      <-- server done      --
	       -- client pkg       -->
	      <-- server pkg       --

	improve key agree process
	client                        server
               -- client hello     --> query client session and get the session key
	      <-- server done      --
	       -- client pkg       -->
	      <-- server pkg       --

	the session key expired
	client                       server
	      <-- server expired   --
	       -- client hello     -->
	      <-- server hello     --
	      ...
	       -- client pkg       -->
	      <-- server pkg       --


	*/
	return grpc.Dial(addr,grpc.WithInsecure())
}

func NewClient(addr string) (*Client,error){
	connCreator := func(endpoint string) (interface{}, error) { }
	connCloser  := func(v interface{}) error { return v.(*grpc.ClientConn).Close() }
	poolConfig := &pool.PoolConfig{
		InitialCap: 2,
		MaxCap:     10,
		Factory:    connCreator,
		Close:      connCloser,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 15 * time.Second,
		Addr:addr,
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil,err
	}
	return &Client{
		MsgChan: make(chan *pb.Message,100000),
		addr: addr,
		connCreator:connCreator,
		connCloser:connCloser,
		connPool:p,
	},nil
}

func(c *Client)Connect(client ChatClient) error{
	if client != nil{
		c.client = client
		return nil
	}

	//get a connection from pool
	v, err := c.connPool.Get()
	if err != nil {
		logger.Errorf("cannot get a connection from connection pool: %s \n",c.addr)
		fmt.Printf("err: %v",err)
		return err
	}
	//do something
	conn:=v.(*grpc.ClientConn)
	c.conn = conn
	c.client = NewChatClient(conn)
	return nil
}

func(c *Client)Close() error{
	if c.conn != nil{
		return c.conn.Close()
	}else{
		return nil
	}
}


func(c *Client)Chat() (error){
	if c.client == nil{
		logger.Warningf("the client is nil %v \n",c.client)
		return nil
	}
	stream,err := c.client.Chat(context.Background())
	if err != nil{
		logger.Warningf("cannot create stream! %v \n" ,err)
		return err
	}
	for msg := range c.MsgChan{
		fmt.Println("actual send", string(msg.Payload), time.Now().UnixNano())
		err := stream.Send(msg)
		if err != nil{
			fmt.Errorf(err.Error())
		}
	}
	return nil
}
// Greeting doube arrow greeting message transfer
func(c *Client)Greeting(in *pb.Message) (*pb.Message, error){
	if c.client == nil{
		logger.Warningf("the client is nil %v \n",c.client)
		return nil,errors.New(fmt.Sprintf("the client is nil %v \n",c.client))
	}
	in.From.Extend.IP =[]byte(utils.GetLocalIP())
	return c.client.Greeting(context.Background(),in)
}
// Whisper Transfer the the node health information
func(c *Client)Whisper(in *pb.Message) (*pb.Message, error){
	if c.client == nil{
		logger.Warningf("the client is nil %v \n",c.client)
		return nil,errors.New(fmt.Sprintf("the client is nil %v \n",c.client))
	}
	return c.client.Whisper(context.Background(),in)
}
