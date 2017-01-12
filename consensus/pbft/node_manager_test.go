//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"testing"

	"hyperchain/protos"
	"hyperchain/event"
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

	initDB()
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

	initDB()
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

	initDB()
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

	initDB()
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

	initDB()
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
	if err.Error() != "Not enough add message to update table" {
		t.Error("Fail to reject update table when addcount < 2f + 1")
	}

	cert.addCount = 3
	pbft.inAddingNode = false
	cert.finishAdd = true
	err = pbft.maybeUpdateTableForAdd("key")
	if err.Error() != "Replica has already finished adding node" {
		t.Error("Fail to reject others useless add msg")
	}

	cert.addCount = 5
	err = pbft.maybeUpdateTableForAdd("key")
	if err.Error() != "Replica has already finished adding node, but still recevice add msg from someone else" {
		t.Error("Fail to reject someone's addnode msg, something wrong maybe happening")
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

	initDB()
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
	if err.Error() != "Not enough del message to update table" {
		t.Error("Fail to reject update table when delcount < 2f + 1")
	}

	cert.delCount = 3
	pbft.inAddingNode = false
	cert.finishDel = true
	err = pbft.maybeUpdateTableForDel("key", "hash")
	if err.Error() != "Replica has already finished deleting node" {
		t.Error("Fail to reject others useless del msg")
	}

	cert.delCount = 4
	err = pbft.maybeUpdateTableForDel("key", "hash")
	if err.Error() != "Replica has already finished deleting node, but still recevice del msg from someone else" {
		t.Error("Fail to reject someone's delnode msg, something wrong maybe happening")
	}

	cert.delCount = 3
	pbft.inAddingNode = false
	cert.finishDel = false
	err = pbft.maybeUpdateTableForDel("key", "hash")
	if err != nil {
		t.Error("Fail to handle valid del node msg")
	}
}

func TestSendReadyForN(t *testing.T) {

	initDB()
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	pbft.isNewNode = false
	err := pbft.sendReadyForN()
	if err.Error() != "Replica is an old node, but try to send redy_for_n" {
		t.Error("Fail to reject send ready_for_n as old node")
	}

	pbft.isNewNode = true
	err = pbft.sendReadyForN()
	if err.Error() != "Rplica doesn't have local key for ready_for_n" {
		t.Error("Fail to reject send ready_for_n as there's no local key")
	}

	pbft.localKey = "key"
	err = pbft.sendReadyForN()
	if err != nil {
		t.Error("Fail to send ready_for_n in normal case")
	}
}

func TestRecvReadyforNforAdd(t *testing.T) {

	initDB()
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	ready := &ReadyForN{
		ReplicaId:	pbft.id,
		Key:		"key",
	}

	pbft.id = 2
	err := pbft.recvReadyforNforAdd(ready)
	if err.Error() != "Replica is not primary but received ready_for_n" {
		t.Error("Fail to reject ready_for_n as not primary")
	}

	pbft.id = 1
	pbft.activeView = false
	err = pbft.recvReadyforNforAdd(ready)
	if err.Error() != "Primary is in view change, reject the ready_for_n message" {
		t.Error("Fail to reject ready_for_n as not in active view")
	}

	pbft.activeView = true
	err = pbft.recvReadyforNforAdd(ready)
	if err.Error() != "Primary has not done with addnode" {
		t.Error("Fail to reject ready_for_n as addnode hasn't finished")
	}

	cert := pbft.getAddNodeCert(ready.Key)
	cert.finishAdd = true
	cert.updateCount = 4
	err = pbft.recvReadyforNforAdd(ready)
	if err != nil {
		t.Error("Fail to handle valid ready_for_n msg")
	}
}

func TestSendUpdateNforDel(t *testing.T) {

	initDB()
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	pbft.activeView = false
	err := pbft.sendUpdateNforDel("key", "hash")
	if err.Error() != "Primary is in view change, choose not send the ready_for_n message" {
		t.Error("Fail to reject sending ready_for_n msg as it's in viewchange")
	}

	pbft.activeView = true
	cert := pbft.getDelNodeCert("key", "hash")
	cert.finishDel = false
	err = pbft.sendUpdateNforDel("key", "hash")
	if err.Error() != "Primary hasn't done with delnode" {
		t.Error("Fail to reject sending ready_for_n msg as it hasn't done with delnode")
	}
}

