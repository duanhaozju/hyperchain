package network

import (
	"net"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/pkg/errors"
	"hyperchain/p2p/msg"
	pb "hyperchain/p2p/message"
	"strconv"
	"fmt"
)

type Server struct {
	hostname string
	server *grpc.Server
	// different filed has different different solts
	slots *msg.MsgSlots
}

func NewServer(hostname string) *Server{
	return &Server{
		hostname:hostname,
		slots:msg.NewMsgSlots(),
	}
}

func(s *Server) Claim() string{
	return s.hostname
}

// StartServer start the gRPC server
func (s *Server) StartServer(port int) error {
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	if err != nil {
		return err
	}
	s.server = grpc.NewServer()
	if s.server == nil{
		return errors.New("s.server is nil, cannot initialize a grpc.server")
	}
	RegisterChatServer(s.server,*s)
	go s.server.Serve(lis)
	return nil
}

func (s *Server)StopServer(){
	if s.server != nil{
		s.server.Stop()
	}
}

func (s *Server)RegisterSlot(filed string,msgType pb.MsgType,msgHandler msg.MsgHandler) error{
	if s.slots == nil{
		s.slots =  msg.NewMsgSlots()
	}

	if slot,err:= s.slots.GetSlot(filed);err == nil{
		slot.Register(msgType,msgHandler)
	}else{
		slot = msg.NewMsgSlot()
		slot.Register(msgType,msgHandler)
		s.slots.Register(filed,slot)

	}
	go msgHandler.Process()
	return nil
}

func (s *Server)DeregisterSlot(filed string,msgType pb.MsgType) error{
	slot,e :=s.slots.GetSlot(filed)
	if e != nil{
		return e
	}
	slot.DeRegister(msgType)
	return nil
}

func (s *Server)DeregisterSlots(filed string){
	if slot,err :=  s.slots.GetSlot(filed);err == nil{
		slot.Clear()
	}
	s.slots.DeRegister(filed)
}


// dibi data tranfer
func (s Server) Chat(ccServer Chat_ChatServer) error{
	if s.slots == nil{
		return errors.New(fmt.Sprintf("this server (%s) hasn't register any handler.cannot handle this massage",s.hostname))
	}
	for{
		in,err := ccServer.Recv()
		if err != nil {
			return err
		}
		go func(in *pb.Message){
			fmt.Printf("chat got a message %+v \n",in)
			if in.From == nil || in.From.Filed == nil{
				logger.Errorf("this msg (%+v) hasn't it's from filed, reject! \n",in)
				return
			}
			slot,err := s.slots.GetSlot(string(in.From.Filed))
			if err != nil{
				logger.Info("got a unkown filed message: %v \n",in.MessageType)
				return
			}
			handler,err  := slot.GetHandler(in.MessageType)
			if err != nil{
				logger.Info("got a unkown filed message: %v \n",in.MessageType)
				return
			}else{
				handler.Receive() <- in
			}
		}(in)
	}
	return nil
}

// Greeting doube arrow greeting message transfer
func (s Server) Greeting(ctx context.Context, msg *pb.Message) (*pb.Message, error){
	fmt.Printf("greeting got a message %+v \n",msg)
	if s.slots == nil{
		return nil,errors.New(fmt.Sprintf("this server (%s) hasn't register any handler.cannot handle this massage",s.hostname))
	}
	if msg.From == nil || msg.From.Filed == nil{
		return nil,errors.New(fmt.Sprintf("this msg (%+v) hasn't it's from filed, reject!",msg))
	}
	slot,err := s.slots.GetSlot(string(msg.From.Filed))
	if err != nil{
		return nil,err
	}
	handler,err := slot.GetHandler(msg.MessageType)
	if err !=nil{
		return nil,err
	}else{
		fmt.Println("greeting handler got the message, and execute it ")
		retMsg,err := handler.Execute(msg)
		return retMsg,err
	}
	return nil,errors.New(fmt.Sprintf("This message type is not support, %v",msg.MessageType))
}

// Wisper Transfer the the node health infomation
func(s Server) Whisper(ctx context.Context, msg *pb.Message) (*pb.Message, error){
	fmt.Printf("whisper got a message %+v \n",msg)
	if s.slots == nil{
		return nil,errors.New(fmt.Sprintf("this server (%s) hasn't register any handler.cannot handle this massage",s.hostname))
	}
	if msg.From == nil || msg.From.Filed == nil{
		return nil,errors.New(fmt.Sprintf("this msg (%+v) hasn't it's from filed, reject!",msg))
	}
	solt,err := s.slots.GetSlot(string(msg.From.Filed))
	if err != nil{
		return nil,err
	}
	handler,err := solt.GetHandler(msg.MessageType)
	if err !=nil{
		return nil,err
	}else{
		retMsg,err := handler.Execute(msg)
		return retMsg,err
	}
	return nil,errors.New(fmt.Sprintf("This message type is not support, %v",msg.MessageType))
}
