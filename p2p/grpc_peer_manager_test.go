package p2p

import (
	"testing"
	"hyperchain/admittance"
	"github.com/stretchr/testify/mock"
	"net"
	"time"
	"fmt"
	"github.com/terasum/viper"
	"github.com/stretchr/testify/assert"
	"hyperchain/common"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "hyperchain/p2p/peermessage"
	"strconv"
	"hyperchain/manager/event"
	"hyperchain/core/crypto/primitives"
	"crypto/ecdsa"
	"hyperchain/p2p/transport/ecdh"
	"crypto/elliptic"
	"sync"
	"hyperchain/recovery"
	"hyperchain/core"
)

type fakeNode struct{
	id int
	ip string
	port int
	cm *admittance.CAManager
	gRPCServer *grpc.Server
	mock mock.Mock
}
func newFakeNode(id int) *fakeNode{
	fnode := new(fakeNode)
	fnode.ip = "127.0.0.1"
	fnode.id = id

	core.InitDB("test/db.yaml",8001)
	fnode.port = 20000+id

	fakeviper := viper.New()
	fakeviper.Set("global.configs.caconfig","test/caconfig.toml")

	var cmerr error
	testCaManager,cmerr := admittance.GetCaManager(fakeviper)
	if cmerr != nil {
		panic("cannot initliazied the camanager")
	}
	fnode.cm = testCaManager
	return fnode
}

func (node *fakeNode) StartServer(wg *sync.WaitGroup) {
	log.Info("Starting the grpc listening server...","127.0.0.1:"+strconv.Itoa(node.port))
	lis, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(node.port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		log.Fatal("PLEASE RESTART THE SERVER NODE!")
	}
	//opts := membersrvc.GetGrpcServerOpts()
	opts := node.cm.GetGrpcServerOpts2()
	node.gRPCServer = grpc.NewServer(opts...)
	//this.gRPCServer = grpc.NewServer()
	pb.RegisterChatServer(node.gRPCServer, node)
	go node.gRPCServer.Serve(lis)

	wg.Done()
}

func (node *fakeNode) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Debugf("\n###########################\nSTART OF NEW MESSAGE")
	log.Debugf("LOCAL=> ID:%d IP:%s PORT:%d",node.id,node.ip,node.port)
	log.Debugf("MSG FORM=> ID: %d IP: %s PORT: %d",msg.From.ID,msg.From.IP,msg.From.Port)
	log.Debugf("MSG TYPE: %v",msg.MessageType)
	defer log.Debugf("END OF NEW MESSAGE\n###########################\n")

	var response pb.Message
	response.MessageType = pb.Message_RESPONSE
	response.MsgTimeStamp = time.Now().UnixNano()
	response.From = pb.NewPeerAddr(node.ip,node.port,node.port+80,node.id).ToPeerAddress()

	fmt.Println(msg.MessageType)

	//handle the message
	switch msg.MessageType {
	case pb.Message_HELLO:
		{
			contentPri := node.cm.GetECertPrivateKeyByte()
			pri, err := primitives.ParseKey(string(contentPri))
			privateKey := pri.(*ecdsa.PrivateKey)
			if err != nil {
				panic("Parse PrivateKey or Ecert failed,please check the privateKey or Ecert and restart the node!")
			}
			e := ecdh.NewEllipticECDH(elliptic.P256())
			pubkey := e.Marshal((*privateKey).PublicKey)
			response.MessageType = pb.Message_HELLO_RESPONSE
			response.Payload = pubkey
		}
	case pb.Message_HELLO_RESPONSE:
		{

		}
	case pb.Message_RECONNECT:
		{
		}
	case pb.Message_RECONNECT_RESPONSE:
		{
		}
	case pb.Message_INTRODUCE:
		{
		}
	case pb.Message_INTRODUCE_RESPONSE:
		{
		}
	case pb.Message_ATTEND:
		{
		}
	case pb.Message_ATTEND_RESPONSE:
		{
		}
	case pb.Message_CONSUS:
		{
		}
	case pb.Message_SYNCMSG:
		{
		}
	case pb.Message_KEEPALIVE:
		{
		}
	case pb.Message_RESPONSE:
		{
		}
	case pb.Message_PENDING:
		{
		}
	default:
	}
	return &response, nil

}