func TestRecvUpdateN(t *testing.T) {

	initDB()
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	update := &UpdateN{
		ReplicaId:	2,
		Key:		"key",
		N:			5,
		View: 		1,
		Flag:		true,
		SeqNo:		60,
	}

	pbft.activeView = false
	err := pbft.recvUpdateN(update)
	if err.Error() != "Replica reject update_n msg as it's in viewchange" {
		t.Error("Fail to reject update_n msg as it's in viewchange")
	}

	pbft.activeView = true
	err = pbft.recvUpdateN(update)
	if err.Error() != "Replica reject update_n msg as it's not from primary" {
		t.Error("Fail to reject update_n msg as it's not from primary")
	}

	update.ReplicaId = 1
	err = pbft.recvUpdateN(update)
	if err.Error() != "Replica has different idea about n and view" {
		t.Error("Fail to reject update_n, as it has diff idea about n and view")
	}

	update.View = 0
	err = pbft.recvUpdateN(update)
	if err.Error() != "Replica reject not-in-view msg" {
		t.Error("Fail to reject update_n, as seqNo not in watermarks")
	}

	update.SeqNo = 6
	pbft.recvUpdateN(update)
	addCert := pbft.getAddNodeCert("key")
	if addCert.update == nil {
		t.Error("Fail to handle valid update msg")
	}

	update.Flag = false
	update.View = 1
	err = pbft.recvUpdateN(update)
	if err.Error() != "Replica has different idea about n and view" {
		t.Error("Fail to reject update_n, as it has diff idea about n and view")
	}

	update.View = 0
	pbft.N = 6
	update.RouterHash = "hash"
	pbft.recvUpdateN(update)
	delCert := pbft.getDelNodeCert("key", "hash")
	if delCert.update == nil {
		t.Error("Fail to handle valid update msg")
	}

}

func TestRecvAgreeUpdateN(t *testing.T) {

	initDB()
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	agree := &AgreeUpdateN{
		ReplicaId:	2,
		Key:		"key",
		N:			5,
		View:		0,
		Flag:		true,
	}

	pbft.recvAgreeUpdateN(agree)
	addCert := pbft.getAddNodeCert(agree.Key)
	ok := addCert.agrees[*agree]
	if !ok {
		t.Error("Fail to handle valid agree msg about updating n after adding")
	}
	err := pbft.recvAgreeUpdateN(agree)
	if err.Error() != "Replica ignored duplicate agree updateN msg" {
		t.Error("Fail to reject duplicate updateN msg")
	}

	agree.RouterHash = "hash"
	agree.Flag = false
	pbft.recvAgreeUpdateN(agree)
	delCert := pbft.getDelNodeCert(agree.Key, agree.RouterHash)
	ok = delCert.agrees[*agree]
	if !ok {
		t.Error("Fail to handle valid agree msg about updating n after deleting")
	}
	err = pbft.recvAgreeUpdateN(agree)
	if err.Error() != "Replica ignored duplicate agree updateN msg" {
		t.Error("Fail to reject duplicate updateN msg")
	}
}

func TestMaybeUpdateN(t *testing.T) {

	initDB()
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	addCert := pbft.getAddNodeCert("key")
	addCert.updateCount = 3
	err := pbft.maybeUpdateN("key", "", true)
	if err.Error() != "Replica hasn't locally prepared for updating n after adding" {
		t.Error("Fail to reject updating n when it hasn't locally prepared")
	}
	update := &UpdateN{
		ReplicaId:	1,
		Key:		"key",
		N:			5,
		View: 		1,
		Flag:		true,
		SeqNo:		60,
	}
	addCert.update = update
	addCert.finishUpdate = true
	addCert.updateCount = 4
	err = pbft.maybeUpdateN("key", "", true)
	if err.Error() != "Replica has already finished updating n after adding" {
		t.Error("Fail to reject updating n when it had already updated")
	}
	addCert.updateCount = 6
	err = pbft.maybeUpdateN("key", "", true)
	if err.Error() != "Replica has already finished updating n after adding, but still recevice agree msg from someone else" {
		t.Error("Fail to reject updating n when it had already updated but still received msg")
	}

	delCert := pbft.getDelNodeCert("key", "hash")
	delCert.updateCount = 3
	err = pbft.maybeUpdateN("key", "hash", false)
	if err.Error() != "Replica hasn't locally prepared for updating n after deleting" {
		t.Error("Fail to reject updating n when it hasn't locally prepared")
	}
	update.RouterHash = "hash"
	update.Flag = false
	delCert.update = update
	delCert.finishUpdate = true
	delCert.updateCount = 3
	err = pbft.maybeUpdateN("key", "hash", false)
	if err.Error() != "Replica has already finished updating n after deleting" {
		t.Error("Fail to reject updating n when it had already updated")
	}
	delCert.updateCount = 5
	err = pbft.maybeUpdateN("key", "hash", false)
	if err.Error() != "Replica has already finished updating n after deleting, but still recevice agree msg from someone else" {
		t.Error("Fail to reject updating n when it had already updated but still received msg")
	}

	delCert.finishUpdate = false
	delCert.updateCount = 3
	update.ReplicaId = 3
	err = pbft.maybeUpdateN("key", "hash", false)
	if err != nil {
		t.Error("Fail to handle valid try to update n")
	}
}