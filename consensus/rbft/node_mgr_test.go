package rbft

import (
	"github.com/stretchr/testify/assert"
	"hyperchain/common"
	"hyperchain/consensus/helper/persist"
	"hyperchain/hyperdb/mdb"
	"hyperchain/manager/protos"
	"testing"
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
