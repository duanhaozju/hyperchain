//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"errors"
	"hyperchain/admittance"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/transport"
	"sync"
	"time"

	"google.golang.org/grpc"
	//"fmt"
	"fmt"
	"hyperchain/common"
	"hyperchain/manager/event"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/p2p/network"
	"hyperchain/p2p/info"
)

var (
	errPeerClosed = errors.New("this peer was cloesd.")
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope
type peer struct {
	PeerAddr   *pb.PeerAddr
	Connection *grpc.ClientConn
	LocalAddr  *pb.PeerAddr
	TM         *transport.TransportManager
	Alive      bool
	chatMux    sync.Mutex
	IsPrimary  bool
	//PeerPool   PeersPool
	Certificate string
	CM          *admittance.CAManager
	eventMux    *event.TypeMux
	eventSub    event.Subscription
	logger  *logging.Logger

	// release 1.3 feature

	client *network.Client
	info *info.Info

}

// release 1.3 feature
//Chat send a stream message to remote peer
func (peer *peer)Chat(in *pb.Message) (*pb.Message, error){
	response,err := peer.client.Greeting(in)
	if err != nil{
		return nil,err
	}
	return response,nil
}

//Greeting send a single message to remote peer
func (peer *peer)Greeting(in *pb.Message) (*pb.Message,error){
	response,err := peer.client.Greeting(in)
	if err != nil{
		return nil,err
	}
	return response,nil
}

//Whisper send a whisper message to remote peer
func (peer *peer)Whisper(in *pb.Message) (*pb.Message,error){
	response,err := peer.client.Greeting(in)
	if err != nil{
		return nil,err
	}
	return response,nil
}


// NewPeer to create a Peer which with a connection
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance
// NewPeer 用于返回一个新的NewPeer 用于与远端的peer建立连接，这个peer将会存储在peerPool中
// 如果取得相应的连接返回值，将会将peer存储在单例的PeersPool中进行存储
func NewPeer(peerAddr *pb.PeerAddr, localAddr *pb.PeerAddr, TM *transport.TransportManager, cm *admittance.CAManager, namespace string) *peer {
	logger := common.GetLogger(namespace, "peer")
	return &peer{
		TM:                TM,
		CM:                cm,
		LocalAddr:         localAddr,
		eventMux:          new(event.TypeMux),
		PeerAddr:          peerAddr,
		IsPrimary:         false,
		logger:            logger,
	}
}


// Close the peer connection
// this function should ensure no thread use this thread
// this is not thread safety
func (peer *peer) Close() error {
	peer.logger.Warning("now close the peer ID: %d, IP: %s Port: %d", peer.PeerAddr.ID, peer.PeerAddr.IP, peer.PeerAddr.Port)
	return peer.Connection.Close()
}


//newMsg create a new peer msg
func (peer *peer) newMsg(payload []byte, msgType pb.Message_MsgType) *pb.Message {
	newMsg := &pb.Message{
		MessageType:  msgType,
		Payload:      payload,
		MsgTimeStamp: time.Now().UnixNano(),
		From:         peer.LocalAddr.ToPeerAddress(),
	}
	signmsg, err := peer.TM.SignMsg(newMsg)
	if err != nil {
		peer.logger.Warningf("sign msg failed err %v", err)
	}
	return &signmsg
}

//setPubkey set share public key
func (peer *peer) setKey(msg *pb.Message) error {
	//verify the rcert and set the status
	if err := peer.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr(msg.From)); err != nil {
peer.logger.Errorf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID, err)
		return errors.New(fmt.Sprintf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID, err))
	}
	return nil
}

//verify the signature
func (peer *peer) verSign(msg *pb.Message) error {
	_, err := peer.TM.VerifyMsg(msg)
	return err
}

//HelloHandler hello response handler
func (peer *peer) HelloHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	if retMsg.MessageType == pb.Message_HELLO_RESPONSE {
		if err := peer.setKey(retMsg); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("ret message isn't Message_HELLO_RESPONSE!")
}

//ReconnectHandler reconnect response handler
func (peer *peer) ReconnectHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	if retMsg.MessageType == pb.Message_RECONNECT_RESPONSE {
		if err := peer.setKey(retMsg); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("ret message isn't Message_RECONNECT_RESPONSE")
}

//ReverseHandler handle reverse return message
func (peer *peer) ReverseHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	if retMsg.MessageType == pb.Message_HELLOREVERSE_RESPONSE {
		if err := peer.setKey(retMsg); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("ret message isn't Message_HELLOREVERSE_RESPONSE")
}

//NothingHandler do noting just offer a call back function
func (peer *peer) NothingHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	return nil, nil
}

func (peer *peer) IntroHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	if retMsg.MessageType == pb.Message_INTRODUCE_RESPONSE {
		routers := new(pb.Routers)
		err := proto.Unmarshal(retMsg.Payload, routers)
		if err != nil {
			return nil, err
		}
		return routers, nil

	}

	return nil, errors.New("ret message isn't Message_INTRODUCE_RESPONSE")

}

func (peer *peer) AttendHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	if retMsg.MessageType == pb.Message_ATTEND_RESPONSE {
		err := peer.setKey(retMsg)
		if err != nil {
			peer.logger.Errorf("generate the share secret key, from node id: %d, error info %s ", retMsg.From.ID, err)
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("ret message isn't Message_ATTEND_RESPONSE")

}

func (peer *peer) AttendNotifyHandler(retMsg *pb.Message) (interface{}, error) {
	if retMsg.MessageType == pb.Message_ATTEND_NOTIFY_RESPONSE {
		err := peer.setKey(retMsg)
		if err != nil {
			peer.logger.Errorf("generate the share secret key, from node id: %d, error info %s ", retMsg.From.ID, err)
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("ret message isn't Message_ATTEND_NOTIFY_RESPONSE")

}
