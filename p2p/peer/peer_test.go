// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:58
// last Modified Author: chenquan
// change log: add a comment of this file function
//

package client

import (
	"testing"
	node "hyperchain/p2p/node"
	"hyperchain/event"
	"hyperchain/p2p/transport"
	"hyperchain/p2p/peerComm"
	"time"
	pb "hyperchain/p2p/peermessage"
	hypermessage "hyperchain/protos"
	"github.com/golang/protobuf/proto"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
)
var fakeNodeTEM = transport.NewHandShakeManger()
var fakeNode = node.NewNode(8001,new(event.TypeMux),1,"test.cn",fakeNodeTEM)
var fakeNodeAddr =peerComm.ExtractAddress(peerComm.GetLocalIp(),8001,1)


var localAddr = peerComm.ExtractAddress(peerComm.GetLocalIp(),8002,2)
var localTEM = transport.NewHandShakeManger()

func init(){
	fakeNode.StartServer()
}

func TestNewPeerByIpAndPort(t *testing.T) {
	peer,err :=NewPeerByIpAndPort(fakeNodeAddr.Ip,fakeNodeAddr.Port,fakeNodeAddr.ID, fakeNodeTEM,localAddr)
	if err != nil{
		t.Error(err)
	}
	t.Log(peer.Addr)
}

func TestPeer_Chat(t *testing.T) {
	peer,err :=NewPeerByIpAndPort(fakeNodeAddr.Ip,fakeNodeAddr.Port,fakeNodeAddr.ID,localTEM,localAddr)
	if err != nil{
		t.Error(err)
	}
	//t.Log(peer.Addr)

	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:        localAddr,
		Payload:     []byte("TEST"),
		MsgTimeStamp: time.Now().UnixNano(),
	}
	ret,err := peer.Chat(broadCastMessage)
	t.Log(string(ret.Payload))
	decrypted := localTEM.DecWithSecret(ret.Payload,peer.Addr.Hash)
	t.Log(string(decrypted))
}

func TestPeer_Chat2(t *testing.T) {
	peer,err :=NewPeerByIpAndPort(fakeNodeAddr.Ip,fakeNodeAddr.Port,fakeNodeAddr.ID,localTEM,localAddr)
	if err != nil{
		t.Error(err)
	}
	//t.Log(peer.Addr)

	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:        localAddr,
		Payload:     []byte("TEST"),
		MsgTimeStamp: time.Now().UnixNano(),
	}
	ret,err := peer.Chat(broadCastMessage)
	t.Log(string(ret.Payload))
	decrypted := localTEM.DecWithSecret(ret.Payload,peer.Addr.Hash)
	t.Log(string(decrypted))
}

func TestPeer_Chat3(t *testing.T) {
	peer,err :=NewPeerByIpAndPort(fakeNodeAddr.Ip,fakeNodeAddr.Port,fakeNodeAddr.ID,localTEM,localAddr)
	if err != nil{
		t.Error(err)
	}
	//t.Log(peer.Addr)

	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:        localAddr,
		Payload:     []byte("TEST"),
		MsgTimeStamp: time.Now().UnixNano(),
	}

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
	log.Critical("marshal之后",hex.EncodeToString(tranferData))

	broadCastMessage.Payload = tranferData
	//传输
	retmsg,err := peer.Chat(broadCastMessage)

	if err != nil{
		t.Error("chat failed", err)
	}

	//retData := localTEM.DecWithSecret(retmsg.Payload,retmsg.From.Hash)
	//retData :=retmsg

	//log.Critical("解密之后",hex.EncodeToString(*retmsg))
	//msg := &hypermessage.Message{}
	//umerr := proto.Unmarshal(retmsg,msg)
	//if umerr!=nil {
	//	t.Error("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message", err)
	//}

	t.Log("返回信息",retmsg)
	//t.Log("返回信息Id: ",msg.Id)
	//t.Log("返回信息Timestamp: ",msg.Timestamp)
	//t.Log("返回信息Type: ",msg.Type)
	//t.Log("返回信息Payl:",msg.Payload)

	assert.Exactly(t,[]byte{0x47, 0x4f, 0x54, 0x5f, 0x41, 0x5f, 0x43, 0x4f, 0x4e, 0x53, 0x45, 0x4e, 0x53, 0x55, 0x53, 0x5f, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45},retmsg.Payload)
}
