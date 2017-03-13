//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/compiler"
	"hyperchain/crypto/hmEncryption"
	"hyperchain/event"
	"hyperchain/manager"
	"math/big"
	"time"
	"strconv"
	"hyperchain/hyperdb/db"
	edb "hyperchain/core/db_utils"
)

type PublicContractAPI struct {
	namespace   string
	eventMux    *event.TypeMux
	pm          *manager.EventHub
	db          db.Database
	tokenBucket *ratelimit.Bucket
	config      *common.Config
}

func NewPublicContractAPI(namespace string, eventMux *event.TypeMux, pm *manager.EventHub, hyperDb db.Database, config *common.Config) *PublicContractAPI {
	fillrate, err := getFillRate(config, CONTRACT)
	if err != nil {
		log.Errorf("invalid ratelimit fill rate parameters.")
		fillrate = 10 * time.Millisecond
	}
	peak := getRateLimitPeak(config, CONTRACT)
	if peak == 0 {
		log.Errorf("got invalid ratelimit peak parameters as 0. use default peak parameters 500")
		peak = 500
	}
	return &PublicContractAPI{
		namespace:   namespace,
		eventMux:    eventMux,
		pm:          pm,
		db:          hyperDb,
		tokenBucket: ratelimit.NewBucket(fillrate, peak),
		config:      config,
	}
}

func deployOrInvoke(contract *PublicContractAPI, args SendTxArgs, txType int) (common.Hash, error) {
	var tx *types.Transaction
	realArgs, err := prepareExcute(args, txType)
	if err != nil {
		return common.Hash{}, err
	}

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), realArgs.Value.ToInt64(), payload, args.Update)

	value, err := proto.Marshal(txValue)
	if err != nil {
		return common.Hash{}, &CallbackError{err.Error()}
	}

	if args.To == nil {

		// 部署合约
		//tx = types.NewTransaction(realArgs.From[:], nil, value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp, realArgs.Nonce)

	} else {

		// 调用合约或者普通交易(普通交易还需要加检查余额)
		//tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	}

	tx.Id = uint64(contract.pm.PeerManager.GetNodeId())
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.Hash().Bytes()
	//delete repeated tx
	var exist, _ = edb.JudgeTransactionExist(contract.namespace, tx.TransactionHash)

	if exist {
		return common.Hash{}, &RepeadedTxError{"repeated tx"}
	}

	// Unsign Test
	if !tx.ValidateSign(contract.pm.AccountManager.Encryption, kec256Hash) {
		log.Error("invalid signature")
		// ATTENTION, return invalid transactino directly
		return common.Hash{}, &SignatureInvalidError{"invalid signature"}
	}

	if txBytes, err := proto.Marshal(tx); err != nil {
		log.Errorf("proto.Marshal(tx) error: %v", err)
		return common.Hash{}, &CallbackError{"proto.Marshal(tx) happened error"}
	} else if manager.GetEventObject() != nil {
		go contract.eventMux.Post(event.NewTxEvent{Payload: txBytes, Simulate: args.Simulate})
	} else {
		log.Error("manager is Nil")
		return common.Hash{}, &CallbackError{"eventObject is nil"}
	}
	return tx.GetHash(), nil

}

type CompileCode struct {
	Abi   []string `json:"abi"`
	Bin   []string `json:"bin"`
	Types []string `json:"types"`
}

// ComplieContract complies contract to ABI
func (contract *PublicContractAPI) CompileContract(ct string) (*CompileCode, error) {
	abi, bin, names, err := compiler.CompileSourcefile(ct)

	if err != nil {
		return nil, &CallbackError{err.Error()}
	}

	return &CompileCode{
		Abi:   abi,
		Bin:   bin,
		Types: names,
	}, nil
}

// DeployContract deploys contract.
func (contract *PublicContractAPI) DeployContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &SystemTooBusyError{"system is too busy to response "}
	}
	return deployOrInvoke(contract, args, 1)
}

// InvokeContract invokes contract.
func (contract *PublicContractAPI) InvokeContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &SystemTooBusyError{"system is too busy to response "}
	}
	return deployOrInvoke(contract, args, 2)
}

// GetCode returns the code from the given contract address.
func (contract *PublicContractAPI) GetCode(addr common.Address) (string, error) {

	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return "", err
	}

	return fmt.Sprintf(`0x%x`, stateDb.GetCode(addr)), nil
}

// GetContractCountByAddr returns the number of contract that has been deployed by given account address,
// if addr is nil, returns the number of all the contract that has been deployed.
func (contract *PublicContractAPI) GetContractCountByAddr(addr common.Address) (*Number, error) {

	stateDb, err := getBlockStateDb(contract.namespace, contract.config)

	if err != nil {
		return nil, err
	}

	return NewUint64ToNumber(stateDb.GetNonce(addr)), nil

}

