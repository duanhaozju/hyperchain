package executor

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/event"
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
	sandBox := initEnvironment(statedb, fakeBlockNumber)
	receipt, _, _, err := ExecTransaction(tx, sandBox)
	if err != nil {
		errType := executor.classifyInvalid(err)
		t := &types.InvalidTransactionRecord{
			Tx:      tx,
			ErrType: errType,
			ErrMsg:  []byte(err.Error()),
		}
		payload, err := proto.Marshal(t)
		if err != nil {
			log.Error("Marshal tx error")
			return err
		}
		// persist execution result to local
		executor.StoreInvalidTransaction(event.InvalidTxsEvent{
			Payload: payload,
		})
		return nil
	} else {
		// persist execution result to local
		err, _ := edb.PersistReceipt(executor.db.NewBatch(), receipt, true, true)
		if err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}
