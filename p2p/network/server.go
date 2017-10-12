package network

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/msg"
	"net"
)

type Server struct {
	selfIdentifier string
	server         *grpc.Server
	// different filed has different different solts
	slots    *msg.MsgSlots
	hostchan chan [2]string
	sec      *Sec
}

// NewServer creates and returns a new hypernet server instance.
func NewServer(identifier string, cn chan [2]string, sec *Sec) *Server {
	return &Server{
		selfIdentifier: identifier,
		slots:          msg.NewMsgSlots(),
		hostchan:       cn,
		sec:            sec,
	}
}

func (s *Server) Claim() string {
	return s.selfIdentifier
}

// StartServer creates a the gRPC server and listens gRPC port.
func (s *Server) StartServer(port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	s.server = grpc.NewServer(s.sec.GetGrpcServerOpts()...)
	if s.server == nil {
		return errors.New("s.server is nil, cannot initialize a grpc.server")
	}
	RegisterChatServer(s.server, *s)
	go s.server.Serve(lis)
	return nil
}

// StopServer stops the gRPC server.
func (s *Server) StopServer() {
	if s.server != nil {
		s.server.Stop()
	}
}

// RegisterSlot registers a handler to slot under the namespace for specific message type,
// if the slot under the namespace doesn't exist, create a new slot instance for the namespace.
// Parameter `filed` is namespace name.
func (s *Server) RegisterSlot(filed string, msgType pb.MsgType, msgHandler msg.MsgHandler) error {
	if s.slots == nil {
		s.slots = msg.NewMsgSlots()
	}

	if slot, err := s.slots.GetSlot(filed); err == nil {

		// if slot under the namespace exists, register handler
		slot.Register(msgType, msgHandler)
	} else {

		// if there is no slot under the namespace, create it and register handler
		slot = msg.NewMsgSlot()
		slot.Register(msgType, msgHandler)
		s.slots.Register(filed, slot)

	}
	go msgHandler.Process()
	return nil
}

// DeregisterSlot deregisters a handler from a slot under given namespace name.
func (s *Server) DeregisterSlot(filed string, msgType pb.MsgType) error {
	slot, e := s.slots.GetSlot(filed)
	if e != nil {
		return e
	}
	slot.DeRegister(msgType)
	return nil
}

// DeregisterSlots deregisters a slot from slot set.
func (s *Server) DeregisterSlots(filed string) {
	if slot, err := s.slots.GetSlot(filed); err == nil {
		slot.Clear()
	}
	s.slots.DeRegister(filed)
}

// Chat chats with remote peer by bi-directional streaming.
func (s Server) Chat(ccServer Chat_ChatServer) error {
	if s.slots == nil {
		return errors.New(fmt.Sprintf("this server (%s) hasn't register any handler.cannot handle this massage", s.selfIdentifier))
	}
	for {
		in, err := ccServer.Recv()
		if err != nil {
			return err
		}
		go func(msg *pb.Message) {
			if msg.From != nil && msg.From.Hostname != nil && msg.From.Extend != nil && msg.From.Extend.IP != nil {
				go func(from, ip string) {
					m := [2]string{from, ip}
					s.hostchan <- m
				}(string(msg.From.Hostname), string(msg.From.Extend.IP))
			}
			logger.Info("chat got a message %+v \n", msg)
			if msg.From == nil || msg.From.Field == nil {
				logger.Errorf("this msg (%+v) hasn't it's from filed, reject! \n", msg)
				return
			}
			slot, err := s.slots.GetSlot(string(msg.From.Field))
			if err != nil {
				logger.Info("got a unkown filed message: %v \n", msg.MessageType)
				return
			}
			handler, err := slot.GetHandler(msg.MessageType)
			if err != nil {
				logger.Info("got a unkown filed message: %v \n", msg.MessageType)
				return
			} else {
				handler.Receive() <- msg
			}
		}(in)
	}
	return nil
}

// Greeting will execute handler for given message type.
func (s Server) Greeting(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	if msg.From != nil && msg.From.Hostname != nil && msg.From.Extend != nil && msg.From.Extend.IP != nil {
		go func(from, ip string) {
			m := [2]string{from, ip}
			s.hostchan <- m
		}(string(msg.From.Hostname), string(msg.From.Extend.IP))
	}
	if msg.From == nil || msg.From.Field == nil {
		return nil, errors.New(fmt.Sprintf("this msg (%+v) hasn't it's from filed, reject!", msg))
	}
	solt, err := s.slots.GetSlot(string(msg.From.Field))
	if err != nil {
		return nil, err
	}
	handler, err := solt.GetHandler(msg.MessageType)
	if err != nil {
		return nil, err
	} else {
		retMsg, err := handler.Execute(msg)
		return retMsg, err
	}
	return nil, errors.New(fmt.Sprintf("This message type is not support, %v", msg.MessageType))
}

// Whisper will execute handler for given message type.
func (s Server) Whisper(ctx context.Context, msg *pb.Message) (*pb.Message, error) {

	// if the server doesn't connect to client, to connect
	if msg.From != nil && msg.From.Hostname != nil && msg.From.Extend != nil && msg.From.Extend.IP != nil {
		go func(from, ip string) {
			m := [2]string{from, ip}
			s.hostchan <- m
		}(string(msg.From.Hostname), string(msg.From.Extend.IP))
	}

	if s.slots == nil {
		return nil, errors.New(fmt.Sprintf("this server (%s) hasn't register any handler.cannot handle this massage", s.selfIdentifier))
	}
	if msg.From == nil || msg.From.Field == nil {
		return nil, errors.New(fmt.Sprintf("this msg (%+v) hasn't it's from filed, reject!", msg))
	}

	solt, err := s.slots.GetSlot(string(msg.From.Field))
	if err != nil {
		return nil, err
	}
	logger.Debugf("Got messsage type :%s from: %s", msg.MessageType, string(msg.From.Hostname))
	handler, err := solt.GetHandler(msg.MessageType)
	if err != nil {
		return nil, err
	} else {

		retMsg, err := handler.Execute(msg)
		return retMsg, err
	}

	return nil, errors.New(fmt.Sprintf("This message type is not support, %v", msg.MessageType))
}

// Discuss transfers the the node health information.
func (s Server) Discuss(ctx context.Context, pkg *pb.Package) (*pb.Package, error) {
	if pkg.Type == pb.ControlType_Close {
		logger.Warning("[Discuss] close this peer (current will not close the remote peer)")
	}
	if pkg.Payload != nil {
		logger.Debugf("[Discuss] receive a discuss msg %s \n", string(pkg.Payload))
	}
	if pkg.Type == pb.ControlType_Notify {
		logger.Noticef("got a message from %s, content is %s", pkg.SrcHost, string(pkg.Payload))
	}
	resp := pb.NewPkg(nil, pb.ControlType_Response)
	return resp, nil
}
