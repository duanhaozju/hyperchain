package blockpool

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/hyperdb"
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
	// IMPORTANT all change to state will not been persist cause never commit will been invoked
	state, err := pool.GetStateInstanceForSimulate(initStatus, db)
	if err != nil {
		return err
	}
	// initialize execution environment
	fakeBlockNumber := core.GetHeightOfChain()
	sandBox := initEnvironment(state, fakeBlockNumber+1)
	receipt, _, _, err := core.ExecTransaction(tx, sandBox)
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
		pool.StoreInvalidResp(event.RespInvalidTxsEvent{
			Payload: payload,
		})
		return nil
	} else {
		// persist execution result to local
		err, _ := core.PersistReceipt(db.NewBatch(), receipt, pool.GetTransactionVersion(), true, true)
		if err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}
