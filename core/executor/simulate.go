package executor

import (
	"github.com/golang/protobuf/proto"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/core/vm"
)

// run transaction in a sandbox
// execution result will not been add to database
func (executor *Executor) RunInSandBox(tx *types.Transaction, snapshotId string) error {
	var statedb vm.Database
	var err error
	var callback func()
	if snapshotId == "" {
		statedb, err = executor.newStateDb()
	} else {
		statedb, err, callback = executor.initHistoryStateDb(snapshotId)
	}
	if err != nil {
		return err
	}
	// release history db handler
	defer func() {
		if callback != nil {
			callback()
		}
	}()
	// initialize execution environment
	receipt, _, _, err := executor.ExecTransaction(tx, statedb)
	if err != nil {
		errType := executor.classifyInvalid(err)
		t := &types.InvalidTransactionRecord{
			Tx:      tx,
			ErrType: errType,
			ErrMsg:  []byte(err.Error()),
		}
		payload, err := proto.Marshal(t)
		if err != nil {
			executor.logger.Error("Marshal tx error")
			return err
		}
		// persist execution result to local
		executor.StoreInvalidTransaction(payload)
		return nil
	} else {
		// persist execution result to local
		err, _ := edb.PersistReceipt(executor.db.NewBatch(), receipt, true, true)
		if err != nil {
			executor.logger.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}
