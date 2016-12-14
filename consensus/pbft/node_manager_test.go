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