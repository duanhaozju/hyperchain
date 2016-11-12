//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

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
	"hyperchain/membersrvc"
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
var fakeRemoteTem *transport.HandShakeManager
var fakeRemotePublicKey []byte

//var fakeConsusData = "080110d9c3a7a29b8ba9bc141a710a6f08d69497989b8ba9bc1412610a28363230316362303434383936346163353937666166366664663166343732656466326132326238391228303030303030303030303030303030303030303030303030303030303030303030303030303030321a01312081f28b989b8ba9bc1418012001"
var fakeConsusData = "TEST"


func init(){
	mux := event.TypeMux{}
	fakeRemoteTem =transport.NewHandShakeManger()
	fakeRemotePublicKey = fakeRemoteTem.GetLocalPublicKey()

	tem := transport.NewHandShakeManger()
	testNode = NewNode(8001,&mux,1,tem)
	membersrvc.Start("../../config/test/local_membersrvc.yaml",1)
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



func TestNode_Chat(t *testing.T) {

	//pretend a remote node send a message
	FakeRemoteMsg.Payload = fakeRemotePublicKey
	ret,err := testNode.Chat(context.Background(),&FakeRemoteMsg)
	if err != nil{
		t.Error("传输错误",err)
	}
	//t.Log("返回信息")
	//t.Log("返回信息Type: ",ret.MessageType)
	//t.Log("返回信息From: ",ret.From)
	//t.Log("返回信息Time: ",ret.MsgTimeStamp)
	//t.Log("返回信息Payl:",ret.Payload)

	fakeRemoteTem.GenerateSecret(ret.Payload,ret.From.Hash)
	fakeRemoteSecret := fakeRemoteTem.GetSecret(ret.From.Hash)
	localSecret := testNode.TEM.GetSecret(fakerRemoteAddr.Hash)
	assert.Exactly(t,fakeRemoteSecret,localSecret)
}

func TestNode_Chat2(t *testing.T) {

	//testNode.StartServer()
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
	assert.Exactly(t,"474f545f415f544553545f434f4e53454e5355535f4d455353414745",retMsg)
	//testNode.StopServer()
}



