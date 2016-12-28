//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/event"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/recovery"
	"net"
	"strconv"
	"time"
	"hyperchain/membersrvc"
	"sync"
	"errors"
	"math"
	"hyperchain/p2p/transport"
	"hyperchain/core/crypto/primitives"
	"hyperchain/p2p/transport/ecdh"
	"crypto/elliptic"
)


type Node struct {
	localAddr          *pb.PeerAddr
	gRPCServer         *grpc.Server
	higherEventManager *event.TypeMux
	//common information
	IsPrimary          bool
	delayTable         map[int]int64
	delayTableMutex    sync.RWMutex
	DelayChan          chan UpdateTable
	sentEvent          bool
	attendChan         chan int
	PeersPool          PeersPool
	N                  int
	DelayTableMutex    sync.Mutex
	TEM                transport.TransportEncryptManager

}

type UpdateTable struct {
	updateID   int
	updateTime int64
}

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(localAddr *pb.PeerAddr, hEventManager *event.TypeMux, TEM transport.TransportEncryptManager, peersPool PeersPool) *Node {
	var newNode Node
	newNode.localAddr = localAddr
	newNode.TEM = TEM
	newNode.higherEventManager = hEventManager
	newNode.PeersPool = peersPool
	newNode.attendChan = make(chan int)
	newNode.sentEvent = false
	newNode.delayTable = make(map[int]int64)
	newNode.DelayChan = make(chan UpdateTable)
	//listen the update
	go newNode.UpdateDelayTableThread();

	log.Debug("node start ...")
	log.Debug("local node addr", localAddr)
	return &newNode
}

//监听节点状态更新线程
func (node *Node)UpdateDelayTableThread(){
	for v := range node.DelayChan{
		if v.updateID > 0{
		node.delayTableMutex.Lock()
		node.delayTable[v.updateID] = v.updateTime
		node.delayTableMutex.Unlock();
		}

	}
}


//新节点需要监听相应的attend类型
func (node *Node)attendNoticeProcess(N int) {
	isPrimaryConnectFlag := false
	f := int(math.Floor(float64((N - 1) / 3)))
	num := 0
	for {
		select {
		case attendFlag := <-node.attendChan:
			{
				log.Debug("Connect to a new peer ... N:", N, "f", f, "num", num)
				if attendFlag == 1 {
					num += 1
					if num >= (N - f) && !node.sentEvent && isPrimaryConnectFlag{
						//TODO 修改向上post的消息类型
						log.Debug("new node has online ")
						node.higherEventManager.Post(event.AlreadyInChainEvent{})
						node.sentEvent = true
						num = 0

					}
				} else if attendFlag == 2{
					isPrimaryConnectFlag =true;
					if num >= (N - f) && !node.sentEvent && isPrimaryConnectFlag{
						//TODO 修改向上post的消息类型
						log.Debug("new node has online ")
						node.higherEventManager.Post(event.AlreadyInChainEvent{})
						node.sentEvent = true
						num = 0

					}

			}else{
					log.Warning("invalid connection ... N:", N, "f", f, "num", num)
				}
		}
		}
	}
}


func (node *Node) GetNodeAddr() *pb.PeerAddr {
	return node.localAddr
}

// GetNodeID which init by new function
func (this *Node) GetNodeHash() string {
	return this.localAddr.Hash
}
func (node *Node) GetNodeID() int {
	return node.localAddr.ID
}

