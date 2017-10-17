package rbft

import (
	"testing"
	"hyperchain/hyperdb/mdb"
	"github.com/stretchr/testify/assert"
	"hyperchain/common"
	"hyperchain/consensus/helper/persist"
	"hyperchain/manager/protos"
	"hyperchain/consensus/consensusMocks"
	"sync/atomic"
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

func TestRecvLocalNewNode(t *testing.T){
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
		Id: 1,
		Del: 5,
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
	addNode1 := AddNode{ReplicaId:1, Key: key}
	addNode2 := AddNode{ReplicaId:2, Key: key}
	addNode3 := AddNode{ReplicaId:3, Key: key}
	addNode4 := AddNode{ReplicaId:4, Key: key}
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
	delNode1 := DelNode{ReplicaId:1, Key: key}
	delNode2 := DelNode{ReplicaId:2, Key: key}
	delNode3 := DelNode{ReplicaId:3, Key: key}
	delNode4 := DelNode{ReplicaId:4, Key: key}
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
	delNode1 := DelNode{ReplicaId:1, Key: key}
	delNode2 := DelNode{ReplicaId:2, Key: key}
	delNode3 := DelNode{ReplicaId:3, Key: key}
	delNode4 := DelNode{ReplicaId:4, Key: key}
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
	atomic.StoreUint32(&rbft.activeView, 0)
	rbft.sendAgreeUpdateNforDel(key)
	ast.Equal(0, atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN), "Replica should not be in updating N.")

	// test sendAgreeUpdateNforDel with finishDel being false.
	atomic.StoreUint32(&rbft.activeView, 1)
	rbft.nodeMgr.delNodeCertStore[key].finishDel = false
	rbft.sendAgreeUpdateNforDel(key)
	ast.Equal(1, atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN), "Replica should not in updating N.")

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

	delKey := "test-agree-update-n"

	delNode1 := DelNode{ReplicaId:1, Key: delKey}
	delNode2 := DelNode{ReplicaId:2, Key: delKey}
	delNode3 := DelNode{ReplicaId:3, Key: delKey}
	delNode4 := DelNode{ReplicaId:4, Key: delKey}
	rbft.nodeMgr.delNodeCertStore[delKey] = &delNodeCert{
		delNodes: map[DelNode]bool{
			delNode1: true,
			delNode2: true,
			delNode3: true,
			delNode4: true,
		},
		delId: 5,
		finishDel: true,
	}

	agreeAdd1 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 2,
			View:      4,
			H:         0,
		},
		Flag: false,
		Key:  delKey,
		N:    4,
	}

	agreeAdd2 := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: 3,
			View:      4,
			H:         0,
		},
		Flag: false,
		Key:  delKey,
		N:    4,
	}

	// test recvAgreeUpdateN when in viewchange.
	atomic.StoreUint32(&rbft.activeView, 0)
	rbft.recvAgreeUpdateN(agreeAdd1)

	// test recvAgreeUpdateN when in negotiate view.
	atomic.StoreUint32(&rbft.activeView, 1)
	rbft.status.activeState(&rbft.status.inNegoView)
	rbft.recvAgreeUpdateN(agreeAdd1)

	// test recvAgreeUpdateN when in recovery.
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.activeState(&rbft.status.inRecovery)
	rbft.recvAgreeUpdateN(agreeAdd1)

	// test recvAgreeUpdateN in normal status.
	rbft.status.inActiveState(&rbft.status.inRecovery)
	rbft.N = 5
	rbft.view = 0

	// test recvAgreeUpdateN when finishDel being false.
	rbft.nodeMgr.delNodeCertStore[delKey].finishDel = false
	rbft.recvAgreeUpdateN(agreeAdd1)

	// test recvAgreeUpdateN when finishDel being false.
	rbft.nodeMgr.delNodeCertStore[delKey].finishDel = true
	rbft.recvAgreeUpdateN(agreeAdd1)
	rbft.recvAgreeUpdateN(agreeAdd1)
	ast.Equal(1, len(rbft.nodeMgr.agreeUpdateStore))

	rbft.recvAgreeUpdateN(agreeAdd2)
	updateTarget := uidx{v: 4, n: 4, flag: false, key: delKey}
	ast.Equal(updateTarget, rbft.nodeMgr.updateTarget)


}
