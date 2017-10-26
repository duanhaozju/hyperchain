//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitOperation(t *testing.T) {
	ast := assert.New(t)

	stMgr := newStatusMgr()
	stMgr.reset()
	stMgr.setBit(0)
	ast.EqualValues(1, stMgr.status, "status should be 1.")
	stMgr.clearBit(0)
	ast.EqualValues(0, stMgr.status, "status should be 0.")
	ast.Equal(false, stMgr.hasBit(0), "should not have bit on 0 position.")
	stMgr.setBit(0)
	ast.Equal(true, stMgr.hasBit(0), "should have bit on 0 position.")
}

func TestOnOffIn(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	ast.Equal(true, rbft.in(inNegotiateView), "Rbft should be in negotiate view after start.")
	ast.Equal(true, rbft.in(inRecovery), "Rbft should be in recovery after start.")

	rbft.status.reset()
	rbft.on(inUpdatingN)
	ast.Equal(true, rbft.in(inUpdatingN), "Rbft should be in updating N.")
	rbft.off(inUpdatingN)
	ast.Equal(false, rbft.in(inUpdatingN), "Rbft should not be in updating N.")

	rbft.status.reset()
	rbft.on(inRecovery)
	ast.Equal(false, rbft.inAll(inRecovery, stateTransferring), "Rbft is not in state transfer now.")
	rbft.on(stateTransferring)
	ast.Equal(true, rbft.inAll(inRecovery, stateTransferring), "Rbft is in recovery and state transfer.")

	rbft.status.reset()
	rbft.on(inRecovery)
	ast.Equal(true, rbft.inOne(inRecovery, inViewChange), "Rbft is in recovery now.")
	rbft.off(inRecovery)
	ast.Equal(false, rbft.inOne(inRecovery, inViewChange), "Rbft is not in recovery or viewchange.")

}

func TestNormalAndFull(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	ast.Equal(true, rbft.isNormal(), "Rbft is in normal after start.")
	ast.Equal(false, rbft.isPoolFull(), "TxPool is not full after start.")

	rbft.setAbNormal()
	ast.Equal(false, rbft.isNormal(), "Rbft is abnormal now.")
	rbft.setNormal()
	ast.Equal(true, rbft.isNormal(), "Rbft is normal now.")

	rbft.setFull()
	ast.Equal(true, rbft.isPoolFull(), "TxPool is full now.")
	rbft.setNotFull()
	ast.Equal(false, rbft.isPoolFull(), "TxPool is not full now.")

}