func TestNewGrpcManager(t *testing.T) {
	config := common.NewConfig("test/global.yaml")
	grpcManager := NewGrpcManager(config)
	assert.True(t,!grpcManager.isOnline)
	assert.Nil(t,grpcManager.cm)
	assert.NotNil(t,grpcManager.configs)
	assert.NotNil(t,grpcManager.introducer)
	assert.True(t,grpcManager.isOriginal)
	assert.True(t,grpcManager.isVP)
	assert.NotNil(t,grpcManager.localAddr)
	assert.Nil(t,grpcManager.pool)


	assert.Equal(t,grpcManager.configs.IsVP(),true)
	assert.Equal(t,grpcManager.configs.IntroID(),1)
	assert.Equal(t,grpcManager.configs.IntroIP(),"127.0.0.1")
	assert.Equal(t,grpcManager.configs.LocalGRPCPort(),8001)

	assert.Equal(t,grpcManager.configs.GetID(1),1)
	assert.Equal(t,grpcManager.configs.GetPort(1),20001)
	assert.Equal(t,grpcManager.configs.GetRPCPort(1),20081)
}
var (
	config *common.Config
	grpcm *GRPCPeerManager
	fakeAliveChain chan int
	fakeEventMux *event.TypeMux
	fakeviper *viper.Viper
	testCaManager *admittance.CAManager
)

func init(){
	config = common.NewConfig("test/global.yaml")
	wg := new(sync.WaitGroup)
	wg.Add(4)
	// start all fake nodes
	fakeNode1 := newFakeNode(1)
	fakeNode1.StartServer(wg)
	fakeNode2 := newFakeNode(2)
	fakeNode2.StartServer(wg)
	fakeNode3 := newFakeNode(3)
	fakeNode3.StartServer(wg)
	fakeNode4 := newFakeNode(4)
	fakeNode4.StartServer(wg)
	wg.Wait()

	fakeAliveChain = make(chan int,1)
	go channelCostomer(fakeAliveChain)
	fakeEventMux = new(event.TypeMux)
	fakeviper = viper.New()
	fakeviper.Set("global.configs.caconfig","test/caconfig.toml")
	var cmerr error
	testCaManager,cmerr = admittance.GetCaManager(fakeviper)
	if cmerr != nil {
		panic("cannot initliazied the camanager")
	}
}
func channelCostomer(alive chan int){
	for v := range alive{
		if v == 0{
			fmt.Println("get a signal")
		}
	}
}

func TestGRPCPeerManager_Start(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	assert.NotNil(t,grpcm.pool)
	assert.NotNil(t,grpcm.cm)
	assert.True(t,grpcm.isOnline)
	assert.True(t,grpcm.isVP)
}

func TestGRPCPeerManager_GetAllPeers(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8801
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)

	allPeers := grpcm.GetAllPeers()
	assert.Equal(t,2,allPeers[0].PeerAddr.ID)
	assert.Equal(t,20002,allPeers[0].PeerAddr.Port)
	assert.Equal(t,20082,allPeers[0].PeerAddr.RPCPort)
	assert.Equal(t,3,allPeers[1].PeerAddr.ID)
	assert.Equal(t,20003,allPeers[1].PeerAddr.Port)
	assert.Equal(t,20083,allPeers[1].PeerAddr.RPCPort)
	assert.Equal(t,4,allPeers[2].PeerAddr.ID)
	assert.Equal(t,20004,allPeers[2].PeerAddr.Port)
	assert.Equal(t,20084,allPeers[2].PeerAddr.RPCPort)

}

