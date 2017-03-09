package executor

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/hyperdb"
	edb "hyperchain/core/db_utils"
)

// run transaction in a sandbox
// execution result will not been add to database
func (executor *Executor) RunInSandBox(tx *types.Transaction) error {
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		return err
	}
	statedb, err := executor.newStateDb()
	if err != nil {
		return err
	}
	// initialize execution environment
	fakeBlockNumber := edb.GetHeightOfChain(executor.namespace) + 1
	sandBox := initEnvironment(statedb, fakeBlockNumber)
	receipt, _, _, err := core.ExecTransaction(tx, sandBox)
	log.Notice("DEBUG")
	if err != nil {
		var errType types.InvalidTransactionRecord_ErrType
		if core.IsValueTransferErr(err) {
			errType = types.InvalidTransactionRecord_OUTOFBALANCE
		} else if core.IsExecContractErr(err) {
			tmp := err.(*core.ExecContractError)
			if tmp.GetType() == 0 {
				errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
			} else if tmp.GetType() == 1 {
				errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
			}
		}
		t := &types.InvalidTransactionRecord{
			Tx:      tx,
			ErrType: errType,
			ErrMsg:  []byte(err.Error()),
		}
		payload, err := proto.Marshal(t)
		if err != nil {
			log.Error("Marshal tx error")
			return nil
		}
		// persist execution result to local
		executor.StoreInvalidTransaction(event.RespInvalidTxsEvent{
			Payload: payload,
		})
		return nil
	} else {
		// persist execution result to local
		err, _ := edb.PersistReceipt(db.NewBatch(), receipt, true, true)
		if err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}
