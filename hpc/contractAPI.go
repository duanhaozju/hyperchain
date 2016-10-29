package hpc

import (
	"hyperchain/common"
	"hyperchain/core/vm/compiler"
	"time"
	"github.com/golang/protobuf/proto"
	"hyperchain/manager"
	"hyperchain/core/types"
	"hyperchain/event"
	"fmt"
	"errors"
	"encoding/hex"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/core"
)

type PublicContractAPI struct {
	eventMux *event.TypeMux
	pm *manager.ProtocolManager
	db *hyperdb.LDBDatabase
}

func NewPublicContractAPI(eventMux *event.TypeMux, pm *manager.ProtocolManager, hyperDb *hyperdb.LDBDatabase) *PublicContractAPI {
	return &PublicContractAPI{
		eventMux :eventMux,
		pm:pm,
		db:hyperDb,
	}
}

func deployOrInvoke(contract *PublicContractAPI, args SendTxArgs) (common.Hash, error) {
	var tx *types.Transaction
	//var found bool
	realArgs := prepareExcute(args)

	// ################################# 测试代码 START ####################################### //
	if realArgs.Timestamp == 0 {
		realArgs.Timestamp = time.Now().UnixNano()
	}
	// ################################# 测试代码 END ####################################### //

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(),realArgs.Gas.ToInt64(),realArgs.Value.ToInt64(),payload)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, err
	}

	if args.To == nil {

		// 部署合约
		//tx = types.NewTransaction(realArgs.From[:], nil, value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp)

	} else {

		// 调用合约或者普通交易(普通交易还需要加检查余额)
		//tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp)
	}

	//if contract.pm == nil {
	//
	//	// Test environment
	//	found = true
	//} else {
	//
	//	// Development environment
	//	am := contract.pm.AccountManager
	//	_, found = am.Unlocked[args.From]
	//
	//}
	//am := tran.pm.AccountManager

	//if found == true {
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		tx.Id = uint64(contract.pm.Peermanager.GetNodeId())

		if realArgs.PrivKey == "" {
			// For Hyperchain test

			var signature []byte
			if realArgs.Signature == "" {

				signature, err = contract.pm.AccountManager.Sign(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes())
				if err != nil {
					log.Errorf("Sign(tx) error :%v", err)
				}
			} else {
				tx.Signature = common.FromHex(realArgs.Signature)
			}

			tx.Signature = signature
		} else {
			// For Hyperboard test

			key, err := hex.DecodeString(args.PrivKey)
			if err != nil {
				return common.Hash{}, err
			}
			pri := crypto.ToECDSA(key)

			hash := tx.SighHash(kec256Hash).Bytes()
			sig, err := encryption.Sign(hash, pri)
			if err != nil {
				return common.Hash{}, err
			}

			tx.Signature = sig
		}

		tx.TransactionHash = tx.BuildHash().Bytes()
		// Unsign Test
		if !tx.ValidateSign(contract.pm.AccountManager.Encryption, kec256Hash) {
			log.Error("invalid signature")
			// 不要返回，因为要将失效交易存到db中
			//return common.Hash{}, errors.New("invalid signature")
		}

		txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Errorf("proto.Marshal(tx) error: %v", err)
		}
		if manager.GetEventObject() != nil {
			go contract.eventMux.Post(event.NewTxEvent{Payload: txBytes})
			//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		} else {
			log.Warning("manager is Nil")
		}

		start_getErr := time.Now().Unix()
		end_getErr :=start_getErr + TIMEOUT
		var errMsg string
		for start_getErr := start_getErr; start_getErr < end_getErr; start_getErr = time.Now().Unix() {
			errType, _ := core.GetInvaildTxErrType(contract.db, tx.GetTransactionHash().Bytes());

			if errType != -1 {
				errMsg = errType.String()
				break;
			} else if rept := core.GetReceipt(tx.GetTransactionHash());rept != nil {
				break
			}


		}
		if start_getErr != end_getErr && errMsg != "" {
			return common.Hash{}, errors.New(errMsg)
		} else if start_getErr == end_getErr {
			return common.Hash{}, errors.New("Sending return timeout,may be something wrong.")
	}


		log.Infof("############# %d: end send request#############", time.Now().Unix())
	//} else {
	//	return common.Hash{}, errors.New("account don't unlock")
	//}

	time.Sleep(2000 * time.Millisecond)

	return tx.GetTransactionHash(), nil
}

type CompileCode struct{
	Abi []string
	Bin []string
	Types []string
}

// ComplieContract complies contract to ABI
func (contract *PublicContractAPI) CompileContract(ct string) (*CompileCode,error){
	abi, bin, names, err := compiler.CompileSourcefile(ct)

	if err != nil {
		return nil, err
	}

	return &CompileCode{
		Abi: abi,
		Bin: bin,
		Types: names,
	}, nil
}

// DeployContract deploys contract.
func (contract *PublicContractAPI) DeployContract(args SendTxArgs) (common.Hash, error) {
	return deployOrInvoke(contract, args)
}

// InvokeContract invokes contract.
func (contract *PublicContractAPI) InvokeContract(args SendTxArgs) (common.Hash, error) {
	return deployOrInvoke(contract, args)
}

// GetCode returns the code from the given contract address and block number.
func (contract *PublicContractAPI) GetCode(addr common.Address, n Number) (string, error) {

	stateDb, err := getBlockStateDb(n, contract.db)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return "", err
	}

	return fmt.Sprintf(`0x%x`, stateDb.GetCode(addr)), nil
}

// GetContractCountByAddr returns the number of contract that has been deployed by given account address and block number,
// if addr is nil, returns the number of all the contract that has been deployed.
func (contract *PublicContractAPI) GetContractCountByAddr(addr common.Address, n Number) (*Number, error) {

	stateDb, err := getBlockStateDb(n, contract.db)

	if err != nil {
		return nil, err
	}

	return NewUint64ToNumber(stateDb.GetNonce(addr)), nil

}

// GetStorageByAddr returns the storage by given contract address and bock number.
// The method is offered for hyperchain internal test.
func (contract *PublicContractAPI) GetStorageByAddr(addr common.Address, n Number) (map[string]string, error) {
	stateDb, err := getBlockStateDb(n, contract.db)

	if err != nil {
		return nil, err
	}
	mp := make(map[string]string)

	if obj := stateDb.GetStateObject(addr);obj == nil {
		return nil, nil
	} else {
		storages := obj.Storage()
		if len(storages) == 0 {
			return nil, nil
		}

		for k,v := range storages {
			mp[k.Hex()] = v.Hex()
		}
	}
	return mp,nil
}

