package blockpool

import (
	"hyperchain/hyperdb"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/core"
	"hyperchain/common"
	"errors"
	"hyperchain/event"
)

// run transaction in a sandbox
// execution result will not been add to database
func (pool *BlockPool) RunInSandBox(tx *types.Transaction) error {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	// load latest state status
	v := pool.lastValidationState.Load()
	initStatus, ok := v.(common.Hash)
	if ok == false {
		return errors.New("Get StateDB Status Failed!")
	}
	// initialize state
	// todo use state copy instead of itself
	state, err := pool.GetStateInstance(initStatus, db)
	if err != nil {
		return err
	}
	// initialize execution environment
	fakeBlockNumber := core.GetHeightOfChain()
	sandBox := initEnvironment(state, fakeBlockNumber)
	receipt, _, _, err := core.ExecTransaction(*tx, sandBox)
	if err != nil{
		var errType types.InvalidTransactionRecord_ErrType
		if core.IsValueTransferErr(err) {
			errType = types.InvalidTransactionRecord_OUTOFBALANCE
		} else if core.IsExecContractErr(err) {
			tmp := err.(*core.ExecContractError)
			if tmp.GetType() == 0 {
				errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
			} else if tmp.GetType() == 1{
				errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
			} else {
				// For extension
			}
		} else {
			// For extension
		}
		t :=  &types.InvalidTransactionRecord{
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
		pool.StoreInvalidResp(event.RespInvalidTxsEvent{
			Payload: payload,
		})
		return nil
	} else {
		// persist execution result to local
		err, _ := core.PersistReceipt(db.NewBatch(), receipt, pool.conf.TransactionVersion, true, true)
		if err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}