func TestGRPCPeerManager_GetLocalNode(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8802
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	localNode := grpcm.GetLocalNode()
	assert.Equal(t,localNode.GetNodeAddr().ID,1)
	assert.Equal(t,localNode.GetNodeAddr().Port,8802)
	assert.Equal(t,localNode.GetNodeAddr().RPCPort,8081)
	assert.Equal(t,localNode.GetNodeAddr().IP,"127.0.0.1")
}

func TestGRPCPeerManager_BroadcastPeers(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8803
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	payload := []byte("NVPPeers")
	grpcm.BroadcastPeers(payload)

}

func TestGRPCPeerManager_BroadcastNVPPeers(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8804
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	payload := []byte("NVPPeers")
	grpcm.BroadcastNVPPeers(payload)
}

func TestGRPCPeerManager_BroadcastVPPeers(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8805
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	payload := []byte("VPPeers")
	grpcm.BroadcastVPPeers(payload)
}

func TestGRPCPeerManager_GetLocalNodeHash(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8806
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	hash := grpcm.GetLocalNodeHash()
	assert.Equal(t,len("d4a4e3ad24ce52eac10497846f8d97e86f7bd2fc8119631b6cde09fa75a2dd3c"),len(hash))
	assert.Equal(t,"c605d50c3ed56902ec31492ed43b238b36526df5d2fd6153c1858051b6635f6e",hash)
}

func TestGRPCPeerManager_GetNodeId(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8807
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	assert.Equal(t,grpcm.GetNodeId(),1)
}

func TestGRPCPeerManager_GetPeerInfo(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8808
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	peerinfo := grpcm.GetPeerInfo()
	assert.Equal(t,4,peerinfo.GetNumber())
}

func TestGRPCPeerManager_GetVPPeers(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8809
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	vp := grpcm.GetVPPeers()
	assert.Equal(t,0,len(vp))
	grpcm.localNode.StopServer()
}

func TestGRPCPeerManager_GetNVPPeers(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8810
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	nvp := grpcm.GetVPPeers()
	assert.Equal(t,0,len(nvp))
}

func TestGRPCPeerManager_SendMsgToPeers(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8811
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	payload := []byte("VPPeers")
	grpcm.SendMsgToPeers(payload,[]uint64{uint64(2),uint64(3),uint64(4)},recovery.Message_SYNCSINGLE)
}

func TestGRPCPeerManager_SetPrimary(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8812
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	assert.NoError(t,grpcm.SetPrimary(uint64(2)))
	grpcm.localNode.StopServer()
}



func TestGRPCPeerManager_GetAllPeersWithTemp(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8814
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	peersWithTemp := grpcm.GetAllPeersWithTemp()
	assert.Equal(t,3,len(peersWithTemp))
}

func TestGRPCPeerManager_DeleteNode(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8815
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	assert.NoError(t,grpcm.DeleteNode("d4a4e3ad24ce52eac10497846f8d97e86f7bd2fc8119631b6cde09fa75a2dd3c"))
	assert.Equal(t,2,len(grpcm.GetAllPeers()))
}

func TestGRPCPeerManager_GetLocalAddressPayload(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8816
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	payload := grpcm.GetLocalAddressPayload()
	assert.NotEmpty(t,payload,"local peer address payload should not be empty")
}

func TestGRPCPeerManager_GetRouterHashifDelete(t *testing.T) {
	grpcm = NewGrpcManager(config)
	grpcm.localAddr.Port = 8817
	grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	hash,id,_ := grpcm.GetRouterHashifDelete("d4a4e3ad24ce52eac10497846f8d97e86f7bd2fc8119631b6cde09fa75a2dd3c")
	assert.NotEmpty(t,hash)
	//if delete the node should be id
	assert.Equal(t,uint64(0x1),id)
}

func TestGRPCPeerManager_UpdateRoutingTable(t *testing.T) {
	//grpcm.Start(fakeAliveChain,fakeEventMux,testCaManager)
	//payload := grpcm.GetLocalAddressPayload()
	//grpcm.UpdateRoutingTable(payload)
}