// Chat Implements the ServerSide Function
func (node *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	var response pb.Message
	response.MessageType = pb.Message_RESPONSE
	response.MsgTimeStamp = time.Now().UnixNano()
	response.From = node.localAddr.ToPeerAddress()

	//验签
	signPub := node.TEM.GetSignPublicKey(msg.From.Hash)
	ecdsaEncrypto := primitives.NewEcdsaEncrypto("ecdsa")
	bol,_ := ecdsaEncrypto.VerifySign(signPub,msg.Payload,msg.Signature.Signature)

	if bol == false {
		return &response,errors.New("signature is wrong!!")
	}


	//handle the message
	log.Debug("MSG Type: ", msg.MessageType)
	switch msg.MessageType {
	case pb.Message_HELLO:
		{
			log.Debug("=================================")
			log.Debug("negotiating key")
			log.Debug("local addr is ", node.localAddr)
			log.Debug("remote addr is", msg.From.ID, msg.From.IP, msg.From.Port)
			log.Debug("=================================")
			response.MessageType = pb.Message_HELLO_RESPONSE
			//review 协商密钥

			//TODO
			ecertByte := msg.Signature.Ecert
			rcertByte := msg.Signature.Rcert
			var verifyEcert bool
			var verifyRcert bool

			eca,getErr1 := primitives.GetConfig("./config/cert/eca.ca")
			if getErr1 != nil{
				log.Error("cannot read ecert.",getErr1)
			}
			ecaBtye := []byte(eca)

			rca,getErr2 := primitives.GetConfig("./config/cert/rca.ca")
			if getErr2 != nil{
				log.Error("cannot read ecert.",getErr2)
			}
			rcaByte := []byte(rca)

			ecaPem := primitives.ParseCertificate(string(ecaBtye))
			rcaPem := primitives.ParseCertificate(string(rcaByte))

			if len(ecertByte) == 0{
				return &response,errors.New("ecert is miss.")
			}else {
				ecert := primitives.ParseCertificate(string(ecertByte))
				verifyEcert = primitives.VerifySignature(ecert,ecaPem)

			}

			if len(rcertByte) == 0{
				node.TEM.SetIsVerified(false,msg.From.Hash)
			}else {
				rcert := primitives.ParseCertificate(string(rcertByte))
				verifyRcert = primitives.VerifySignature(rcert,rcaPem)
				node.TEM.SetIsVerified(verifyRcert,msg.From.Hash)
			}
			remotePublicKey := msg.Payload
			genErr := node.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)

			if !verifyEcert {
				return &response,errors.New("ecert is wrong.")
			}

			e := ecdh.NewEllipticECDH(elliptic.P384())
			pub,_ := e.Unmarshal(msg.Payload)
			node.TEM.SetSignPublicKey(pub,msg.From.Hash)
			if genErr != nil {
				log.Error("gen sec error", genErr)
			}
			log.Notice("remote addr hash：", msg.From.Hash)
			log.Notice("negotiated key is ", node.TEM.GetSecret(msg.From.Hash))
			log.Notice("remote publickey is:",remotePublicKey)
			//every times get the public key is same
			transportPublicKey := node.TEM.GetLocalPublicKey()
			//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
			response.Payload = transportPublicKey
			//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
			//REVIEW This no need to call hello event handler
			//判断是否需要反向建立链接

			return &response, nil
		}
	case pb.Message_HELLO_RESPONSE:
		{
			log.Warning("Invalidate HELLO_RESPONSE message")
		}
	case pb.Message_RECONNECT:
		{
			log.Warning("node is reconnecting.")
			response.MessageType = pb.Message_RECONNECT_RESPONSE
			remotePublicKey := msg.Payload
			genErr := node.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("gen sec error", genErr)
			}
			log.Warning("reconnect the remote node id:", msg.From.ID, node.TEM.GetSecret(msg.From.Hash))
			//every times get the public key is same
			transportPublicKey := node.TEM.GetLocalPublicKey()
			//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
			//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
			//REVIEW This no need to call hello event handler
			//判断是否需要反向建立链接需要重新建立新链接
			go node.reconnect(msg)
			response.Payload = transportPublicKey

		}
	case pb.Message_RECONNECT_RESPONSE:
		{
			log.Debug("=================================")
			log.Debug("exchange the secret")
			log.Debug("local node is ", node.localAddr)
			log.Debug("remote node is", msg.From.ID, msg.From.IP, msg.From.Port)
			log.Debug("=================================")
			response.MessageType = pb.Message_HELLO_RESPONSE
			remotePublicKey := msg.Payload
			genErr := node.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("gen sec error", genErr)
			}
			log.Warning("Message_HELLO_RESPONSE remote id:", msg.From.ID, node.TEM.GetSecret(msg.From.Hash))
			//every times get the public key is same
			transportPublicKey := node.TEM.GetLocalPublicKey()
			//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
			response.Payload = transportPublicKey
			//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
			//REVIEW This no need to call hello event handler
			//判断是否需要反向建立链接

			return &response, nil

		}
	case pb.Message_INTRODUCE:
		{
			//返回路由表信息
			response.MessageType = pb.Message_INTRODUCE_RESPONSE
			routers := node.PeersPool.ToRoutingTable()
			response.Payload, _ = proto.Marshal(&routers)
		}
	case pb.Message_INTRODUCE_RESPONSE:
		{
			log.Warning("节点已经接受请求", msg.From)
			//this.higherEventManager.Post(event.)

		}
	case pb.Message_ATTEND:
		{
			log.Debug("Message_ATTEND############")
			//新节点全部连接上之后通知
			node.higherEventManager.Post(event.NewPeerEvent{
				Payload: msg.Payload,
			})
			//response
			//response.MessageType = pb.Message_ATTEND_RESPNSE
		}
	case pb.Message_ATTEND_RESPNSE:
		{
			// here need to judge if update
			//if primary
			if node.IsPrimary {
				node.attendChan <- 2
			}else{
				node.attendChan <- 1
			}

		}
	case pb.Message_CONSUS:
		{
			log.Debug("<<<< GOT A CONSUS MESSAGE >>>>")

			log.Debug("**** Node Decode MSG ****")
			log.Debug("Node need to decode msg: ", hex.EncodeToString(msg.Payload))
			transferData,err := node.TEM.DecWithSecret(msg.Payload, msg.From.Hash)
			if err != nil{
				log.Error("cannot decode the message",err)
				return nil,err
			}
			//log.Debug("Node解密后信息", hex.EncodeToString(transferData))
			//log.Debug("Node解密后信息2", string(transferData))
			response.Payload = []byte("GOT_A_CONSENSUS_MESSAGE")
			if string(transferData) == "TEST" {
				response.Payload = []byte("GOT_A_TEST_CONSENSUS_MESSAGE")
			}
			log.Debug("from Node", msg.From.ID)
			log.Debug(hex.EncodeToString(transferData))

			go node.higherEventManager.Post(
				event.ConsensusEvent{
					Payload: transferData,
				})
		}
	case pb.Message_SYNCMSG:
		{
			// package the response msg
			response.MessageType = pb.Message_RESPONSE
			transferData,err := node.TEM.DecWithSecret(msg.Payload, msg.From.Hash)
			if err != nil{
				log.Error("cannot decode the message",err);
				return nil,err
			}
			response.Payload = []byte("got a sync msg")
			log.Debug("<<<< GOT A Unicast MESSAGE >>>>")
			var SyncMsg recovery.Message
			unMarshalErr := proto.Unmarshal(transferData, &SyncMsg)
			if unMarshalErr != nil {
				response.Payload = []byte("Sync message Unmarshal error")
				log.Error("sync UnMarshal error!")
			}
			switch SyncMsg.MessageType {
			case recovery.Message_SYNCBLOCK:
				{

					go node.higherEventManager.Post(event.ReceiveSyncBlockEvent{
						Payload: SyncMsg.Payload,
					})

				}
			case recovery.Message_SYNCCHECKPOINT:
				{
					go node.higherEventManager.Post(event.StateUpdateEvent{
						Payload: SyncMsg.Payload,
					})

				}
			case recovery.Message_SYNCSINGLE:
				{
					go node.higherEventManager.Post(event.StateUpdateEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_RELAYTX:
				{
					//log.Warning("Message_RELAYTX: ")
					go node.higherEventManager.Post(event.ConsensusEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_INVALIDRESP:
				{
					go node.higherEventManager.Post(event.RespInvalidTxsEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_SYNCREPLICA:
				{
					go node.higherEventManager.Post(event.ReplicaStatusEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_BROADCAST_NEWPEER:
				{
					log.Debug("receive Message_BROADCAST_NEWPEER")
					go node.higherEventManager.Post(event.RecvNewPeerEvent{
						Payload:SyncMsg.Payload,
					})
				}
			case recovery.Message_BROADCAST_DELPEER:
				{
					log.Debug("receive Message_BROADCAST_DELPEER")
					go node.higherEventManager.Post(event.RecvDelPeerEvent{
						Payload:SyncMsg.Payload,
					})
				}
			}
		}
	case pb.Message_KEEPALIVE:
		{
			//客户端会发来keepAlive请求,返回response即可
			// client may send a keep alive request, just response A response type message,if node is not ready, send a pending status message
			response.MessageType = pb.Message_RESPONSE
			response.Payload = []byte("RESPONSE FROM SERVER")
		}
	case pb.Message_RESPONSE:
		{
			// client couldn't send a response message to server, so server should never receive a response type message
			log.Warning("Client Send a Response Message to Server, this is not allowed!")

		}
	case pb.Message_PENDING:
		{
			log.Warning("Got a PADDING Message")
		}
	default:
		log.Warning(msg.MessageType)
		log.Warning("Unkown Message type!")
	}
	if msg.MessageType != pb.Message_HELLO && msg.MessageType != pb.Message_HELLO_RESPONSE && msg.MessageType != pb.Message_RECONNECT_RESPONSE && msg.MessageType != pb.Message_RECONNECT {
		var err error

		response.Payload,err = node.TEM.EncWithSecret(response.Payload, msg.From.Hash)
		if err != nil {
			log.Error("encode error",err)
		}

	}
	return &response, nil
}

// StartServer start the gRPC server
func (node *Node) StartServer() {
	log.Info("Starting the grpc listening server...")
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(node.localAddr.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		log.Fatal("PLEASE RESTART THE SERVER NODE!")
	}
	opts := membersrvc.GetGrpcServerOpts()
	node.gRPCServer = grpc.NewServer(opts...)
	//this.gRPCServer = grpc.NewServer()
	pb.RegisterChatServer(node.gRPCServer, node)
	log.Info("Listening gRPC request...")
	go node.gRPCServer.Serve(lis)
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (node *Node) StopServer() {
	node.gRPCServer.GracefulStop()

}

func (node *Node)reconnect(msg *pb.Message) {
	opts := membersrvc.GetGrpcClientOpts()
	conn, err := grpc.Dial(msg.From.IP + ":" + strconv.Itoa(int(msg.From.Port)), opts...)
	//conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:", err)
	}
	Client := pb.NewChatClient(conn)
	peer := node.PeersPool.GetPeerByHash(msg.From.Hash)
	if  peer == nil {
		log.Warning("This remote Node already existed, and try to reconnect...")
	}

	node.PeersPool.SetClientByHash(msg.From.Hash,Client)
	node.PeersPool.SetConnectionByHash(msg.From.Hash,conn)

	//esatblish the connect and regenerate the secrate
	helloMessage := pb.Message{
		MessageType:  pb.Message_RECONNECT_RESPONSE,
		Payload:      node.TEM.GetLocalPublicKey(),
		From:         node.localAddr.ToPeerAddress(),
		MsgTimeStamp: time.Now().UnixNano(),
	}

	retMessage, err := peer.Client.Chat(context.Background(), &helloMessage)

	if err != nil {
		log.Error("cannot establish a connection", err)
	} else {
		//review 取得对方的秘钥
		if retMessage.MessageType == pb.Message_HELLO_RESPONSE {
			remotePublicKey := retMessage.Payload
			genErr := node.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("genErr", err)
			}

			if err != nil {
				log.Error("cannot decrypt the nodeidinfo!")
				errors.New("Decrypt ERROR")
			}
			log.Critical("reconnect to reverse ID :", msg.From.ID, node.TEM.GetSecret(msg.From.Hash))
			log.Critical(node.TEM.GetSecret(msg.From.Hash))
		}
	}
}
