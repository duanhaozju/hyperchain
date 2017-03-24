package executor

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	edb "hyperchain/core/db_utils"
)

// run transaction in a sandbox
// execution result will not been add to database
func (executor *Executor) RunInSandBox(tx *types.Transaction) error {
	statedb, err := executor.newStateDb()
	if err != nil {
		return err
	}
	// initialize execution environment
	fakeBlockNumber := edb.GetHeightOfChain(executor.namespace) + 1
	sandBox := executor.initEnvironment(statedb, fakeBlockNumber)
	receipt, _, _, err := executor.ExecTransaction(tx, sandBox)
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
