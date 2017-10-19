package rbft

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"hyperchain/common"
	"hyperchain/consensus/consensusMocks"
	"hyperchain/consensus/helper/persist"
	"hyperchain/hyperdb/mdb"
	"hyperchain/manager/protos"
	"testing"
	"time"
)

func TestNewNodeMgr(t *testing.T) {
	nodeMgr := newNodeMgr()
	structName, nilElems, err := checkNilElems(nodeMgr)
	if err != nil {
		t.Error(err.Error())
	}
	if nilElems != nil {
		t.Errorf("There exists some nil elements: %v in struct: %s", nilElems, structName)
	}
}

func TestRecvLocalNewNode(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	payload := []byte{'a', 'b', 'c'}
	msg := &protos.NewNodeMessage{
		Payload: payload,
	}

	// test recvLocalNewNode with isNewNode being true.
	rbft.status.activeState(&rbft.status.isNewNode)
	rbft.recvLocalNewNode(msg)
	ast.Equal(true, rbft.status.getState(&rbft.status.isNewNode), "isNewNode should be true.")

	// test recvLocalNewNode with nil payload.
	rbft.status.inActiveState(&rbft.status.isNewNode)
	msg.Payload = nil
	rbft.recvLocalNewNode(msg)
	ast.Equal(false, rbft.status.getState(&rbft.status.isNewNode), "isNewNode should be false.")

	// test recvLocalNewNode with isNewNode being false.
	msg.Payload = payload
	rbft.recvLocalNewNode(msg)
	ast.Equal(true, rbft.status.getState(&rbft.status.isNewNode), "isNewNode should be true.")

}

func TestPrepareAddNode(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("BroadcastAddNode").Return(nil)

	payload := []byte{'a', 'b', 'c'}
	msg := &protos.AddNodeMessage{
		Payload: payload,
	}

	// test recvLocalAddNode with isNewNode being true.
	rbft.status.activeState(&rbft.status.isNewNode)
	rbft.recvLocalAddNode(msg)
	ast.Equal(false, rbft.status.getState(&rbft.status.inAddingNode), "inAddingNode should be false.")

	// test recvLocalAddNode with nil payload.
	rbft.status.inActiveState(&rbft.status.isNewNode)
	msg.Payload = nil
	rbft.recvLocalAddNode(msg)
	ast.Equal(false, rbft.status.getState(&rbft.status.inAddingNode), "inAddingNode should be false.")

	// test recvLocalAddNode with normal msg.
	msg.Payload = payload
	rbft.recvLocalAddNode(msg)
	ast.Equal(true, rbft.status.getState(&rbft.status.inAddingNode), "inAddingNode should be true.")
	rbft.recvLocalAddNode(msg)

}

func TestPrepareDelNode(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("BroadcastDelNode").Return(nil)

	payload := []byte{'a', 'b', 'c'}
	msg := &protos.DelNodeMessage{
		RouterHash: "delete",
		Id:         1,
		Del:        5,
		DelPayload: payload,
	}

	// test recvLocalDelNode with less than 5 nodes in system.
	rbft.N = 4
	rbft.recvLocalDelNode(msg)
	ast.Equal(false, rbft.status.getState(&rbft.status.inDeletingNode), "inDeletingNode should be false.")

	// test recvLocalDelNode with nil payload.
	rbft.N = 5
	msg.DelPayload = nil
	rbft.recvLocalDelNode(msg)
	ast.Equal(false, rbft.status.getState(&rbft.status.inDeletingNode), "inDeletingNode should be false.")

	// test recvLocalAddNode with normal msg.
	msg.DelPayload = payload
	rbft.recvLocalDelNode(msg)
	ast.Equal(true, rbft.status.getState(&rbft.status.inDeletingNode), "inDeletingNode should be true.")
	rbft.recvLocalDelNode(msg)
}

