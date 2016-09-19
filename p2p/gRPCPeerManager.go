// author: chenquan
// date: 16-9-19
// last modified: 16-9-19 20:46
// last Modified Author: chenquan
// change log: 
//		
package p2p

import (
	node "hyperchain/p2p/node"
	peer "hyperchain/p2p/peer"
	"hyperchain/p2p/peerComm"
	"hyperchain/p2p/peerEventManager"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/recovery"
	"strconv"
	"time"
	"hyperchain/p2p/peerPool"
	"hyperchain/event"
	"hyperchain/crypto"
	"github.com/golang/protobuf/proto"
	"hyperchain/p2p/transport"
	"golang.org/x/net/context"
	"encoding/hex"
)


// MAX PEER NUMBER
// var MAXPEERNODE int
// 3DES Secret Key
var DESKEY = []byte("sfe023f_sefiel#fi32lf3e!")

// gRPC peer manager struct, which to manage the gRPC peers
type GrpcPeerManager struct {
	EventManager *peerEventManager.PeerEventManager
	//localNodeHash
	LocalNode    *node.Node
	AliveChain   *chan bool
	peersPool    *peerPool.PeersPool
}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(path string, NodeId int, aliveChan chan bool,isTest bool,eventMux *event.TypeMux) {
	// Read the config
	configs := peerComm.GetConfig(path)
	port, _ := strconv.Atoi(configs["port"+strconv.Itoa(NodeId)])
	MAXPEERNODE,_ := strconv.Atoi(configs["MAXPEERS"])


	// start local node
	this.LocalNode = node.NewNode(port,isTest,eventMux,NodeId)
	//节点处理完成通道
	this.AliveChain = &aliveChan

	//初始化peerEvent处理器
	//this.EventManager = peerEventManager.NewPeerEventManager()
	//注册事件
	//this.EventManager.RegisterEvent(pb.Message_HELLO, peerEventHandler.NewHelloHandler())
	//this.EventManager.RegisterEvent(pb.Message_RESPONSE, peerEventHandler.NewResponseHandler())
	//this.EventManager.RegisterEvent(pb.Message_CONSUS, peerEventHandler.NewBroadCastHandler())
	//this.EventManager.RegisterEvent(pb.Message_KEEPALIVE, peerEventHandler.NewKeepAliveHandler())
	//开启事件处理监听循环
	//this.EventManager.Start()

	// connect to peer
	// 如果进行单元测试,需要将参数设置为true
	// 重构peerpool 不采用单例模式进行管理 TODO
	this.peersPool = peerPool.NewPeerPool(isTest,!isTest)
	// 读取待连接的节点信息
	this.connectToPeers(MAXPEERNODE,NodeId,configs)
	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  │")
	log.Notice("└────────────────────────────┘")

	*this.AliveChain <- true
}

func (this *GrpcPeerManager)connectToPeers(MaxPeerNumber int,NodeId int,configs map[string]string){
	alivePeerMap := make(map[int]bool)
	for i := 1;i<=MaxPeerNumber;i++{
		if i == NodeId{
			alivePeerMap[i] = true
		}else{
			alivePeerMap[i] = false
		}
	}

	log.Notice(alivePeerMap)

	// connect other peers
	for this.peersPool.GetAliveNodeNum() < MaxPeerNumber - 1{
		log.Debug("node:",NodeId,"process connecting task...")
		log.Debug("nodes number:",this.peersPool.GetAliveNodeNum())

		nid := 1
		for range time.Tick(200 * time.Millisecond) {
			//log.Println("status map", nid, status)
			if nid > MaxPeerNumber{
				break
			}
			if alivePeerMap[nid] {
				nid++
				continue
			}
			//if this node is not online, connect it
			peerIp   := configs["node"+strconv.Itoa(nid)]
			peerPort_s := configs["port"+strconv.Itoa(nid)]
			peerAddress := extactAddress(peerIp,peerPort_s)

			peer,peerErr := this.connectToPeer(peerAddress.Ip,peerAddress.Port)

			if peerErr != nil {
				// cannot connect to other peer
				log.Error("Node: ",peerAddress.Ip,":",peerAddress.Port," can not connect!\n", peerErr)
			} else {
				// add  peer to peer pool
				this.peersPool.PutPeer(*peerAddress, peer)
				alivePeerMap[nid] = true
				log.Debug("Peer Node hash:", peerAddress.Hash,"has connected!")
			}
		}
	}
}


func (this *GrpcPeerManager)connectToPeer(peerIp string,peerPort int32)(*peer.Peer,error){
	//if this node is not online, connect it
	peer, peerErr := peer.NewPeerByIpAndPort(peerIp,peerPort)
	if peerErr != nil {
		// cannot connect to other peer
		log.Error("Node: " + peerIp + ":" + strconv.Itoa(int(peerPort)) + " can not connect!\n", peerErr)
		return nil,peerErr
	} else {
		return peer,nil
	}

}



// GetAllPeers get all connected peer in the peer pool
func (this *GrpcPeerManager) GetAllPeers() []*peer.Peer {
	peerPool := peerPool.NewPeerPool(false,false)
	return peerPool.GetPeers()
}

