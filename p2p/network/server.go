package network

import (
	"net"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/pkg/errors"
	"hyperchain/p2p/msg"
	pb "hyperchain/p2p/message"
	"fmt"
	"github.com/op/go-logging"
)

type Server struct {
	hostname string
	server *grpc.Server
	// different filed has different different solts
	slots *msg.MsgSlots
	hostchan chan [2]string
	sec *Sec
	logger *logging.Logger
}

func NewServer(hostname string,cn chan [2]string,sec *Sec,logger *logging.Logger) *Server{
	return &Server{
		hostname:hostname,
		slots:msg.NewMsgSlots(),
		hostchan:cn,
		sec:sec,
		logger:logger,
	}
}

func(s *Server) Claim() string{
	return s.hostname
}

// StartServer start the gRPC server
func (s *Server) StartServer(port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	s.server = grpc.NewServer(s.sec.GetGrpcServerOpts()...)
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
		go func(msg *pb.Message){
			if msg.From!= nil && msg.From.Hostname != nil && msg.From.Extend!= nil && msg.From.Extend.IP !=nil{
				go func(from,ip string){
					m := [2]string{from,ip}
					s.hostchan <- m
				}(string(msg.From.Hostname),string(msg.From.Extend.IP))
			}
			fmt.Printf("chat got a message %+v \n", msg)
			if msg.From == nil || msg.From.Field == nil{
				s.logger.Errorf("this msg (%+v) hasn't it's from filed, reject! \n", msg)
				return
			}
			slot,err := s.slots.GetSlot(string(msg.From.Field))
			if err != nil{
				s.logger.Info("got a unkown filed message: %v \n", msg.MessageType)
				return
			}
			handler,err  := slot.GetHandler(msg.MessageType)
			if err != nil{
				s.logger.Info("got a unkown filed message: %v \n", msg.MessageType)
				return
			}else{
				handler.Receive() <- msg
			}
		}(in)
	}
	return nil
}

// Greeting doube arrow greeting message transfer
func (s Server) Greeting(ctx context.Context, msg *pb.Message) (*pb.Message, error){
	if msg.From!= nil && msg.From.Hostname != nil && msg.From.Extend!= nil && msg.From.Extend.IP !=nil{
		go func(from,ip string){
			m := [2]string{from,ip}
			s.hostchan <- m
		}(string(msg.From.Hostname),string(msg.From.Extend.IP))
	}
	if s.slots == nil{
		return nil,errors.New(fmt.Sprintf("this server (%s) hasn't register any handler.cannot handle this massage",s.hostname))
	}
	if msg.From == nil || msg.From.Field == nil{
		return nil,errors.New(fmt.Sprintf("this msg (%+v) hasn't it's from filed, reject!",msg))
	}
	slot,err := s.slots.GetSlot(string(msg.From.Field))
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
	if msg.From!= nil && msg.From.Hostname != nil && msg.From.Extend!= nil && msg.From.Extend.IP !=nil{
		go func(from,ip string){
			//fmt.Println("check to reverse...",from,ip)
			m := [2]string{from,ip}
			s.hostchan <- m
		}(string(msg.From.Hostname),string(msg.From.Extend.IP))
	}

	if s.slots == nil{
		return nil,errors.New(fmt.Sprintf("this server (%s) hasn't register any handler.cannot handle this massage",s.hostname))
	}
	if msg.From == nil || msg.From.Field == nil{
		return nil,errors.New(fmt.Sprintf("this msg (%+v) hasn't it's from filed, reject!",msg))
	}
	solt,err := s.slots.GetSlot(string(msg.From.Field))
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