func TestMaybeUpdateTableForAdd(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("UpdateTable").Return(nil)

	key := "test_add_key"
	addNode1 := AddNode{ReplicaId: 1, Key: key}
	addNode2 := AddNode{ReplicaId: 2, Key: key}
	addNode3 := AddNode{ReplicaId: 3, Key: key}
	addNode4 := AddNode{ReplicaId: 4, Key: key}
	rbft.nodeMgr.addNodeCertStore[key] = &addNodeCert{
		addNodes: map[AddNode]bool{
			addNode1: true,
			addNode2: true,
			addNode3: true,
			addNode4: true,
		},
		finishAdd: true,
	}

	rbft.status.activeState(&rbft.status.inAddingNode)
	// test maybeUpdateTableForAdd with isNewNode being true.
	rbft.status.activeState(&rbft.status.isNewNode)
	rbft.maybeUpdateTableForAdd(key)
	ast.Equal(true, rbft.status.getState(&rbft.status.inAddingNode), "inAddingNode should be true.")

	// test maybeUpdateTableForAdd with inAddingNode being false.
	rbft.status.inActiveState(&rbft.status.isNewNode)
	rbft.status.inActiveState(&rbft.status.inAddingNode)
	rbft.maybeUpdateTableForAdd(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inAddingNode), "inAddingNode should be false.")

	// test maybeUpdateTableForAdd with finishAdd being false.
	rbft.nodeMgr.addNodeCertStore[key].finishAdd = false
	rbft.maybeUpdateTableForAdd(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inAddingNode), "inAddingNode should be false.")

	// test maybeUpdateTableForAdd with normal case.
	rbft.status.activeState(&rbft.status.inAddingNode)
	rbft.maybeUpdateTableForAdd(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inAddingNode), "inAddingNode should be false.")
}

func TestMaybeUpdateTableForDel(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	key := "test_delete_key"
	delNode1 := DelNode{ReplicaId: 1, Key: key}
	delNode2 := DelNode{ReplicaId: 2, Key: key}
	delNode3 := DelNode{ReplicaId: 3, Key: key}
	delNode4 := DelNode{ReplicaId: 4, Key: key}
	rbft.nodeMgr.delNodeCertStore[key] = &delNodeCert{
		delNodes: map[DelNode]bool{
			delNode1: true,
			delNode2: true,
			delNode3: true,
			delNode4: true,
		},
		finishDel: true,
	}

	// test maybeUpdateTableForAdd with inDeletingNode being false.
	rbft.status.inActiveState(&rbft.status.inDeletingNode)
	rbft.maybeUpdateTableForDel(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inDeletingNode), "inDeletingNode should be false.")

	// test maybeUpdateTableForAdd with finishDel being false.
	rbft.nodeMgr.delNodeCertStore[key].finishDel = false
	rbft.maybeUpdateTableForDel(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inDeletingNode), "inDeletingNode should be false.")

	// test maybeUpdateTableForAdd with normal case.
	rbft.status.activeState(&rbft.status.inDeletingNode)
	rbft.maybeUpdateTableForDel(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inDeletingNode), "inDeletingNode should be false.")

}

