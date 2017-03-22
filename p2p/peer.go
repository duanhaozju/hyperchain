//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/admittance"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"strconv"
	"sync"
	"time"
	//"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/event"
	"fmt"
)

var (
	errPeerClosed = errors.New("this peer was cloesd.")
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope

type KAV struct {
	StopKAV      chan bool
	KAVFailCount int
	InKAV        bool
}

type Retry struct {
	StopRetry  chan bool
	RetryCount int
	RetryLimit int
}

type Peer struct {
	PeerAddr          *pb.PeerAddr
	Connection        *grpc.ClientConn
	LocalAddr         *pb.PeerAddr
	Client            pb.ChatClient
	TM                *transport.TransportManager
	Alive             bool
	chatMux           sync.Mutex
	IsPrimary         bool
	//PeerPool   PeersPool
	Certificate       string
	CM                *admittance.CAManager
	eventMux          *event.TypeMux
	eventSub          event.Subscription

	//stop chan
	StopKAV           chan bool
	StopRetry         chan bool
	//count
	RetryCount        int
	KAVCount          int

	//retry
	retryTimeLimit    int
	recoveryTimeLimit int
	keepAliveTimeLimit int

	//flags
	InKAV             bool
	InRetry           bool


}

// NewPeer to create a Peer which with a connection
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance
// NewPeer 用于返回一个新的NewPeer 用于与远端的peer建立连接，这个peer将会存储在peerPool中
// 如果取得相应的连接返回值，将会将peer存储在单例的PeersPool中进行存储
func NewPeer(peerAddr *pb.PeerAddr, localAddr *pb.PeerAddr, TM *transport.TransportManager, cm *admittance.CAManager) *Peer {
	return &Peer{
		TM:TM,
		CM:cm,
		LocalAddr:localAddr,
		eventMux:new(event.TypeMux),
		PeerAddr:peerAddr,
		StopKAV:make(chan bool, 1),
		StopRetry:make(chan bool, 1),
		RetryCount:0,
		IsPrimary:false,
		retryTimeLimit:cm.RetryTimeLimit,
		recoveryTimeLimit:cm.RecoveryTimeLimit,

	}
}
//connect connect method must call after newPeer
func (peer *Peer)Connect(payload []byte, msgType pb.Message_MsgType, isSign bool, callback func(*pb.Message)(interface{},error)) (interface{}, error) {
	peer.eventSub = peer.eventMux.Subscribe(KeepAliveEvent{}, RetryEvent{}, SelfNarrateEvent{}, RecoveryEvent{}, CloseEvent{}, PendingEvent{})
	go peer.listening()
	opts := peer.CM.GetGrpcClientOpts()
	opts = append(opts,grpc.FailOnNonTempDialError(true))
	conn, err := grpc.Dial(peer.PeerAddr.IP + ":" + strconv.Itoa(peer.PeerAddr.Port), opts...)
	if err != nil {
		conn.Close()
		log.Error("err:", errors.New("Cannot establish a connection!"))
		return nil, err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(conn)

	msg := peer.newMsg(payload, msgType)
	retMessage, err := peer.Client.Chat(context.Background(), msg)
	if err != nil {
		log.Error(peer.LocalAddr.ID, ">>", peer.PeerAddr.ID, ": cannot establish a connection", err)
		return nil, err
	}
	data, err := callback(retMessage)
	if err != nil {
		log.Error(peer.LocalAddr.ID, ">>", peer.PeerAddr.ID, ": cannot establish a connection", err)
		return nil, err
	}
	peer.Alive = true
	//TODO self narrate configurable
	go peer.eventMux.Post(SelfNarrateEvent{
		content:fmt.Sprintf("I'am node %d , peer to -> %d (create time %s)", peer.LocalAddr.ID, peer.PeerAddr.ID, time.Now().Format("20060102 15:04:05")),
	})
	go peer.eventMux.Post(KeepAliveEvent{
		Interval:peer.CM.KeepAliveInterval,
	})
	return data, nil
}

//listening the event change
func (peer *Peer)listening() {
	for ev := range peer.eventSub.Chan() {
		switch ev.Data.(type) {
		case KeepAliveEvent:
			kavev := ev.Data.(KeepAliveEvent)
			//TODO change the timeout times as configurable
			go peer.keepAlive(kavev.Interval, peer.CM.KeepAliveTimeLimit)
		case RetryEvent:
			retryev := ev.Data.(RetryEvent)
			go peer.retry(retryev.RetryTimeout, retryev.RetryTimes)
		case CloseEvent:
			go peer.Close()
		case SelfNarrateEvent:{
			con := ev.Data.(SelfNarrateEvent)
			go peer.selfNarrate(con.content)
		}
		case RecoveryEvent:{
			recov := ev.Data.(RecoveryEvent)
			go peer.recovery(recov.addr, recov.recoveryTimeout, recov.recoveryTimes)
		}
		}

	}
}

// keep alive call back function
func (peer *Peer)keepAlive(interval time.Duration, timeoutTimes int) {
	peer.InKAV = true
	peer.InRetry = false
	for {
		select {
		case <-time.Tick(interval):
			{
				log.Debugf("[KAV] %d >> %d (failed time: %d)", peer.LocalAddr.ID, peer.PeerAddr.ID, peer.KAVCount)
				kavMsg := peer.newMsg([]byte(fmt.Sprintf("[KAV] %d >> %d", peer.LocalAddr.ID, peer.PeerAddr.ID)), pb.Message_KEEPALIVE)
				_, err := peer.Client.Chat(context.Background(), kavMsg, grpc.FailFast(true))
				if err == nil {
					peer.Alive = true
					continue
				}
				// Thread safe
				peer.KAVCount++
				if peer.KAVCount > timeoutTimes {
					log.Warning("Reach keep alive faild time limit and need to retry connect")
					peer.Alive = false
					peer.KAVCount = 0
					go peer.eventMux.Post(RetryEvent{
						RetryTimeout:peer.CM.RetryTimeout,
						RetryTimes:peer.KAVCount,
					})
					return
				}
			}
		case <-peer.StopKAV:
			{
				peer.Alive = false
				peer.InKAV = false
				return
			}
		}
	}
}

// retry connection call back function
func (peer *Peer)retry(timeout time.Duration, retryTimes int) error {
	peer.InRetry = true
	peer.InKAV = false
	if peer.Alive {
		peer.StopKAV <- true
		go peer.eventMux.Post(KeepAliveEvent{
			Interval:peer.CM.KeepAliveInterval,
		})
		return nil
	}
	//ignore this error, because this error means the connection already closed.
	_ = peer.Connection.Close()
	log.Infof("now retry connect to peer Id: %d, IP: %s, port: %d (retry times: %d)", peer.PeerAddr.ID, peer.PeerAddr.IP, peer.PeerAddr.Port, peer.RetryCount)
	//todo check this connection options
	if conn, bol := peer.slightTest(peer.LocalAddr.IP, peer.LocalAddr.Port, peer.newMsg([]byte("SLIGHTTEST"), pb.Message_KEEPALIVE)); bol {
		//failed, this is failed condition branch
		log.Errorf("retry connect to peer failed,  will retry again after %s ( retry times:%d )", timeout.String(), retryTimes)
		if retryTimes <= peer.retryTimeLimit {
			log.Warning("retry not reach the limit, and go to select...")
			// this channel select will select two channel
			select {
			case <-time.After(timeout):
				{
					log.Warning("go to post retry event")
					go peer.eventMux.Post(RetryEvent{
						RetryTimeout:timeout,
						RetryTimes:retryTimes + 1,

					})
					return nil
				}
			// this channel will write by node, not peer part
			// when a new peer reconnect event happen
			case <-peer.StopRetry:
				{
					return nil
				}
			}

		} else {
			log.Errorf("retry connect to peer failed after %d times,close this peer", retryTimes)
			// change to close state
			go peer.eventMux.Post(CloseEvent{})
			return errors.New(fmt.Sprintf("retry connect to peer failed after %d times,close this peer", retryTimes))

		}
	} else {
		log.Infof("retry connect to peer %d success", peer.PeerAddr.ID)
		//success
		peer.Connection = conn
		peer.Client = pb.NewChatClient(conn)
		peer.Alive = true
		go peer.eventMux.Post(KeepAliveEvent{
			Interval:peer.CM.KeepAliveInterval,
		})
		return nil
	}

}

//reverse connection call back function
func (peer *Peer)recovery(addr *pb.PeerAddr, timeout time.Duration, recoveryTimes int) error {
	log.Criticalf("gointo recovery process (id: %d,ip:%s,port:%d times:%d)", addr.ID, addr.IP, addr.Port, recoveryTimes)
	if conn, bol := peer.slightTest(addr.IP, addr.Port, peer.newMsg([]byte("SLIGHTTEST#RECOVERY"), pb.Message_KEEPALIVE)); bol {
		log.Infof("recovery the peer (id: %d,ip:%s,port:%d) success!", addr.ID, addr.IP, addr.Port)
		// could connect to given address
		client := pb.NewChatClient(conn)
		//ignore the error
		_ = peer.Connection.Close()
		//if in retry process
		if peer.InRetry {
			peer.StopRetry <- true
		}

		// if in keep alive process
		if peer.InKAV {
			peer.StopKAV <- true
		}
		peer.Connection = conn
		peer.Client = client
		peer.Alive = true
		peer.StopKAV = make(chan bool, 1)
		peer.StopRetry = make(chan bool, 1)
		peer.KAVCount = 0
		peer.RetryCount = 0
		go peer.eventMux.Post(KeepAliveEvent{
			Interval:peer.CM.KeepAliveInterval,
		})
		return nil

	} else {
		log.Warningf("cannot recovery the connection to peer: %d (ip: %s port:%d ),and will retry after %s", addr.ID, addr.IP, addr.Port, timeout.String())
		//do some thing
		if recoveryTimes <= peer.recoveryTimeLimit {
			select {
			case <-time.After(timeout):
				{
					log.Warningf("recovery again to peer: %d (ip: %s port:%d ) retry times:%d", addr.ID, addr.IP, addr.Port, recoveryTimes)
					go peer.eventMux.Post(RecoveryEvent{
						addr:addr,
						recoveryTimeout:timeout,
						recoveryTimes:recoveryTimes + 1,

					})
					return nil
				}
			// this channel will write by node, not peer part
			// when a new peer reconnect event happen
			case <-peer.StopRetry:
				{
					return nil
				}
			}
		} else {
			log.Warningf("cannot recovery the connection to peer: %d (ip: %s port:%d )<reach the limit(%d)>,and cancel the recovery process", addr.ID, addr.IP, addr.Port, peer.recoveryTimeLimit)
			return errors.New(fmt.Sprintf("cannot recovery the connection to peer: %d (ip: %s port:%d )<reach the limit(%d)>,and cancel the recovery process", addr.ID, addr.IP, addr.Port, peer.recoveryTimeLimit))

		}
	}

}
// Close the peer connection
// this function should ensure no thread use this thread
// this is not thread safety
func (peer *Peer) Close() error {
	log.Warning("now close the peer ID: %d, IP: %s Port: %d", peer.PeerAddr.ID, peer.PeerAddr.IP, peer.PeerAddr.Port)
	peer.Alive = false
	peer.InKAV = false
	peer.InRetry = false
	peer.KAVCount = 0
	peer.RetryCount = 0
	close(peer.StopKAV)
	close(peer.StopRetry)
	return peer.Connection.Close()
}

// self Narrate call back function
func (peer *Peer)selfNarrate(content string) {
	for range time.Tick(5 * time.Second) {
		log.Debug("[SN]", content)
	}
}

//newMsg create a new peer msg
func (peer *Peer)newMsg(payload []byte, msgType pb.Message_MsgType) *pb.Message {
	newMsg := &pb.Message{
		MessageType:msgType,
		Payload:payload,
		MsgTimeStamp:time.Now().UnixNano(),
		From:peer.LocalAddr.ToPeerAddress(),
	}
	signmsg,err := peer.TM.SignMsg(newMsg)
	if err != nil{
		log.Warningf("sign msg failed err %v",err)
	}
	return &signmsg
}

//setPubkey set share public key
func (peer *Peer)setKey(msg *pb.Message) error {
	//verify the rcert and set the status
	if err := peer.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr( msg.From)); err != nil {
		log.Errorf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID, err)
		return errors.New(fmt.Sprintf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID, err))
	}
	return nil
}

