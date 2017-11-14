//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package consensusMocks

import (
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
	pb "github.com/hyperchain/hyperchain/manager/protos"

	"github.com/stretchr/testify/mock"
)

type MockHelper struct {
	mock.Mock
	EventChan chan interface{}
}

func (m *MockHelper) InnerBroadcast(msg *pb.Message) error {
	m.EventChan <- msg
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) InnerUnicast(msg *pb.Message, to uint64) error {
	m.EventChan <- msg
	m.EventChan <- to
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) Execute(seqNo uint64, hash string, flag bool, isPrimary bool, time int64) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) UpdateState(myId uint64, height uint64, blockHash []byte, replicas []event.SyncReplica) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) ValidateBatch(digest string, txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) VcReset(seqNo uint64) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) InformPrimary(primary uint64) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) BroadcastAddNode(msg *pb.Message) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) BroadcastDelNode(msg *pb.Message) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) UpdateTable(payload []byte, flag bool) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHelper) SendFilterEvent(informType int, message ...interface{}) error {
	args := m.Called()
	return args.Error(0)
}