func TestSendAgreeUpdateNforDel(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	key := "test_delete_key"
	delNode1 := DelNode{ReplicaId: 1, Key: key}
	delNode2 := DelNode{ReplicaId: 2, Key: key}
	delNode3 := DelNode{ReplicaId: 3, Key: key}
	delNode4 := DelNode{ReplicaId: 4, Key: key}
	rbft.nodeMgr.delNodeCertStore[key] = &delNodeCert{
		delNodes: map[DelNode]bool{
			delNode1: true,
			delNode2: true,
			delNode3: true,
			delNode4: true,
		},
		finishDel: true,
	}

	// test sendAgreeUpdateNforDel when in viewchange.
	rbft.status.activeState(&rbft.status.inViewChange)
	rbft.sendAgreeUpdateNforDel(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should not be in updating N.")

	// test sendAgreeUpdateNforDel with finishDel being false.
	rbft.status.inActiveState(&rbft.status.inViewChange)
	rbft.nodeMgr.delNodeCertStore[key].finishDel = false
	rbft.sendAgreeUpdateNforDel(key)
	ast.Equal(false, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should not be in updating N.")

	// test sendAgreeUpdateNforDel with normal case.
	rbft.nodeMgr.delNodeCertStore[key].finishDel = true
	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		agree := &AgreeUpdateN{}
		err = proto.Unmarshal(consensus.Payload, agree)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(false, agree.Flag, "flag of delete node should be false.")
	}()
	rbft.sendAgreeUpdateNforDel(key)
	ast.Equal(true, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should be in updating N.")

}

func TestRecvAgreeUpdateN(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	addKey := "test-add"
	addNode1 := AddNode{ReplicaId: 1, Key: addKey}
	addNode2 := AddNode{ReplicaId: 2, Key: addKey}
	addNode3 := AddNode{ReplicaId: 3, Key: addKey}
	addNode4 := AddNode{ReplicaId: 4, Key: addKey}
	rbft.nodeMgr.addNodeCertStore[addKey] = &addNodeCert{
		addNodes: map[AddNode]bool{
			addNode1: true,
			addNode2: true,
			addNode3: true,
			addNode4: true,
		},
		finishAdd: true,
	}

	delKey := "test-delete"
	delNode1 := DelNode{ReplicaId: 1, Key: delKey}
	delNode2 := DelNode{ReplicaId: 2, Key: delKey}
	delNode3 := DelNode{ReplicaId: 3, Key: delKey}
	delNode4 := DelNode{ReplicaId: 4, Key: delKey}
	rbft.nodeMgr.delNodeCertStore[delKey] = &delNodeCert{
		delNodes: map[DelNode]bool{
			delNode1: true,
			delNode2: true,
			delNode3: true,
			delNode4: true,
		},
		delId:     5,
		finishDel: true,
	}

	agreeAdd1 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 2,
			View:      5,
			H:         0,
		},
		Flag: true,
		Key:  addKey,
		N:    5,
	}

	agreeAdd2 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 3,
			View:      5,
			H:         0,
		},
		Flag: true,
		Key:  addKey,
		N:    5,
	}

	agreeAdd3 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 4,
			View:      5,
			H:         0,
		},
		Flag: true,
		Key:  addKey,
		N:    5,
	}

	agreeDel1 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 2,
			View:      4,
			H:         0,
		},
		Flag: false,
		Key:  delKey,
		N:    4,
	}

	agreeDel2 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 3,
			View:      4,
			H:         0,
		},
		Flag: false,
		Key:  delKey,
		N:    4,
	}

	agreeDel3 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 4,
			View:      4,
			H:         0,
		},
		Flag: false,
		Key:  delKey,
		N:    4,
	}

	// test recvAgreeUpdateN when in viewchange.
	rbft.status.activeState(&rbft.status.inViewChange)
	rbft.recvAgreeUpdateN(agreeDel1)

	// test recvAgreeUpdateN when in negotiate view.
	rbft.status.inActiveState(&rbft.status.inViewChange)
	rbft.status.activeState(&rbft.status.inNegoView)
	rbft.recvAgreeUpdateN(agreeDel1)

	// test recvAgreeUpdateN when in recovery.
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.activeState(&rbft.status.inRecovery)
	rbft.recvAgreeUpdateN(agreeDel1)

	// test recvAgreeUpdateN in normal status.
	rbft.status.inActiveState(&rbft.status.inRecovery)
	rbft.N = 5
	rbft.view = 0

	// test recvAgreeUpdateN when finishDel being false.
	rbft.nodeMgr.delNodeCertStore[delKey].finishDel = false
	rbft.recvAgreeUpdateN(agreeDel1)

	// test recvAgreeUpdateN when finishDel being false.
	rbft.nodeMgr.delNodeCertStore[delKey].finishDel = true
	rbft.recvAgreeUpdateN(agreeDel1)
	rbft.recvAgreeUpdateN(agreeDel1)
	ast.Equal(1, len(rbft.nodeMgr.agreeUpdateStore))

	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		agree := &AgreeUpdateN{}
		err = proto.Unmarshal(consensus.Payload, agree)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(false, agree.Flag, "flag of delete node should be false.")
	}()
	rbft.recvAgreeUpdateN(agreeDel2)
	rbft.recvAgreeUpdateN(agreeDel3)
	updateTarget := uidx{v: 4, n: 4, flag: false, key: delKey}
	ast.Equal(updateTarget, rbft.nodeMgr.updateTarget)
	time.Sleep(time.Nanosecond)

	rbft.N = 4
	rbft.view = 0
	// test recvAgreeUpdateN with finishAdd being false.
	rbft.status.inActiveState(&rbft.status.inUpdatingN)
	rbft.nodeMgr.addNodeCertStore[addKey].finishAdd = false
	rbft.recvAgreeUpdateN(agreeAdd1)

	// test recvAgreeUpdateN with incorrect N.
	rbft.nodeMgr.addNodeCertStore[addKey].finishAdd = true
	agreeAdd1.N = 4
	rbft.recvAgreeUpdateN(agreeAdd1)

	// test recvAgreeUpdateN with normal case.
	agreeAdd1.N = 5
	rbft.recvAgreeUpdateN(agreeAdd1)

	rbft.recvAgreeUpdateN(agreeAdd2)

	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		agree := &AgreeUpdateN{}
		err = proto.Unmarshal(consensus.Payload, agree)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(true, agree.Flag, "flag of add node should be true.")
	}()
	rbft.recvAgreeUpdateN(agreeAdd3)

}

