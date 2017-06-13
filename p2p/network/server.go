package network

import (
	"hyperchain/p2p/message"
	"net"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/pkg/errors"
	"hyperchain/p2p/msg"
	"strconv"
	"fmt"
)

type Server struct {
	server *grpc.Server
	slots map[message.Message_MsgType]msg.MsgHandler
}

func NewServer() *Server{
	return &Server{
		slots:make(map[message.Message_MsgType]msg.MsgHandler),
	}
}

// StartServer start the gRPC server
func (s *Server) StartServer(port int) error {
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	if err != nil {
		return err
	}
	s.server = grpc.NewServer()
	if s.server == nil{
		return errors.New("s.server is nil")
	}
	RegisterChatServer(s.server,*s)
	go s.server.Serve(lis)
	return nil
}

func (s *Server)StopServer(){
	if s.server != nil{
		s.server.Stop()
	}
	fmt.Println("s.server is nil")
}

func (s *Server)RegisterSlot(msgType message.Message_MsgType,msgHandler msg.MsgHandler) error{
	fmt.Println("regisiter a new slot",msgType)
	if s.slots == nil{
		s.slots = make(map[message.Message_MsgType]msg.MsgHandler)
	}
	if _,ok := s.slots[msgType];ok{
		return errors.New("solt already registered.")
	}
	s.slots[msgType] = msgHandler
	go s.slots[msgType].Process()
	fmt.Println(s.slots)
	return nil
}

func (s *Server)DeregisterSlot(msgType message.Message_MsgType) error{
	if s.slots == nil{
		return errors.New("solts hasn't initialed.")
	}
	delete(s.slots,msgType)
	return nil
}

// dibi data tranfer
func (s Server) Chat(ccServer Chat_ChatServer) error{
		for{
			in,err := ccServer.Recv()
			if err != nil {
				return err
			}
			go func(in *message.Message){
				if s,ok := s.slots[in.MessageType];ok{
					s.Recive() <- in
				} else {
					logger.Info("Ingore unknow message type: %v \n",in.MessageType)
				}
			}(in)
		}
	return nil
}

// Greeting doube arrow greeting message transfer
func (s Server) Greeting(ctx context.Context, msg *message.Message) (*message.Message, error){
	fmt.Println("got a greeting msg")
	fmt.Println(msg.MessageType)
	fmt.Println(s.slots)
	 _,ok := s.slots[msg.MessageType]
	fmt.Println(ok)
	if ok{
		fmt.Println("can handle this msg")
		return s.slots[msg.MessageType].Execute(msg)
	}
	return nil,errors.New(fmt.Sprintf("This message type is not support, %v",msg.MessageType))
}

// Wisper Transfer the the node health infomation
func(s Server) Wisper(ctx context.Context, msg *message.Message) (*message.Message, error){
	if _,ok := s.slots[msg.MessageType];ok {
		return s.slots[msg.MessageType].Execute(msg)
	}
	return nil,errors.New(fmt.Sprintf("This message type is not support, %v",msg.MessageType))
}