// BroadcastPeers Broadcast Massage to connected peers
func (this *GrpcPeerManager) BroadcastPeers(payLoad []byte) {
	result, err := transport.TripleDesEncrypt(payLoad, DESKEY)
	if err!=nil{
		log.Fatal("TripleDesEncrypt Failed!")
	}
	localNodeAddr := node.GetNodeAddr()
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         &localNodeAddr,
		Payload:      result,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	pPool := peerPool.NewPeerPool(false, false)
	//go this.EventManager.PostEvent(pb.Message_CONSUS, broadCastMessage)
	go broadcast(broadCastMessage,pPool)
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(broadCastMessage pb.Message,pPool *peerPool.PeersPool){
	for _, peer := range pPool.GetPeers() {
		go peer.Chat(&broadCastMessage)
		//go func(){
		//	resMsg, err := peer.Chat(&broadCastMessage)
		//	if err != nil {
		//		log.Error("Broadcast failed,Node", peer.Addr)
		//	} else {
		//		log.Debug("resMsg:", string(resMsg.Payload))
		//		//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
		//	}
		//}()
	}
}


// SendMsgToPeers Send msg to specific peer peerlist
func (this *GrpcPeerManager) SendMsgToPeers(payLoad []byte,peerList []uint64,MessageType recovery.Message_MsgType){
	var mpPaylod = &recovery.Message{
		MessageType:MessageType,
		MsgTimeStamp:time.Now().UnixNano(),
		Payload:payLoad,
	}
	realPayload, err := proto.Marshal(mpPaylod)
	if err != nil{
		log.Error("marshal failed")
	}
	result, err := transport.TripleDesEncrypt(realPayload, DESKEY)
	if err!=nil{
		log.Fatal("TripleDesEncrypt Failed!")
	}
	localNodeAddr := node.GetNodeAddr()
	var syncMessage = pb.Message{
		MessageType:  pb.Message_SYNCMSG,
		From:         &localNodeAddr,
		Payload:      result,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	pPool := peerPool.NewPeerPool(false, false)


	// broadcast to special peers
	//TODO for stateUpdate
	go func(){for _, peer := range pPool.GetPeers() {

		for _,nodeID := range peerList{
			nid:=strconv.FormatUint(nodeID,10)
			//peerId:=uint64(strconv.Atoi(peer.Idetity))

			//nid := strconv.Itoa(nodeID)

			//if peerId==nodeID{
			if peer.Idetity == nid {
				log.Error(nid)
				resMsg, err := peer.Chat(&syncMessage)
				if err != nil {
					log.Error("enter error")
					log.Error("Broadcast failed,Node", peer.Addr)
				} else {
					log.Info("resMsg:", string(resMsg.Payload))
					//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
				}
			}
		}

	}
	}()


}


func (this *GrpcPeerManager) GetPeerInfos() peer.PeerInfos{
	peerpool := peerPool.NewPeerPool(false,false);
	peers := peerpool.GetPeers()
	var perinfo peer.PeerInfo
	localNodeAddr := node.GetNodeAddr()
	result, err := transport.TripleDesEncrypt([]byte("Query Status"), DESKEY)
	if err!=nil{
		log.Fatal("TripleDesEncrypt Failed!")
	}

	var keepAliveMessage = pb.Message{
		MessageType:  pb.Message_KEEPALIVE,
		From:         &localNodeAddr,
		Payload:      result,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	var perinfos peer.PeerInfos
	for _,per := range peers{
		log.Debug("rage the peer")
		perinfo.IP = per.Addr.Ip
		perinfo.Port = int(per.Addr.Port)
		perinfo.CName = per.CName
		retMsg, err := per.Client.Chat(context.Background(),&keepAliveMessage)
		if err != nil{
			perinfo.Status = peer.STOP
		}else if retMsg.MessageType == pb.Message_RESPONSE{
			perinfo.Status = peer.ALIVE
		}else if retMsg.MessageType == pb.Message_PENDING{
			perinfo.Status = peer.PENDING
		}
		perinfos = append(perinfos,&perinfo)
	}
	return perinfos
}

// GetNodeId GetLocalNodeIdHash string
func (this *GrpcPeerManager) GetNodeId() string{
	addr := node.GetNodeAddr()
	return getHash(addr.String())
}

func getHash(needHashString string)string {
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	return hex.EncodeToString(hasher.Hash(needHashString).Bytes())
}

func extactAddress(peerIp string, peerPort_s string) *pb.PeerAddress{
	peerPort_i,err := strconv.Atoi(peerPort_s)
	if err != nil{
		log.Error("port cannot convert into int",err)
	}
	peerPort_i32 := int32(peerPort_i)
	peerAddrString := peerIp + ":" + peerPort_s
	peerAddress := pb.PeerAddress{
		Ip:peerIp,
		Port:peerPort_i32,
		Address:peerAddrString,
		Hash:getHash(peerAddrString),
	}
	return &peerAddress
}