func TestSendReadyForN(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	// test sendReadyForN with a non-new node.
	rbft.status.inActiveState(&rbft.status.isNewNode)
	rbft.sendReadyForN()
	ast.Equal(false, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should not be in updating N.")

	// test sendReadyForN with a blank localKey.
	rbft.nodeMgr.localKey = ""
	rbft.sendReadyForN()
	ast.Equal(false, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should not be in updating N.")

	// test sendReadyForN when in viewchange.
	rbft.status.activeState(&rbft.status.inViewChange)
	rbft.sendReadyForN()
	ast.Equal(false, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should not be in updating N.")

	// test sendReadyForN in normal case.
	rbft.status.activeState(&rbft.status.isNewNode)
	key := "test-local-key"
	rbft.nodeMgr.localKey = key
	rbft.status.inActiveState(&rbft.status.inViewChange)
	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		readyForN := &ReadyForN{}
		err = proto.Unmarshal(consensus.Payload, readyForN)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(key, readyForN.Key, "local key shoule be equal.")
	}()
	rbft.sendReadyForN()
	ast.Equal(true, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should be in updating N.")

}

func TestRecvReadyforNforAdd(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	key := "test-ready-for-n"
	ready := &ReadyForN{
		ReplicaId: 2,
		Key:       key,
	}
	rbft.status.activeState(&rbft.status.inUpdatingN)

	// test sendReadyForN when in viewchange.
	rbft.status.activeState(&rbft.status.inViewChange)
	rbft.recvReadyforNforAdd(ready)

	// test sendReadyForN with finishAdd being false.
	rbft.status.inActiveState(&rbft.status.inViewChange)
	rbft.recvReadyforNforAdd(ready)

	// test sendReadyForN with normal case.
	addNode1 := AddNode{ReplicaId: 1, Key: key}
	addNode2 := AddNode{ReplicaId: 2, Key: key}
	addNode3 := AddNode{ReplicaId: 3, Key: key}
	addNode4 := AddNode{ReplicaId: 4, Key: key}
	rbft.nodeMgr.addNodeCertStore[key] = &addNodeCert{
		addNodes: map[AddNode]bool{
			addNode1: true,
			addNode2: true,
			addNode3: true,
			addNode4: true,
		},
		finishAdd: true,
	}
	rbft.N = 4
	rbft.view = 0
	rbft.recvReadyforNforAdd(ready)

}

func TestSendAgreeUpdateNforAdd(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	rbft.N = 4
	rbft.view = 0
	key := "test_add_key"
	agree := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: rbft.id,
			View:      0,
			H:         0,
		},
		Flag: true,
		Key:  key,
		N:    4,
	}

	// test sendAgreeUpdateNForAdd with a new node.
	rbft.status.activeState(&rbft.status.isNewNode)
	rbft.sendAgreeUpdateNForAdd(agree)
	ast.Equal(false, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should not be in updating N.")

	// test sendAgreeUpdateNForAdd with incorrect N or view.
	rbft.status.inActiveState(&rbft.status.isNewNode)
	rbft.sendAgreeUpdateNForAdd(agree)
	ast.Equal(false, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should not be in updating N.")

	// test sendAgreeUpdateNForAdd in normal case.
	agree.N = 5
	agree.Basis.View = 5
	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		agree := &AgreeUpdateN{}
		err = proto.Unmarshal(consensus.Payload, agree)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(key, agree.Key, "local key shoule be equal.")
	}()
	rbft.sendAgreeUpdateNForAdd(agree)
	ast.Equal(true, rbft.status.getState(&rbft.status.inUpdatingN), "Replica should be in updating N.")

}

func TestSendUpdateN(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	// test sendUpdateN when in negotiate view.
	rbft.status.activeState(&rbft.status.inNegoView)
	rbft.sendUpdateN()
	ast.Equal(0, len(rbft.nodeMgr.updateStore), "updateStore should be nil.")
	rbft.status.inActiveState(&rbft.status.inNegoView)

	// test sendUpdateN with existed updateTarget
	updateTarget := uidx{}
	rbft.nodeMgr.updateTarget = updateTarget
	rbft.nodeMgr.updateStore[updateTarget] = &UpdateN{}
	rbft.sendUpdateN()
	ast.Equal(1, len(rbft.nodeMgr.updateStore), "updateStore should only has one element.")
	rbft.nodeMgr.updateStore = make(map[uidx]*UpdateN)

	addKey := "test-add-key"
	aidx1 := aidx{
		v:    5,
		n:    5,
		id:   1,
		flag: true,
	}

	aidx2 := aidx{
		v:    5,
		n:    5,
		id:   2,
		flag: true,
	}

	aidx3 := aidx{
		v:    5,
		n:    5,
		id:   3,
		flag: true,
	}

	aidx4 := aidx{
		v:    5,
		n:    5,
		id:   4,
		flag: true,
	}

	agreeUpdate1 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			ReplicaId: 1,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate2 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			ReplicaId: 2,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate3 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			ReplicaId: 3,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate4 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			ReplicaId: 4,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	// test sendUpdateN when primary is lack of some batches.
	updateTarget = uidx{v: 5, n: 5, flag: true, key: addKey}
	rbft.nodeMgr.updateTarget = updateTarget
	rbft.nodeMgr.agreeUpdateStore[aidx1] = agreeUpdate1
	rbft.nodeMgr.agreeUpdateStore[aidx2] = agreeUpdate2
	rbft.nodeMgr.agreeUpdateStore[aidx3] = agreeUpdate3
	rbft.nodeMgr.agreeUpdateStore[aidx4] = agreeUpdate4
	// go-routine to check result of broadcast UpdateN
	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		updateN := &UpdateN{}
		err = proto.Unmarshal(consensus.Payload, updateN)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(addKey, updateN.Key, "local key shoule be equal.")
	}()
	// go-routine to check result of broadcast fetch missing request batch.
	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		frb := &FetchRequestBatch{}
		err = proto.Unmarshal(consensus.Payload, frb)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal("batch1", frb.BatchDigest, "need to fetch batch1 as we hasn't stored batch1 to txBatchStore.")
	}()
	rbft.sendUpdateN()
	ast.Equal(rbft.id, rbft.nodeMgr.updateStore[updateTarget].ReplicaId)
	rbft.nodeMgr.updateStore = make(map[uidx]*UpdateN)
	rbft.nodeMgr.agreeUpdateStore = make(map[aidx]*AgreeUpdateN)

	// test
	rbft.nodeMgr.updateTarget = uidx{v: 5, n: 5, flag: true, key: addKey}
	agreeUpdate1.Basis.Cset = nil
	agreeUpdate2.Basis.Cset = nil
	agreeUpdate3.Basis.Cset = nil
	agreeUpdate4.Basis.Cset = nil
	rbft.nodeMgr.agreeUpdateStore[aidx1] = agreeUpdate1
	rbft.nodeMgr.agreeUpdateStore[aidx2] = agreeUpdate2
	rbft.nodeMgr.agreeUpdateStore[aidx3] = agreeUpdate3
	rbft.nodeMgr.agreeUpdateStore[aidx4] = agreeUpdate4
	rbft.sendUpdateN()
	ast.Equal(0, len(rbft.nodeMgr.updateStore), "updateStore should be nil.")

}

