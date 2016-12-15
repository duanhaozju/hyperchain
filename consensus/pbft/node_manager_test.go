//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"testing"

	"hyperchain/protos"
	"hyperchain/event"
	"hyperchain/core"
	"hyperchain/consensus/helper"
	"encoding/json"
	"hyperchain/common"
	"hyperchain/core/state"
)

func TestRecvLocalNewNode(t *testing.T) {

	pbft := new(pbftProtocal)

	pbft.isNewNode = true
	msg := &protos.NewNodeMessage{}
	err := pbft.recvLocalNewNode(msg)
	if err.Error() != "New replica received duplicate local newNode message" {
		t.Error("Fail to reject duplicate local newNode message")
	}

	pbft.isNewNode = false
	err = pbft.recvLocalNewNode(msg)
	if err.Error() != "New replica received nil local newNode message" {
		t.Error("Fail to reject nil local newNode message")
	}

	msg = &protos.NewNodeMessage{Payload: []byte("test")}
	err = pbft.recvLocalNewNode(msg)
	if err != nil {
		t.Error("Fail to handle valid local newNode message")
	}

}

func TestRecvLocalAddNode(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	pbft.isNewNode = true
	msg := &protos.AddNodeMessage{}
	err := pbft.recvLocalAddNode(msg)
	if err.Error() != "New replica received local addNode message" {
		t.Error("Fail to handle the case if new replica receive addNode message")
	}

	pbft.isNewNode = false
	err = pbft.recvLocalAddNode(msg)
	if err.Error() != "New replica received nil local addNode message" {
		t.Error("Fail to reject nil local addNode message")
	}

	msg = &protos.AddNodeMessage{Payload: []byte("test")}
	err = pbft.recvLocalAddNode(msg)
	if err != nil {
		t.Error("Fail to handle valid local addNode message")
	}

}

func TestRecvLocalDelNode(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	msg := &protos.DelNodeMessage{}
	err := pbft.recvLocalDelNode(msg)
	if err.Error() != "Deleting is not supported as there're only 4 nodes" {
		t.Error("Fail to reject delete message when there're only 4 nodes")
	}

	pbft.N = 5
	err = pbft.recvLocalDelNode(msg)
	if err.Error() != "New replica received invalid local delNode message" {
		t.Error("Fail to reject invalid local delNode message")
	}

	msg = &protos.DelNodeMessage{
		DelPayload: []byte("del"),
		RouterHash: "routerhash",
		Id: uint64(2),
	}
	err = pbft.recvLocalDelNode(msg)
	if err != nil {
		t.Error("Fail to handle valid local delNode message")
	}
}

func TestSendAgreeAddNode(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	pbft.isNewNode = true
	key := "key"

	pbft.sendAgreeAddNode(key)
	add := &AddNode{
		ReplicaId:	pbft.id,
		Key:		key,
	}
	cert := pbft.getAddNodeCert(key)
	ok := cert.addNodes[*add]
	logger.Error(ok)
	if ok {
		t.Error("Fail to reject addnode message as new node")
	}

}

func TestRecvAgreeAddOrDelNode(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	add := &AddNode{
		ReplicaId: uint64(1),
		Key: "key",
	}
	addCert := pbft.getAddNodeCert("key")
	addCert.addNodes[*add] = true
	err := pbft.recvAgreeAddNode(add)
	if err.Error() != "Receive duplicate addnode message" {
		t.Error("Fail to reject duplicate addnode message")
	}

	del := &DelNode{
		ReplicaId: uint64(1),
		Key: "key",
		RouterHash: "routerhash",
	}
	delCert := pbft.getDelNodeCert("key", "routerhash")
	delCert.delNodes[*del] = true
	err = pbft.recvAgreeDelNode(del)
	if err.Error() != "Receive duplicate delnode message" {
		t.Error("Fail to reject duplicate delnode message")
	}

}


type (
	createObjectChange struct {
		account *common.Address
	}

	resetObjectChange struct {
		prev *state.StateObject
	}
)

func (o createObjectChange) Marshal() ([]byte, error) {
	return json.Marshal(o)
}

func TestMarshal(t *testing.T) {

	origin := []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89")
	add := common.BytesToAddress(origin)
	object := createObjectChange{
		account: &add,
	}
	res, err := json.Marshal(object)
	t.Error("res: ", string(res), "err: ", err)
	t.Error("object: ", object)
	data := createObjectChange{}
	err = json.Unmarshal(res, &data)
	t.Error("data: ", data, "err: ", err)

}

func TestJson(t *testing.T) {

	type Response1 struct {
		Page   int
		Fruits []string
	}
	type Response2 struct {
		Page   int      `json:"page"`
		Fruits []string `json:"fruits"`
	}

	res1D := &Response1{
		Page:   1,
		Fruits: []string{"apple", "peach", "pear"}}
	res1B, _ := json.Marshal(res1D)
	t.Error(string(res1B))

	res := Response1{}
	json.Unmarshal(res1B, &res)
	t.Error(res)

}