type EncryptoArgs struct {
	Balance   Number `json:"balance"`
	Amount    Number `json:"amount"`
	HmBalance string `json:"hmBalance"`
}

type HmResult struct {
	NewBalance_hm string `json:"newBalance"`
	Amount_hm     string `json:"amount"`
}

func (contract *PublicContractAPI) EncryptoMessage(args EncryptoArgs) (*HmResult, error) {

	balance_bigint := new(big.Int)
	balance_bigint.SetInt64(args.Balance.ToInt64())

	amount_bigint := new(big.Int)
	amount_bigint.SetInt64(args.Amount.ToInt64())
	var isValid bool
	var newBalance_hm []byte
	var amount_hm []byte

	if args.HmBalance == "" {
		isValid, newBalance_hm, amount_hm = hmEncryption.PreHmTransaction(balance_bigint.Bytes(), amount_bigint.Bytes(), nil, getPaillierPublickey(contract.config))
	} else {
		hmBalance_bigint := new(big.Int)
		hmBalance_bigint.SetString(args.HmBalance, 10)
		isValid, newBalance_hm, amount_hm = hmEncryption.PreHmTransaction(balance_bigint.Bytes(), amount_bigint.Bytes(), hmBalance_bigint.Bytes(), getPaillierPublickey(contract.config))
	}

	newBalance_hm_bigint := new(big.Int)
	amount_hm_bigint := new(big.Int)

	if !isValid {
		return &HmResult{}, &OutofBalanceError{"out of balance"}
	}

	return &HmResult{
		NewBalance_hm: newBalance_hm_bigint.SetBytes(newBalance_hm).String(),
		Amount_hm:     amount_hm_bigint.SetBytes(amount_hm).String(),
	}, nil
}

type ValueArgs struct {
	RawValue   []int64  `json:"rawValue"`
	EncryValue []string `json:"encryValue"`
	Illegalhm  string   `json:"illegalhm"`
}

type HmCheckResult struct {
	CheckResult []bool  `json:"checkResult"`
	SumIllegalHmAmount string `json:"illegalHmAmount"`
}

func (contract *PublicContractAPI) CheckHmValue(args ValueArgs) (*HmCheckResult, error) {
	if len(args.RawValue) != len(args.EncryValue) {
		return nil, &InvalidParamsError{"invalid params, the length of rawValue is "+strconv.Itoa(len(args.RawValue))+", but the length of encryValue is "+strconv.Itoa(len(args.EncryValue))}
	}

	result := make([]bool, len(args.RawValue))

	illegalHmAmount_bigint := new(big.Int)
	illegalHmAmount:=make([]byte,16)
	sumIllegal := make([]byte,16)

	if(args.Illegalhm!=""){
		illegalHmAmount_bigint.SetString(args.Illegalhm, 10)
		illegalHmAmount = illegalHmAmount_bigint.Bytes()
	}
	var isvalid bool
	for i, v := range args.RawValue {
		encryVlue_bigint := new(big.Int)
		encryVlue_bigint.SetString(args.EncryValue[i], 10)

		rawValue_bigint := new(big.Int)
		rawValue_bigint.SetInt64(v)
		isvalid,sumIllegal = hmEncryption.DestinationVerify(illegalHmAmount,encryVlue_bigint.Bytes(), rawValue_bigint.Bytes(), getPaillierPublickey(contract.config))
		illegalHmAmount = sumIllegal
		result[i] = isvalid
	}

	//todo

	return &HmCheckResult{
		CheckResult: result,
		SumIllegalHmAmount: new(big.Int).SetBytes(sumIllegal).String(),
	}, nil
}

// GetStorageByAddr returns the storage by given contract address and bock number.
// The method is offered for hyperchain internal test.
func (contract *PublicContractAPI) GetStorageByAddr(addr common.Address) (map[string]string, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)

	if err != nil {
		return nil, err
	}
	mp := make(map[string]string)

	if obj := stateDb.GetAccount(addr); obj == nil {
		return nil, nil
	} else {
		cb := func(key, value common.Hash) bool {
			return true
		}
		storages := obj.ForEachStorage(cb)
		if len(storages) == 0 {
			return nil, nil
		}

		for k, v := range storages {
			mp[k.Hex()] = v.Hex()
		}
	}
	return mp, nil
}

func getBlockStateDb(namespace string, config *common.Config) (vm.Database, error) {
	//block, err := getBlockByNumber(n, db)
	stateDB, err := NewStateDb(config, namespace)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return nil, &CallbackError{err.Error()}
	}
	return stateDB, nil
}