func TestPrimaryCheckUpdateN(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	cp := Vc_C{
		SequenceNumber: 10,
		Id:             "test-vc",
	}
	replicas := []replicaInfo{
		{id: 2, height: 10, genesis: 0},
	}
	update := &UpdateN{}

	// test primaryCheckUpdateN with incorrect checkpoint.
	rbft.primaryCheckUpdateN(cp, replicas, update)
	ast.Equal(0, len(rbft.storeMgr.missingReqBatches))
}

func TestRecvUpdateN(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	// test recvUpdateN when in viewchange.
	rbft.status.activeState(&rbft.status.inViewChange)
	updateN := &UpdateN{}
	rbft.recvUpdateN(updateN)
	ast.Equal(0, len(rbft.nodeMgr.updateStore), "updateStore should be nil.")
	rbft.status.inActiveState(&rbft.status.inViewChange)

	// test recvUpdateN when in negotiate view.
	rbft.status.activeState(&rbft.status.inNegoView)
	rbft.recvUpdateN(updateN)
	ast.Equal(0, len(rbft.nodeMgr.updateStore), "updateStore should be nil.")
	rbft.status.inActiveState(&rbft.status.inNegoView)

	// test recvUpdateN when in recovery.
	rbft.status.activeState(&rbft.status.inRecovery)
	rbft.recvUpdateN(updateN)
	ast.Equal(0, len(rbft.nodeMgr.updateStore), "updateStore should be nil.")
	rbft.status.inActiveState(&rbft.status.inRecovery)

	// test recvUpdateN with msg sent from non-primary.
	updateN.ReplicaId = 2
	rbft.recvUpdateN(updateN)
	ast.Equal(0, len(rbft.nodeMgr.updateStore), "updateStore should be nil.")

	addKey := "test-add"
	updateN = &UpdateN{
		Flag:      true,
		ReplicaId: 1,
		Key:       addKey,
		N:         5,
		View:      5,
		Xset: map[uint64]string{
			1: "batch1",
			2: "batch2",
			3: "batch3",
		},
	}

	aidx1 := aidx{
		v:    5,
		n:    5,
		id:   1,
		flag: true,
	}

	aidx2 := aidx{
		v:    5,
		n:    5,
		id:   2,
		flag: true,
	}

	aidx3 := aidx{
		v:    5,
		n:    5,
		id:   3,
		flag: true,
	}

	aidx4 := aidx{
		v:    5,
		n:    5,
		id:   4,
		flag: true,
	}

	agreeUpdate1 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
				{SequenceNumber: 3, BatchDigest: "batch3", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			ReplicaId: 1,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate2 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			ReplicaId: 2,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate3 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			ReplicaId: 3,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate4 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
				{SequenceNumber: 3, BatchDigest: "batch3", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			ReplicaId: 4,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}
	rbft.nodeMgr.agreeUpdateStore[aidx1] = agreeUpdate1
	rbft.nodeMgr.agreeUpdateStore[aidx2] = agreeUpdate2
	rbft.recvUpdateN(updateN)

	rbft.nodeMgr.updateTarget = uidx{
		flag: updateN.Flag,
		key:  updateN.Key,
		n:    updateN.N,
		v:    updateN.View,
	}
	rbft.nodeMgr.agreeUpdateStore[aidx3] = agreeUpdate3
	rbft.nodeMgr.agreeUpdateStore[aidx4] = agreeUpdate4
	rbft.recvUpdateN(updateN)

}

func TestReplicaCheckUpdateN(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	addKey := "test-add"
	updateN := &UpdateN{
		Flag:      true,
		ReplicaId: 1,
		Key:       addKey,
		N:         5,
		View:      5,
		Xset: map[uint64]string{
			1: "batch1",
			2: "batch2",
		},
	}

	aidx1 := aidx{
		v:    5,
		n:    5,
		id:   1,
		flag: true,
	}

	aidx2 := aidx{
		v:    5,
		n:    5,
		id:   2,
		flag: true,
	}

	aidx3 := aidx{
		v:    5,
		n:    5,
		id:   3,
		flag: true,
	}

	aidx4 := aidx{
		v:    5,
		n:    5,
		id:   4,
		flag: true,
	}

	agreeUpdate1 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
				{SequenceNumber: 3, BatchDigest: "batch3", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
			},
			ReplicaId: 1,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate2 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			ReplicaId: 2,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate3 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			ReplicaId: 3,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}

	agreeUpdate4 := &AgreeUpdateN{
		Basis: &VcBasis{
			View: 5,
			H:    0,
			Cset: []*Vc_C{
				{SequenceNumber: 0, Id: "chkpt0"},
			},
			Pset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			Qset: []*Vc_PQ{
				{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
				{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
			},
			ReplicaId: 4,
			Genesis:   0,
		},
		Flag:       true,
		Key:        addKey,
		RouterHash: "",
		N:          5,
	}
	rbft.nodeMgr.agreeUpdateStore[aidx1] = agreeUpdate1
	rbft.nodeMgr.agreeUpdateStore[aidx2] = agreeUpdate2
	rbft.nodeMgr.agreeUpdateStore[aidx3] = agreeUpdate3
	rbft.nodeMgr.agreeUpdateStore[aidx4] = agreeUpdate4
	rbft.nodeMgr.updateTarget = uidx{
		flag: updateN.Flag,
		key:  updateN.Key,
		n:    updateN.N,
		v:    updateN.View,
	}

	rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget] = updateN

	// test replicaCheckUpdateN when in viewchange.
	rbft.status.activeState(&rbft.status.inViewChange)
	rbft.replicaCheckUpdateN()
	ast.Equal(4, rbft.N)
	rbft.status.inActiveState(&rbft.status.inViewChange)

	// test replicaCheckUpdateN when not in updatingN.
	rbft.status.inActiveState(&rbft.status.inUpdatingN)
	rbft.replicaCheckUpdateN()
	ast.Equal(4, rbft.N)
	rbft.status.activeState(&rbft.status.inUpdatingN)

	// test replicaCheckUpdateN with incorrect Cset.
	rbft.nodeMgr.agreeUpdateStore[aidx1].Basis.Cset = []*Vc_C{}
	rbft.nodeMgr.agreeUpdateStore[aidx2].Basis.Cset = []*Vc_C{}
	rbft.nodeMgr.agreeUpdateStore[aidx3].Basis.Cset = []*Vc_C{}
	rbft.nodeMgr.agreeUpdateStore[aidx4].Basis.Cset = []*Vc_C{}
	rbft.replicaCheckUpdateN()
	ast.Equal(4, rbft.N)
	rbft.nodeMgr.agreeUpdateStore[aidx1].Basis.Cset = []*Vc_C{{SequenceNumber: 0, Id: "chkpt0"}}
	rbft.nodeMgr.agreeUpdateStore[aidx2].Basis.Cset = []*Vc_C{{SequenceNumber: 0, Id: "chkpt0"}}
	rbft.nodeMgr.agreeUpdateStore[aidx3].Basis.Cset = []*Vc_C{{SequenceNumber: 0, Id: "chkpt0"}}
	rbft.nodeMgr.agreeUpdateStore[aidx4].Basis.Cset = []*Vc_C{{SequenceNumber: 0, Id: "chkpt0"}}

	// test replicaCheckUpdateN with incorrect msgList.
	rbft.nodeMgr.agreeUpdateStore[aidx4].Basis.Pset = []*Vc_PQ{
		{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
		{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
		{SequenceNumber: 3, BatchDigest: "batch3", View: 0},
	}
	rbft.replicaCheckUpdateN()
	ast.Equal(4, rbft.N)
	rbft.nodeMgr.agreeUpdateStore[aidx4].Basis.Pset = []*Vc_PQ{
		{SequenceNumber: 1, BatchDigest: "batch1", View: 0},
		{SequenceNumber: 2, BatchDigest: "batch2", View: 0},
	}

	// test replicaCheckUpdateN with incorrect Xset.
	rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget].Xset = make(map[uint64]string)
	rbft.replicaCheckUpdateN()
	ast.Equal(4, rbft.N)
	rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget].Xset = map[uint64]string{
		1: "batch1",
		2: "batch2",
	}

	// test replicaCheckUpdateN with updateHandled being true.
	rbft.status.activeState(&rbft.status.updateHandled)
	rbft.replicaCheckUpdateN()
	ast.Equal(4, rbft.N)
	rbft.status.inActiveState(&rbft.status.updateHandled)

	// test replicaCheckUpdateN in normal case.
	rbft.replicaCheckUpdateN()
	ast.Equal(5, rbft.N)

	rbft.nodeMgr.agreeUpdateStore[aidx1] = agreeUpdate1
	rbft.nodeMgr.agreeUpdateStore[aidx2] = agreeUpdate2
	rbft.nodeMgr.agreeUpdateStore[aidx3] = agreeUpdate3
	rbft.nodeMgr.agreeUpdateStore[aidx4] = agreeUpdate4
	rbft.N = 4
	rbft.view = 0
	rbft.status.inActiveState(&rbft.status.updateHandled)
	rbft.status.activeState(&rbft.status.skipInProgress)
	rbft.replicaCheckUpdateN()
	ast.Equal(5, rbft.N)
	rbft.status.inActiveState(&rbft.status.skipInProgress)

}

func TestSendFinishUpdate(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	rbft.N = 5
	rbft.view = 5
	rbft.h = 0

	rbft.status.inActiveState(&rbft.status.inUpdatingN)
	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		finish := &FinishUpdate{}
		err = proto.Unmarshal(consensus.Payload, finish)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(rbft.id, finish.ReplicaId, "id should be equal.")
	}()
	rbft.sendFinishUpdate()

	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		finish := &FinishUpdate{}
		err = proto.Unmarshal(consensus.Payload, finish)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(rbft.id, finish.ReplicaId, "id should be equal.")
	}()
	rbft.sendFinishUpdate()

}

