//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"testing"

	"hyperchain/protos"
	"hyperchain/event"
	"hyperchain/core"
	"hyperchain/consensus/helper"
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

func TestMaybeUpdateTableForAdd(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	cert := pbft.getAddNodeCert("key")
	cert.addCount = 1
	err := pbft.maybeUpdateTableForAdd("key")
	if err != "Not enough add message to update table" {
		t.Error("Fail to reject update table when addcount < 2f + 1")
	}

	cert.addCount = 3
	pbft.inAddingNode = false
	cert.finishAdd = true
	err = pbft.maybeUpdateTableForAdd("key")
	if err != "Replica has already finished adding node" {
		t.Error("Fail to reject others useless add msg")
	}

	cert.addCount = 5
	err = pbft.maybeUpdateTableForAdd("key")
	if err != "Replica has already finished adding node, but still recevice add msg from somewho" {
		t.Error("Fail to reject somewho's addnode msg, something wrong maybe happening")
	}

	cert.addCount = 3
	pbft.inAddingNode = false
	cert.finishAdd = false
	err = pbft.maybeUpdateTableForAdd("key")
	if err != nil {
		t.Error("Fail to handle valid add node msg")
	}
}

func TestMaybeUpdateTableForDel(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	cert := pbft.getDelNodeCert("key", "hash")
	cert.delCount = 1
	err := pbft.maybeUpdateTableForDel("key", "hash")
	if err != "Not enough add message to update table" {
		t.Error("Fail to reject update table when addcount < 2f + 1")
	}

	cert.delCount = 3
	pbft.inAddingNode = false
	cert.finishDel = true
	err = pbft.maybeUpdateTableForDel("key", "hash")
	if err != "Replica has already finished deleting node" {
		t.Error("Fail to reject others useless del msg")
	}

	cert.delCount = 5
	err = pbft.maybeUpdateTableForAdd("key")
	if err != "Replica has already finished adding node, but still recevice add msg from somewho" {
		t.Error("Fail to reject somewho's addnode msg, something wrong maybe happening")
	}

	cert.delCount = 3
	pbft.inAddingNode = false
	cert.finishDel = false
	err = pbft.maybeUpdateTableForAdd("key")
	if err != nil {
		t.Error("Fail to handle valid add node msg")
	}
}