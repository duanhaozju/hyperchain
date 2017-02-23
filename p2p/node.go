//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/core/crypto/primitives"
	"hyperchain/event"
	"hyperchain/admittance"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"hyperchain/p2p/transport/ecdh"
	"hyperchain/recovery"
	"net"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	localAddr          *pb.PeerAddr
	gRPCServer         *grpc.Server
	higherEventManager *event.TypeMux
	//common information
	IsPrimary       bool
	delayTable      map[int]int64
	delayTableMutex sync.RWMutex
	DelayChan       chan UpdateTable
	attendChan      chan int
	PeersPool       PeersPool
	N               int
	DelayTableMutex sync.Mutex
	TEM             transport.TransportEncryptManager
	CM              *admittance.CAManager
}

type UpdateTable struct {
	updateID   int
	updateTime int64
}

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(localAddr *pb.PeerAddr, hEventManager *event.TypeMux, TEM transport.TransportEncryptManager, peersPool PeersPool, cm *admittance.CAManager) *Node {
	var newNode Node
	newNode.localAddr = localAddr
	newNode.TEM = TEM
	newNode.CM = cm
	newNode.higherEventManager = hEventManager
	newNode.PeersPool = peersPool
	newNode.attendChan = make(chan int,1000)
	newNode.delayTable = make(map[int]int64)
	newNode.DelayChan = make(chan UpdateTable)
	//listen the update
	go newNode.UpdateDelayTableThread()

	log.Debug("NODE START...")
	log.Debugf("LOCAL NODE INFO:\nID: %d\nIP: %s\nPORT: %d\nHASH: %s", localAddr.ID,localAddr.IP,localAddr.Port,localAddr.Hash)
	return &newNode
}

//监听节点状态更新线程
func (node *Node) UpdateDelayTableThread() {
	for v := range node.DelayChan {
		if v.updateID > 0 {
			node.delayTableMutex.Lock()
			node.delayTable[v.updateID] = v.updateTime
			node.delayTableMutex.Unlock()
		}

	}
}

