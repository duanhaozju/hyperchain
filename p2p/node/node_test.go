// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:59
// last Modified Author: chenquan
// change log: add a header comment of this file
//

package Server

import (
	"testing"
	pb "hyperchain/p2p/peermessage"

	"github.com/stretchr/testify/assert"
	"hyperchain/p2p/peerComm"
	"golang.org/x/net/context"
	"time"
	"hyperchain/event"
	"hyperchain/p2p/transport"
	"encoding/hex"
	hypermessage "hyperchain/protos"
	"github.com/golang/protobuf/proto"
	"hyperchain/membersrvc"
	"github.com/hyperledger/fabric/core/config"
)

var testNode *Node
var expectAddr = peerComm.ExtractAddress(peerComm.GetLocalIp(),8001,1)
var fakerRemoteAddr = peerComm.ExtractAddress(peerComm.GetLocalIp(),8002,2)

var FakeRemoteMsg = pb.Message{
	MsgTimeStamp:time.Now().UnixNano(),
	From         : fakerRemoteAddr,
	MessageType  : pb.Message_HELLO,
	Payload      :[]byte("hello"),
}
var fakeRemoteTem =transport.NewHandShakeManger()
var fakeRemotePublicKey =fakeRemoteTem.GetLocalPublicKey()

//var fakeConsusData = "080110d9c3a7a29b8ba9bc141a710a6f08d69497989b8ba9bc1412610a28363230316362303434383936346163353937666166366664663166343732656466326132326238391228303030303030303030303030303030303030303030303030303030303030303030303030303030321a01312081f28b989b8ba9bc1418012001"
var fakeConsusData = "TEST"


func init(){
	mux := event.TypeMux{}
	tem := transport.NewHandShakeManger()
	testNode = NewNode(8001,&mux,1,tem)
}

func TestNode_GetNodeAddr(t *testing.T) {
	addr := testNode.GetNodeAddr()
	assert.Exactly(t,expectAddr,addr)
}

func TestNode_GetNodeHash(t *testing.T) {
	nodehash := testNode.GetNodeHash()
	assert.Exactly(t,expectAddr.Hash,nodehash)
}

func TestNode_GetNodeID(t *testing.T) {
	nodeid := testNode.GetNodeID()
	assert.Exactly(t,uint64(1),nodeid)
}
func TestNode_StartServer(t *testing.T) {
	testNode.StartServer()
}

func TestNode_Chat(t *testing.T) {
	testNode.StartServer()
	//pretend a remote node send a message
	FakeRemoteMsg.Payload = fakeRemotePublicKey
	ret,err := testNode.Chat(context.Background(),&FakeRemoteMsg)
	if err != nil{
		t.Error("传输错误",err)
	}
	t.Log("返回信息")
	t.Log("返回信息Type: ",ret.MessageType)
	t.Log("返回信息From: ",ret.From)
	t.Log("返回信息Time: ",ret.MsgTimeStamp)
	t.Log("返回信息Payl:",ret.Payload)

	fakeRemoteTem.GenerateSecret(ret.Payload,ret.From.Hash)
	fakeRemoteSecret := fakeRemoteTem.GetSecret(ret.From.Hash)
	localSecret := testNode.TEM.GetSecret(fakerRemoteAddr.Hash)
	assert.Exactly(t,fakeRemoteSecret,localSecret)
}

func TestNode_Chat2(t *testing.T) {
	//membersrvc.Start("", 1)
	testNode.StartServer()
	//pretend a remote node send a message

	FakeRemoteMsg.Payload = fakeRemotePublicKey
	ret,err := testNode.Chat(context.Background(),&FakeRemoteMsg)
	if err != nil{
		t.Error("传输错误",err)
	}
	fakeRemoteTem.GenerateSecret(ret.Payload,ret.From.Hash)
	fakeRemoteSecret := fakeRemoteTem.GetSecret(ret.From.Hash)
	localSecret := testNode.TEM.GetSecret(fakerRemoteAddr.Hash)
	assert.Exactly(t,fakeRemoteSecret,localSecret)


	FakeRemoteMsg.MessageType = pb.Message_CONSUS
	transfer:= fakeRemoteTem.EncWithSecret([]byte(fakeConsusData),testNode.address.Hash)

	FakeRemoteMsg.Payload = transfer
	ret2,err := testNode.Chat(context.Background(),&FakeRemoteMsg)
	if err != nil{
		t.Error("传输错误",err)
	}
	//t.Log("返回信息")
	//t.Log("返回信息Type: ",ret2.MessageType)
	//t.Log("返回信息From: ",ret2.From)
	//t.Log("返回信息Time: ",ret2.MsgTimeStamp)
	//t.Log("返回信息Payl:",ret2.Payload)
	retMsg := hex.EncodeToString(fakeRemoteTem.DecWithSecret(ret2.Payload,ret2.From.Hash))
	assert.Exactly(t,"474f5441434f4e53454e5355534d455353414745",retMsg)
}


func TestNode_Chat3(t *testing.T) {
	//协商秘钥
	testNode.StartServer()
	//pretend a remote node send a message

	FakeRemoteMsg.Payload = fakeRemotePublicKey
	ret,err := testNode.Chat(context.Background(),&FakeRemoteMsg)
	if err != nil{
		t.Error("传输错误",err)
	}
	fakeRemoteTem.GenerateSecret(ret.Payload,ret.From.Hash)
	fakeRemoteSecret := fakeRemoteTem.GetSecret(ret.From.Hash)
	localSecret := testNode.TEM.GetSecret(fakerRemoteAddr.Hash)
	assert.Exactly(t,fakeRemoteSecret,localSecret)

	//发送协商信息

	FakeRemoteMsg.MessageType = pb.Message_CONSUS
	//fake consensus message
	consensusMsg := &hypermessage.Message{
		Timestamp:time.Now().UnixNano(),
		Type:hypermessage.Message_CONSENSUS,
		Payload:[]byte("TEST"),
		Id:uint64(2),
	}

	tranferData,err := proto.Marshal(consensusMsg)
	if err !=nil{
		log.Error("marshal err", err)
	}
	log.Notice("marshal之后",hex.EncodeToString(tranferData))


	transfer:= fakeRemoteTem.EncWithSecret([]byte(tranferData),testNode.address.Hash)

	FakeRemoteMsg.Payload = transfer
	ret2,err := testNode.Chat(context.Background(),&FakeRemoteMsg)
	if err != nil{
		t.Error("传输错误",err)
	}
	retMsg := hex.EncodeToString(fakeRemoteTem.DecWithSecret(ret2.Payload,ret2.From.Hash))

	originalData := fakeRemoteTem.DecWithSecret(transfer,ret2.From.Hash)

	assert.Exactly(t,"474f545f415f434f4e53454e5355535f4d4553534147455f31302e38322e3139392e3231385f38303031",retMsg)
	assert.Exactly(t,tranferData,originalData)

	msg := &hypermessage.Message{}
	umerr := proto.Unmarshal(tranferData,msg)
	if umerr!=nil {
		t.Error("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message", err)
	}

	t.Log("返回信息")
	t.Log("返回信息Id: ",msg.Id)
	t.Log("返回信息Timestamp: ",msg.Timestamp)
	t.Log("返回信息Type: ",msg.Type)
	t.Log("返回信息Payl:",msg.Payload)

}


