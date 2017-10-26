//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"testing"
	"time"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/consensus/helper/persist"
	mdb "github.com/hyperchain/hyperchain/hyperdb/mdb"
	"github.com/hyperchain/hyperchain/manager/event"

	"github.com/stretchr/testify/assert"
)

func TestRbftTimeFunctions(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	timerName := "test"
	timerDuration := 5 * time.Second
	rbft.timerMgr.newTimer(timerName, timerDuration)

	ok := rbft.timerMgr.containsTimer(timerName)
	ast.Equal(true, ok, "newTimer failed")
	ast.Equal(timerDuration, rbft.timerMgr.getTimeoutValue(timerName), "getTimeoutValue failed")
	ast.Equal(0*time.Second, rbft.timerMgr.getTimeoutValue("notexist"), "getTimeoutValue failed")
	rbft.timerMgr.setTimeoutValue(timerName, 10*time.Second)
	rbft.timerMgr.setTimeoutValue("notexist", 10*time.Second)
	ast.Equal(10*time.Second, rbft.timerMgr.getTimeoutValue(timerName), "setTimeoutValue failed")

	queue := &event.TypeMux{}
	rbft.timerMgr.startTimer(timerName, nil, queue)
	ast.Equal(1, len(rbft.timerMgr.ttimers[timerName].isActive), "startTimer failed")
	rbft.timerMgr.startTimerWithNewTT(timerName, 20*time.Second, nil, queue)
	ast.Equal(2, len(rbft.timerMgr.ttimers[timerName].isActive), "startTimer failed")

	rbft.timerMgr.stopOneTimer(timerName, 1)
	ast.Equal(false, rbft.timerMgr.ttimers[timerName].isActive[1], "stopOneTimer failed")

	isActive := rbft.timerMgr.ttimers[timerName].isActive
	rbft.timerMgr.stopOneTimer("notexist", 1)
	ast.Equal(isActive, rbft.timerMgr.ttimers[timerName].isActive, "stopOneTimer failed")
	rbft.timerMgr.stopOneTimer(timerName, 3)
	ast.Equal(isActive, rbft.timerMgr.ttimers[timerName].isActive, "stopOneTimer failed")
}