//新节点需要监听相应的attend类型
func (n *Node) attendNoticeProcess(N int) {
	log.Critical("AttendProcess N:",N)
	// fix the N as N-1
	// temp
	isPrimaryConnectFlag := false
	f := (N - 1) / 3
	num := 0
	for {
		flag := <-n.attendChan
		log.Debug("attend flag: ", flag, " num: ", num)
		switch flag {
		case 1: {
			num++
			}
		case 2:{
			isPrimaryConnectFlag =true
			num++
			}
		}
		if num >= (N-f) && isPrimaryConnectFlag{
			log.Debug("new node has online ,post already in chain event")
			go n.higherEventManager.Post(event.AlreadyInChainEvent{})
		}

		if num == N {
			break
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
	if msg.MessageType != pb.Message_CONSUS {
		log.Debugf("\n###########################\nSTART OF NEW MESSAGE")
		log.Debugf("LOCAL=> ID:%d IP:%s PORT:%d", node.localAddr.ID, node.localAddr.IP, node.localAddr.Port)
		log.Debugf("MSG FORM=> ID: %d IP: %s PORT: %d", msg.From.ID, msg.From.IP, msg.From.Port)
		log.Debugf("MSG TYPE: %v, from: %d", msg.MessageType, msg.From.ID)
		defer log.Debugf("END OF NEW MESSAGE\n###########################\n")
	}
	response := new(pb.Message)
	response.MessageType = pb.Message_RESPONSE
	response.MsgTimeStamp = time.Now().UnixNano()
	response.From = node.localAddr.ToPeerAddress()

	if (msg.MessageType != pb.Message_HELLO && msg.MessageType != pb.Message_INTRODUCE && msg.MessageType != pb.Message_ATTEND && msg.MessageType != pb.Message_ATTEND_NOTIFY && msg.MessageType == pb.Message_RECONNECT && msg.MessageType == pb.Message_RECONNECT_RESPONSE) && node.CM.GetIsUsed() == true {
		/**
		 * 这里的验证存在问题，首先是通过from取得公钥，这里的from是可以随意伪造的，存在风险
		 * 在 != hello 的时候，没有对证书进行验证，是出于效率的考虑，但这存在安全隐患
		 */
		//验签
		signPubByte := node.TEM.GetSignPublicKey(msg.From.Hash)
		if signPubByte == nil {

			log.Error("cannot get signPublicKey message type is ",msg.MessageType,msg.From.ID)
			return response, errors.New("SignPublickey is missing")
		}
		ecdh256 := ecdh.NewEllipticECDH(elliptic.P256())
		signPub, _ := ecdh256.Unmarshal(signPubByte)
		ecdsaEncrypto := primitives.NewEcdsaEncrypto("ecdsa")
		if signPub == nil {
			log.Warning("signPub is nil")
		}
		if msg.Payload == nil {
			log.Warning("msg.payload is nil")
		}
		if msg.Signature == nil {
			return response, errors.New("signature is nil!!")
		}
		bol, err := ecdsaEncrypto.VerifySign(signPub, msg.Payload, msg.Signature.Signature)
		if !bol || err != nil {
			log.Error("cannot verified the ecert signature", bol)
			return response, errors.New("signature is wrong!!")
		}
		log.Debug("CERT SIGNATURE VERIFY PASS")
		// review 用CM对验证进行管理(此处的必要性需要考虑)
		// TODO 1. 验证ECERT的合法性
		//bol1,err := node.CM.VerifyECert()

		// review 2. 验证传输消息签名的合法性
		// 参数1. PEM 证书 string
		// 参数2. signature
		// 参数3. 原始数据
		//bol2,err := node.CM.VerifySignature(certPEM,signature,signed)

	}

	//handle the message
	switch msg.MessageType {
	case pb.Message_HELLO:
		{
			ecertByte := msg.Signature.ECert
			bol, err := node.CM.VerifyCertSignature(string(ecertByte), msg.Payload, msg.Signature.Signature)
			if !bol || err != nil {
				log.Error("Verify the cert signature failed!",err)
				return response, errors.New("Verify the cert signature failed!")
			}
			log.Debug("CERT SIGNATURE VERIFY PASS")
			// TODO 这里不需要parse,修改VErycERTSignature方法
			ecert, err := primitives.ParseCertificate(string(ecertByte))
			if err != nil {
				log.Error("cannot parse certificate", bol)
				return response, errors.New("signature is wrong!!")
			}
			//再验证证书合法性
			verifyEcert, ecertErr := node.CM.VerifyECert(string(ecertByte))
			if !verifyEcert || ecertErr != nil {
				log.Error(ecertErr)
				return response, ecertErr
			}
			log.Debug("ECERT VERIFY PASS")

			response.MessageType = pb.Message_HELLO_RESPONSE
			//review 协商密钥
			signpub := ecert.PublicKey.(*(ecdsa.PublicKey))
			ecdh256 := ecdh.NewEllipticECDH(elliptic.P256())
			signpubbyte := ecdh256.Marshal(*signpub)
			if node.TEM.GetSecret(msg.From.Hash) == ""{
				genErr := node.TEM.GenerateSecret(signpubbyte, msg.From.Hash)
				if genErr != nil {
					log.Errorf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID,genErr)
				}
			}

			//every times get the public key is same
			response.Payload = node.TEM.GetLocalPublicKey()
			SignCert(response,node.CM)
			return response, nil
		}
	case pb.Message_HELLO_RESPONSE:
		{
			log.Warning("Invalidate HELLO_RESPONSE message")
		}

	case pb.Message_HELLOREVERSE:
		{
			ecertByte := msg.Signature.ECert
			bol, err := node.CM.VerifyCertSignature(string(ecertByte), msg.Payload, msg.Signature.Signature)
			if !bol || err != nil {
				log.Error("Verify the cert signature failed!",err)
				return response, errors.New("Verify the cert signature failed!")
			}
			log.Debug("CERT SIGNATURE VERIFY PASS")
			// TODO 这里不需要parse,修改VErycERTSignature方法
			ecert, err := primitives.ParseCertificate(string(ecertByte))
			if err != nil {
				log.Error("cannot parse certificate", bol)
				return response, errors.New("signature is wrong!!")
			}
			//再验证证书合法性
			verifyEcert, ecertErr := node.CM.VerifyECert(string(ecertByte))
			if !verifyEcert || ecertErr != nil {
				log.Error(ecertErr)
				return response, ecertErr
			}
			log.Debug("ECERT VERIFY PASS")

			response.MessageType = pb.Message_HELLOREVERSE_RESPONSE
			//review 协商密钥
			signpub := ecert.PublicKey.(*(ecdsa.PublicKey))
			ecdh256 := ecdh.NewEllipticECDH(elliptic.P256())
			signpubbyte := ecdh256.Marshal(*signpub)
			if node.TEM.GetSecret(msg.From.Hash) == ""{
				genErr := node.TEM.GenerateSecret(signpubbyte, msg.From.Hash)
				if genErr != nil {
					log.Errorf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID,genErr)
				}
			}
			//every times get the public key is same
			response.Payload = node.TEM.GetLocalPublicKey()
			SignCert(response,node.CM)
			return response, nil
		}
	case pb.Message_HELLOREVERSE_RESPONSE:
		{
			log.Warning(" Message REVERSE")

		}
	case pb.Message_RECONNECT:
		{
			ecertByte := msg.Signature.ECert
			bol, err := node.CM.VerifyCertSignature(string(ecertByte), msg.Payload, msg.Signature.Signature)
			if !bol || err != nil {
				log.Error("Verify the cert signature failed!",err)
				return response, errors.New("Verify the cert signature failed!")
			}
			log.Debug("CERT SIGNATURE VERIFY PASS")
			// TODO 这里不需要parse,修改VErycERTSignature方法
			ecert, err := primitives.ParseCertificate(string(ecertByte))
			if err != nil {
				log.Error("cannot parse certificate", bol)
				return response, errors.New("signature is wrong!!")
			}
			//再验证证书合法性
			verifyEcert, ecertErr := node.CM.VerifyECert(string(ecertByte))
			if !verifyEcert || ecertErr != nil {
				log.Error(ecertErr)
				return response, ecertErr
			}
			log.Debug("ECERT VERIFY PASS")

			response.MessageType = pb.Message_RECONNECT_RESPONSE
			//review 协商密钥
			signpub := ecert.PublicKey.(*(ecdsa.PublicKey))
			ecdh256 := ecdh.NewEllipticECDH(elliptic.P256())
			signpubbyte := ecdh256.Marshal(*signpub)
			genErr := node.TEM.GenerateSecret(signpubbyte, msg.From.Hash)
			if genErr != nil {
				log.Errorf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID,genErr)
			}
			// judge if reconnect TODO judge the idenfication
			go node.reverseConnect(msg)
			//every times get the public key is same
			response.Payload = node.TEM.GetLocalPublicKey()
			SignCert(response,node.CM)
			return response, nil
			log.Warning(" Message RECONNECT")

		}
	case pb.Message_RECONNECT_RESPONSE:
		{
			log.Warning(" Message RECONNECT")
		}
	case pb.Message_INTRODUCE:
		{
			//TODO 验证签名
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
			//新节点全部连接上之后通知
			go node.higherEventManager.Post(event.NewPeerEvent{
				Payload: msg.Payload,
			})

			ecertByte := msg.Signature.ECert
			bol, err := node.CM.VerifyCertSignature(string(ecertByte), msg.Payload, msg.Signature.Signature)
			if !bol || err != nil {
				log.Error("Verify the cert signature failed!",err)
				return response, errors.New("Verify the cert signature failed!")
			}
			log.Debug("CERT SIGNATURE VERIFY PASS")
			// TODO 这里不需要parse,修改VErycERTSignature方法
			ecert, err := primitives.ParseCertificate(string(ecertByte))
			if err != nil {
				log.Error("cannot parse certificate", bol)
				return response, errors.New("signature is wrong!!")
			}
			//再验证证书合法性
			verifyEcert, ecertErr := node.CM.VerifyECert(string(ecertByte))
			if !verifyEcert || ecertErr != nil {
				log.Error(ecertErr)
				return response, ecertErr
			}
			log.Debug("ECERT VERIFY PASS")
			response.MessageType = pb.Message_ATTEND_RESPONSE
			//review 协商密钥
			signpub := ecert.PublicKey.(*(ecdsa.PublicKey))
			ecdh256 := ecdh.NewEllipticECDH(elliptic.P256())
			signpubbyte := ecdh256.Marshal(*signpub)
			genErr := node.TEM.GenerateSecret(signpubbyte, msg.From.Hash)
			if genErr != nil {
				log.Errorf("generate the share secret key, from node id: %d, error info %s ", msg.From.ID,genErr)
			}
			// judge if reconnect TODO judge the idenfication
			response.Payload = node.TEM.GetLocalPublicKey()
			SignCert(response,node.CM)
			return response, nil

		}
	case pb.Message_ATTEND_RESPONSE:
		{
		}
	case pb.Message_ATTEND_NOTIFY:
		{
		// here need to judge if update
			//if primary
			if string(msg.Payload) == "true" {
				node.attendChan <- 2
			} else {
				node.attendChan <- 1
			}
			ecertByte := msg.Signature.ECert
			bol, err := node.CM.VerifyCertSignature(string(ecertByte), msg.Payload, msg.Signature.Signature)
			if !bol || err != nil {
				log.Error("Verify the cert signature failed!",err)
				return response, errors.New("Verify the cert signature failed!")
			}
			log.Debug("CERT SIGNATURE VERIFY PASS")
			// TODO 这里不需要parse,修改VErycERTSignature方法
			//再验证证书合法性
			verifyEcert, ecertErr := node.CM.VerifyECert(string(ecertByte))
			if !verifyEcert || ecertErr != nil {
				log.Error(ecertErr)
				return response, ecertErr
			}
			log.Debug("ECERT VERIFY PASS")
			response.MessageType = pb.Message_ATTEND_NOTIFY_RESPONSE
			//review 协商密钥
			// judge if reconnect TODO judge the idenfication
			response.Payload = node.TEM.GetLocalPublicKey()
			SignCert(response,node.CM)
			return response, nil

		}
	case pb.Message_ATTEND_NOTIFY_RESPONSE:
		{
			log.Warning("get a message ATTEND NOTIFI RESPONSE MESSAGE")
		}
	case pb.Message_CONSUS:
		{

			transferData, err := node.TEM.DecWithSecret(msg.Payload, msg.From.Hash)

			if err != nil {
				log.Error("cannot decode the message", err)
				return nil, err
			}
			response.Payload = []byte("GOT_A_CONSENSUS_MESSAGE")
			if string(transferData) == "TEST" {
				response.Payload = []byte("GOT_A_TEST_CONSENSUS_MESSAGE")
			}

			go node.higherEventManager.Post(
				event.ConsensusEvent{
					Payload: transferData,
				})
		}
	case pb.Message_SYNCMSG:
		{
			// package the response msg
			response.MessageType = pb.Message_RESPONSE
			transferData, err := node.TEM.DecWithSecret(msg.Payload, msg.From.Hash)
			if err != nil {
				log.Error("cannot decode the message", err)
				return nil, err
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
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_BROADCAST_DELPEER:
				{
					log.Debug("receive Message_BROADCAST_DELPEER")
					go node.higherEventManager.Post(event.RecvDelPeerEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_VERIFIED_BLOCK:
				{
					log.Debug("receive Message_BROADCAST_DELPEER")
					go node.higherEventManager.Post(event.ReceiveVerifiedBlock{
						Payload: SyncMsg.Payload,
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
	if msg.MessageType != pb.Message_HELLO && msg.MessageType != pb.Message_HELLO_RESPONSE && msg.MessageType != pb.Message_RECONNECT_RESPONSE && msg.MessageType != pb.Message_RECONNECT && msg.MessageType != pb.Message_ATTEND && msg.MessageType != pb.Message_ATTEND_NOTIFY && msg.MessageType != pb.Message_INTRODUCE {
		var err error

		response.Payload, err = node.TEM.EncWithSecret(response.Payload, msg.From.Hash)
		if err != nil {
			log.Error("encode error", err)
		}

	}
	return response, nil
}

// StartServer start the gRPC server
func (node *Node) StartServer() {
	log.Info("Starting the grpc listening server...")
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(node.localAddr.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		log.Fatal("PLEASE RESTART THE SERVER NODE!")
	}
	//opts := membersrvc.GetGrpcServerOpts()
	opts := node.CM.GetGrpcServerOpts()
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

func (node *Node) reverseConnect(msg *pb.Message) error {
	peer, err := NewPeerReverse(pb.RecoverPeerAddr(msg.From), node.localAddr, node.TEM,node.CM)
	if err != nil {
		log.Critical("new peer failed")
		return err
	} else {
		node.PeersPool.PutPeer(*pb.RecoverPeerAddr(msg.From), peer)
		return nil
	}
}