func TestProcessReqInUpdate(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	finish1 := &FinishUpdate{
		ReplicaId: 1,
		View:      5,
		LowH:      0,
	}
	finish2 := &FinishUpdate{
		ReplicaId: 2,
		View:      5,
		LowH:      0,
	}
	finish3 := &FinishUpdate{
		ReplicaId: 3,
		View:      5,
		LowH:      0,
	}
	finish4 := &FinishUpdate{
		ReplicaId: 4,
		View:      5,
		LowH:      0,
	}
	finish5 := &FinishUpdate{
		ReplicaId: 5,
		View:      5,
		LowH:      0,
	}

	rbft.id = 5
	rbft.view = 5
	rbft.N = 5
	rbft.f = 1
	rbft.exec.lastExec = 2
	// simulate progress of receiving finish update
	rbft.recvFinishUpdate(finish2)
	rbft.recvFinishUpdate(finish3)
	rbft.recvFinishUpdate(finish4)
	rbft.recvFinishUpdate(finish5)
	rbft.status.activeState(&rbft.status.inVcReset)
	rbft.recvFinishUpdate(finish1)
	rbft.status.inActiveState(&rbft.status.inVcReset)
	delete(rbft.nodeMgr.finishUpdateStore, *finish1)
	rbft.recvFinishUpdate(finish1)
	ast.Equal(uint64(0), rbft.seqNo)

	addKey := "test-add"
	updateN := &UpdateN{
		Flag:      true,
		ReplicaId: 1,
		Key:       addKey,
		N:         5,
		View:      5,
		Xset: map[uint64]string{
			1: "batch1",
			2: "batch2",
		},
	}
	rbft.nodeMgr.updateTarget = uidx{
		flag: updateN.Flag,
		key:  updateN.Key,
		n:    updateN.N,
		v:    updateN.View,
	}

	rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget] = updateN
	rbft.processReqInUpdate()
	ast.Equal(uint64(2), rbft.seqNo)
	ast.Equal(uint64(2), rbft.batchVdr.lastVid)
}

func TestPutBackTxBatches(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	xset := map[uint64]string{
		1: "batch1",
		2: "batch2",
		3: "batch3",
	}
	rbft.exec.lastExec = 2
	rbft.storeMgr.txBatchStore = map[string]*TransactionBatch{
		"batch1": {SeqNo: 1},
		"batch2": {SeqNo: 2},
		"batch3": {SeqNo: 3},
		"batch4": {SeqNo: 4},
		"batch5": {SeqNo: 5},
	}

	rbft.putBackTxBatches(xset)

}

func TestRebuildCertStoreForUpdate(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	// test rebuildCertStoreForUpdate with no updateTarget.
	rbft.rebuildCertStoreForUpdate()

	// test rebuildCertStoreForUpdate with no nil Xset.
	addKey := "test-add"
	updateN := &UpdateN{
		Flag:      true,
		ReplicaId: 1,
		Key:       addKey,
		N:         5,
		View:      5,
		Xset:      make(map[uint64]string),
	}
	rbft.nodeMgr.updateTarget = uidx{
		flag: updateN.Flag,
		key:  updateN.Key,
		n:    updateN.N,
		v:    updateN.View,
	}

	rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget] = updateN
}