// slightly connect test
func (peer *Peer)slightTest(ip string, port int, msg *pb.Message) (*grpc.ClientConn, bool) {
	opts := peer.CM.GetGrpcClientOpts()
	opts = append(opts,grpc.WithTimeout(2 * time.Second))
	conn, err1 := grpc.Dial(ip + ":" + strconv.Itoa(port),opts...)
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
func (peer *Peer)verSign(msg *pb.Message) error {
	_, err := peer.TM.VerifyMsg(msg)
	return err
}

//HelloHandler hello response handler
func (peer *Peer)HelloHandler(retMsg *pb.Message) (interface{}, error) {
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
func (peer *Peer)ReconnectHandler(retMsg *pb.Message) (interface{}, error) {
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
func (peer *Peer)ReverseHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil,err
	}
	if retMsg.MessageType == pb.Message_HELLOREVERSE_RESPONSE {
		if err := peer.setKey(retMsg); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil,errors.New("ret message isn't Message_HELLOREVERSE_RESPONSE")
}
//NothingHandler do noting just offer a call back function
func (peer *Peer)NothingHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	return nil, nil
}

func (peer *Peer)IntroHandler(retMsg *pb.Message) (interface{}, error) {
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

func (peer *Peer)AttendHandler(retMsg *pb.Message) (interface{}, error) {
	if err := peer.verSign(retMsg); err != nil {
		return nil, err
	}
	if retMsg.MessageType == pb.Message_ATTEND_RESPONSE {
		err := peer.setKey(retMsg)
		if err != nil {
			log.Errorf("generate the share secret key, from node id: %d, error info %s ", retMsg.From.ID, err)
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("ret message isn't Message_ATTEND_RESPONSE")

}

func (peer *Peer)AttendNotifyHandler(retMsg *pb.Message) (interface{}, error) {
	if retMsg.MessageType == pb.Message_ATTEND_NOTIFY_RESPONSE {
		err := peer.setKey(retMsg)
		if err != nil {
			log.Errorf("generate the share secret key, from node id: %d, error info %s ", retMsg.From.ID, err)
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
	log.Debug("CHAT:", msg.From.ID, ">>>", peer.PeerAddr.ID)
	if !peer.Alive {
		log.Warningf("chat failed, %d -> %d ,this.peer was closed.", msg.From.ID, peer.PeerAddr.ID)
		return nil, errPeerClosed
	}
	msg.Payload, err = peer.TM.Encrypt(msg.Payload, peer.PeerAddr)
	if err != nil {
		log.Errorf("encrypt msg failed(%d -> %d),%v", msg.From.ID, peer.PeerAddr.ID, err)
		return nil, err
	}
	signmsg, err := peer.TM.SignMsg(&msg)
	if err != nil {
		log.Errorf("sign with secret failed ( %d -> %d ),%v", msg.From.ID, peer.PeerAddr.ID, err)
	}
	response, err = peer.Client.Chat(context.Background(), &signmsg)
	if err != nil {
		peer.Alive = false
		log.Error(peer.LocalAddr.ID, ">>", peer.PeerAddr.ID, ": response err:", err)
		return nil, err
	}
	// decode the return message
	response.Payload, err = peer.TM.Decrypt(response.Payload, pb.RecoverPeerAddr(response.From))
	if err != nil {
		log.Errorf("Decrypt the msg faild (%d -> $d) err %v:", msg.From.ID, peer.PeerAddr.ID, err)
		return nil, err
	}
	log.Debugf("response from: %d, type %v, origin msg is :(%d  >> %d)", response.From.ID, response.MessageType, msg.From.ID, peer.PeerAddr.ID)
	return response, err
}


