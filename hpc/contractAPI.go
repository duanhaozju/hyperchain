//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
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
	"hyperchain/hyperdb"
	"github.com/juju/ratelimit"
)

type PublicContractAPI struct {
	eventMux *event.TypeMux
	pm *manager.ProtocolManager
	db *hyperdb.LDBDatabase
	tokenBucket *ratelimit.Bucket
	ratelimitEnable bool
}

func NewPublicContractAPI(eventMux *event.TypeMux, pm *manager.ProtocolManager, hyperDb *hyperdb.LDBDatabase, ratelimitEnable bool, bmax int64, rate time.Duration) *PublicContractAPI {
	return &PublicContractAPI{
		eventMux :eventMux,
		pm:pm,
		db:hyperDb,
		tokenBucket: ratelimit.NewBucket(rate, bmax),
		ratelimitEnable: ratelimitEnable,
	}
}

func deployOrInvoke(contract *PublicContractAPI, args SendTxArgs, txType int) (common.Hash, error) {
	var tx *types.Transaction
	realArgs, err := prepareExcute(args, txType)
	if err != nil {
		return common.Hash{}, err
	}

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

	log.Infof("############# %d: start send request#############", time.Now().Unix())
	tx.Id = uint64(contract.pm.Peermanager.GetNodeId())

	// For Hyperchain test

	if realArgs.Signature == "" {
		if signature, err := contract.pm.AccountManager.Sign(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes()); err != nil {
			log.Errorf("Sign(tx) error :%v", err)
			return common.Hash{}, err
		} else {
			tx.Signature = signature
		}
	} else {
		tx.Signature = common.FromHex(realArgs.Signature)
	}

	tx.TransactionHash = tx.BuildHash().Bytes()

	// Unsign Test
	if !tx.ValidateSign(contract.pm.AccountManager.Encryption, kec256Hash) {
		log.Error("invalid signature")
		// ATTENTION, return invalid transactino directly
		return common.Hash{}, errors.New("invalid signature")
	}

	if txBytes, err := proto.Marshal(tx); err != nil {
		log.Errorf("proto.Marshal(tx) error: %v", err)
		return common.Hash{}, errors.New("proto.Marshal(tx) happened error")
	} else if manager.GetEventObject() != nil {
		go contract.eventMux.Post(event.NewTxEvent{Payload: txBytes})
	} else {
		log.Error("manager is Nil")
		return common.Hash{}, errors.New("EventObject is nil")
	}

	log.Infof("############# %d: end send request#############", time.Now().Unix())

	//time.Sleep(2000 * time.Millisecond)

	return tx.GetTransactionHash(), nil
}

type CompileCode struct{
	Abi []string `json:"abi"`
	Bin []string	`json:"bin"`
	Types []string	`json:"types"`
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
	if contract.ratelimitEnable && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, errors.New("System is too busy to response ")
	}
	return deployOrInvoke(contract, args, 1)
}

// InvokeContract invokes contract.
func (contract *PublicContractAPI) InvokeContract(args SendTxArgs) (common.Hash, error) {
	if contract.ratelimitEnable && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, errors.New("System is too busy to response ")
	}
	return deployOrInvoke(contract, args, 2)
}

// GetCode returns the code from the given contract address and block number.
func (contract *PublicContractAPI) GetCode(addr common.Address, n BlockNumber) (string, error) {

	stateDb, err := getBlockStateDb(n, contract.db)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return "", err
	}

	return fmt.Sprintf(`0x%x`, stateDb.GetCode(addr)), nil
}

// GetContractCountByAddr returns the number of contract that has been deployed by given account address and block number,
// if addr is nil, returns the number of all the contract that has been deployed.
func (contract *PublicContractAPI) GetContractCountByAddr(addr common.Address, n BlockNumber) (*Number, error) {

	stateDb, err := getBlockStateDb(n, contract.db)

	if err != nil {
		return nil, err
	}

	return NewUint64ToNumber(stateDb.GetNonce(addr)), nil

}

// GetStorageByAddr returns the storage by given contract address and bock number.
// The method is offered for hyperchain internal test.
func (contract *PublicContractAPI) GetStorageByAddr(addr common.Address, n BlockNumber) (map[string]string, error) {
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

