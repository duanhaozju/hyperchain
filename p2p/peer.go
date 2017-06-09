//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"errors"
	"hyperchain/admittance"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/transport"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	//"fmt"
	"fmt"
	"hyperchain/common"
	"hyperchain/manager/event"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
)

var (
	errPeerClosed = errors.New("this peer was cloesd.")
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope
type Peer struct {
	PeerAddr   *pb.PeerAddr
	Connection *grpc.ClientConn
	LocalAddr  *pb.PeerAddr
	Client     pb.ChatClient
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
}

// NewPeer to create a Peer which with a connection
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance
// NewPeer 用于返回一个新的NewPeer 用于与远端的peer建立连接，这个peer将会存储在peerPool中
// 如果取得相应的连接返回值，将会将peer存储在单例的PeersPool中进行存储
func NewPeer(peerAddr *pb.PeerAddr, localAddr *pb.PeerAddr, TM *transport.TransportManager, cm *admittance.CAManager, namespace string) *Peer {
	logger := common.GetLogger(namespace, "peer")
	return &Peer{
		TM:                TM,
		CM:                cm,
		LocalAddr:         localAddr,
		eventMux:          new(event.TypeMux),
		PeerAddr:          peerAddr,
		IsPrimary:         false,
		logger:            logger,
	}
}

//connect connect method must call after newPeer
func (peer *Peer) Connect(payload []byte, msgType pb.Message_MsgType, isSign bool, callback func(*pb.Message) (interface{}, error)) (interface{}, error) {
	peer.eventSub = peer.eventMux.Subscribe(KeepAliveEvent{}, RetryEvent{}, SelfNarrateEvent{}, RecoveryEvent{}, CloseEvent{}, PendingEvent{})
	opts := peer.CM.GetGrpcClientOpts()
	opts = append(opts, grpc.FailOnNonTempDialError(true))
	conn, err := grpc.Dial(peer.PeerAddr.IP+":"+strconv.Itoa(peer.PeerAddr.Port), opts...)
	if err != nil {
		conn.Close()
		peer.logger.Error("err:", errors.New("Cannot establish a connection!"))
		return nil, err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(conn)

	msg := peer.newMsg(payload, msgType)
	retMessage, err := peer.Client.Chat(context.Background(), msg)
	if err != nil {
		peer.logger.Error(peer.LocalAddr.ID, ">>", peer.PeerAddr.ID, ": cannot establish a connection", err)
		return nil, err
	}
	data, err := callback(retMessage)
	if err != nil {
		peer.logger.Error(peer.LocalAddr.ID, ">>", peer.PeerAddr.ID, ": cannot establish a connection", err)
		return nil, err
	}
	peer.Alive = true
	//TODO self narrate configurable
	go peer.eventMux.Post(SelfNarrateEvent{
		content: fmt.Sprintf("I'am node %d , peer to -> %d (create time %s)", peer.LocalAddr.ID, peer.PeerAddr.ID, time.Now().Format("20060102 15:04:05")),
	})
	go peer.eventMux.Post(KeepAliveEvent{
		Interval: peer.CM.KeepAliveInterval,
	})
	return data, nil
}


// Close the peer connection
// this function should ensure no thread use this thread
// this is not thread safety
func (peer *Peer) Close() error {
	peer.logger.Warning("now close the peer ID: %d, IP: %s Port: %d", peer.PeerAddr.ID, peer.PeerAddr.IP, peer.PeerAddr.Port)
	return peer.Connection.Close()
}

// self Narrate call back function
func (peer *Peer) selfNarrate(content string) {
	for range time.Tick(5 * time.Second) {
		peer.logger.Debug("[SN]", content)
	}
}

//newMsg create a new peer msg
func (peer *Peer) newMsg(payload []byte, msgType pb.Message_MsgType) *pb.Message {
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
func (peer *Peer) setKey(msg *pb.Message) error {
	//verify the rcert and set the status
	if err := peer.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr(msg.From)); err != nil {
peer.logger.Errorf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID, err)
		return errors.New(fmt.Sprintf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID, err))
	}
	return nil
}

// slightly connect test
func (peer *Peer) slightTest(ip string, port int, msg *pb.Message) (*grpc.ClientConn, bool) {
	opts := peer.CM.GetGrpcClientOpts()
	opts = append(opts, grpc.WithTimeout(2*time.Second))
	conn, err1 := grpc.Dial(ip+":"+strconv.Itoa(port), opts...)
	tempClient := pb.NewChatClient(conn)
	_, err2 := tempClient.Chat(context.Background(), msg, grpc.FailFast(true))
	if err1 != nil || err2 != nil || conn == nil {
		if conn == nil {
			conn.Close()
		}
		return conn, false
	} else {
		return conn, true
	}
}

//verify the signature
func (peer *Peer) verSign(msg *pb.Message) error {
	_, err := peer.TM.VerifyMsg(msg)
	return err
}

//HelloHandler hello response handler
func (peer *Peer) HelloHandler(retMsg *pb.Message) (interface{}, error) {
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
func (peer *Peer) ReconnectHandler(retMsg *pb.Message) (interface{}, error) {
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
func (peer *Peer) ReverseHandler(retMsg *pb.Message) (interface{}, error) {
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
func (peer *Peer) NothingHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	return nil, nil
}

func (peer *Peer) IntroHandler(retMsg *pb.Message) (interface{}, error) {
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

func (peer *Peer) AttendHandler(retMsg *pb.Message) (interface{}, error) {
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

func (peer *Peer) AttendNotifyHandler(retMsg *pb.Message) (interface{}, error) {
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

// Chat is a function to send a message to peer,
// this function invokes the remote function peer-to-peer,
// which implements the service that prototype file declares
//
func (peer *Peer) Chat(msg pb.Message) (response *pb.Message, err error) {
	peer.logger.Debug("CHAT:", msg.From.ID, ">>>", peer.PeerAddr.ID)
	if !peer.Alive {
		peer.logger.Warningf("chat failed, %d -> %d ,this.peer was closed.", msg.From.ID, peer.PeerAddr.ID)
		return nil, errPeerClosed
	}
	msg.Payload, err = peer.TM.Encrypt(msg.Payload, peer.PeerAddr)
	if err != nil {
		peer.logger.Errorf("encrypt msg failed(%d -> %d),%v", msg.From.ID, peer.PeerAddr.ID, err)
		return nil, err
	}
	signmsg, err := peer.TM.SignMsg(&msg)
	if err != nil {
		peer.logger.Errorf("sign with secret failed ( %d -> %d ),%v", msg.From.ID, peer.PeerAddr.ID, err)
	}
	response, err = peer.Client.Chat(context.Background(), &signmsg)
	if err != nil {
		peer.Alive = false
		peer.logger.Error(peer.LocalAddr.ID, ">>", peer.PeerAddr.ID, ": response err:", err)
		return nil, err
	}
	// decode the return message
	response.Payload, err = peer.TM.Decrypt(response.Payload, pb.RecoverPeerAddr(response.From))
	if err != nil {
		peer.logger.Errorf("Decrypt the msg faild (%d -> $d) err %v:", msg.From.ID, peer.PeerAddr.ID, err)
		return nil, err
	}
	peer.logger.Debugf("response from: %d, type %v, origin msg is :(%d  >> %d)", response.From.ID, response.MessageType, msg.From.ID, peer.PeerAddr.ID)
	return response, err
}
