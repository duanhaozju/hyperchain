package blockpool

import (
	"hyperchain/hyperdb"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/core"
	"strconv"
	"hyperchain/common"
	"errors"
	"hyperchain/core/vm/params"
	"hyperchain/event"
)

func (pool *BlockPool) RunInSandBox(tx *types.Transaction) error {
	// TODO add block number to specify the initial status
	var env = make(map[string]string)
	fakeBlockNumber := core.GetHeightOfChain()
	env["currentNumber"] = strconv.FormatUint(fakeBlockNumber, 10)
	env["currentGasLimit"] = "10000000"
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	v := pool.lastValidationState.Load()
	initStatus, ok := v.(common.Hash)
	if ok == false {
		return errors.New("Get StateDB Status Failed!")
	}
	state, err := pool.GetStateInstance(initStatus, db)
	if err != nil {
		return err
	}
	sandBox := core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, state, env)
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
		pool.StoreInvalidResp(event.RespInvalidTxsEvent{
			Payload: payload,
		})
		return nil
	} else {
		err, _ := core.PersistReceipt(db.NewBatch(), receipt, pool.conf.TransactionVersion, true, true)
		if err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}
