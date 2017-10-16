//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/evm/compiler"
	"hyperchain/crypto/hmEncryption"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type Contract struct {
	namespace   string
	eh          *manager.EventHub
	tokenBucket *ratelimit.Bucket
	config      *common.Config
}

func NewPublicContractAPI(namespace string, eh *manager.EventHub, config *common.Config) *Contract {
	log := common.GetLogger(namespace, "api")
	fillrate, err := getFillRate(namespace, config, CONTRACT)
	if err != nil {
		log.Errorf("invalid ratelimit fill rate parameters.")
		fillrate = 10 * time.Millisecond
	}
	peak := getRateLimitPeak(namespace, config, CONTRACT)
	if peak == 0 {
		log.Errorf("got invalid ratelimit peak parameters as 0. use default peak parameters 500")
		peak = 500
	}
	return &Contract{
		namespace:   namespace,
		eh:          eh,
		tokenBucket: ratelimit.NewBucket(fillrate, peak),
		config:      config,
	}
}

func deployOrInvoke(contract *Contract, args SendTxArgs, txType int, namespace string) (common.Hash, error) {
	consentor := contract.eh.GetConsentor()
	normal, full := consentor.GetStatus()
	if !normal || full {
		return common.Hash{}, &common.SystemTooBusyError{Message: "system is too busy to response "}
	}

	log := common.GetLogger(namespace, "api")

	var tx *types.Transaction
	realArgs, err := prepareExcute(args, txType)
	if err != nil {
		return common.Hash{}, err
	}

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(),
		realArgs.Value.ToInt64(), payload, realArgs.Opcode, []byte(realArgs.Extra), parseVmType(realArgs.VmType))

	value, err := proto.Marshal(txValue)
	if err != nil {
		return common.Hash{}, &common.CallbackError{Message: err.Error()}
	}

	if args.To == nil {
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp, realArgs.Nonce)

	} else {
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	}
	if contract.eh.NodeIdentification() == manager.IdentificationVP {
		tx.Id = uint64(contract.eh.GetPeerManager().GetNodeId())
	} else {
		hash := contract.eh.GetPeerManager().GetLocalNodeHash()
		err := tx.SetNVPHash(hash)
		if err != nil {
			log.Errorf("set NVP hash failed! err Msg: %v.", err.Error())
			return common.Hash{}, &common.MarshalError{Message: "marshal nvp hash error"}
		}
	}
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.Hash().Bytes()
	//delete repeated tx
	var exist bool
	if err, exist = edb.LookupTransaction(contract.namespace, tx.GetHash()); err != nil || exist == true {
		exist, _ = edb.JudgeTransactionExist(contract.namespace, tx.TransactionHash)
	}

	if exist {
		return common.Hash{}, &common.RepeadedTxError{Message: "repeated tx " + common.ToHex(tx.TransactionHash)}
	}

	// Unsign Test
	if !tx.ValidateSign(contract.eh.GetAccountManager().Encryption, kec256Hash) {
		log.Error("invalid signature")
		// ATTENTION, return invalid transactino directly
		return common.Hash{}, &common.SignatureInvalidError{Message: "invalid signature"}
	}

	if contract.eh.NodeIdentification() == manager.IdentificationNVP {
		ch := make(chan bool)
		go contract.eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
			Ch:          ch,
		})
		res := <-ch
		close(ch)
		if res == false {
			return common.Hash{}, &common.CallbackError{Message: "send tx to nvp failed."}
		}

	} else {
		go contract.eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
		})
	}
	return tx.GetHash(), nil

}

type CompileCode struct {
	Abi   []string `json:"abi"`
	Bin   []string `json:"bin"`
	Types []string `json:"types"`
}

// ComplieContract complies contract to ABI
func (contract *Contract) CompileContract(ct string) (*CompileCode, error) {
	abi, bin, names, err := compiler.CompileSourcefile(ct)

	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return &CompileCode{
		Abi:   abi,
		Bin:   bin,
		Types: names,
	}, nil
}

// DeployContract deploys contract.
func (contract *Contract) DeployContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{Message: "system is too busy to response "}
	}
	return deployOrInvoke(contract, args, 1, contract.namespace)
}

// InvokeContract invokes contract.
func (contract *Contract) InvokeContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{Message: "system is too busy to response "}
	}
	return deployOrInvoke(contract, args, 2, contract.namespace)
}

func (contract *Contract) MaintainContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{Message: "system is too busy to response "}
	}
	return deployOrInvoke(contract, args, 4, contract.namespace)
}

// GetCode returns the code from the given contract address.
func (contract *Contract) GetCode(addr common.Address) (string, error) {
	log := common.GetLogger(contract.namespace, "api")

	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return "", err
	}

	acc := stateDb.GetAccount(addr)
	if acc == nil {
		return "", &common.AccountNotExistError{Message: addr.Hex()}
	}

	return fmt.Sprintf(`0x%x`, stateDb.GetCode(addr)), nil
}

// GetContractCountByAddr returns the number of contract that has been deployed by given account address,
// if addr is nil, returns the number of all the contract that has been deployed.
func (contract *Contract) GetContractCountByAddr(addr common.Address) (*Number, error) {

	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return nil, err
	}

	acc := stateDb.GetAccount(addr)
	if acc == nil {
		return nil, &common.AccountNotExistError{Message: addr.Hex()}
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

func (contract *Contract) EncryptoMessage(args EncryptoArgs) (*HmResult, error) {

	balance_bigint := new(big.Int)
	balance_bigint.SetInt64(args.Balance.ToInt64())

	amount_bigint := new(big.Int)
	amount_bigint.SetInt64(args.Amount.ToInt64())
	var isValid bool
	var newBalance_hm []byte
	var amount_hm []byte

	if args.HmBalance == "" {
		isValid, newBalance_hm, amount_hm = hmEncryption.PreHmTransaction(balance_bigint.Bytes(),
			amount_bigint.Bytes(), nil, getPaillierPublickey(contract.config))
	} else {
		hmBalance_bigint := new(big.Int)
		hmBalance_bigint.SetString(args.HmBalance, 10)
		isValid, newBalance_hm, amount_hm = hmEncryption.PreHmTransaction(balance_bigint.Bytes(),
			amount_bigint.Bytes(), hmBalance_bigint.Bytes(), getPaillierPublickey(contract.config))
	}

	newBalance_hm_bigint := new(big.Int)
	amount_hm_bigint := new(big.Int)

	if !isValid {
		return &HmResult{}, &common.OutofBalanceError{Message: "out of balance"}
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
	CheckResult        []bool `json:"checkResult"`
	SumIllegalHmAmount string `json:"illegalHmAmount"`
}

func (contract *Contract) CheckHmValue(args ValueArgs) (*HmCheckResult, error) {
	if len(args.RawValue) != len(args.EncryValue) {
		return nil, &common.InvalidParamsError{Message: "invalid params, the length of rawValue is " +
			strconv.Itoa(len(args.RawValue)) + ", but the length of encryValue is " +
			strconv.Itoa(len(args.EncryValue))}
	}

	result := make([]bool, len(args.RawValue))

	illegalHmAmount_bigint := new(big.Int)
	illegalHmAmount := make([]byte, 16)
	sumIllegal := make([]byte, 16)

	if args.Illegalhm != "" {
		illegalHmAmount_bigint.SetString(args.Illegalhm, 10)
		illegalHmAmount = illegalHmAmount_bigint.Bytes()
	}
	var isvalid bool
	for i, v := range args.RawValue {
		encryVlue_bigint := new(big.Int)
		encryVlue_bigint.SetString(args.EncryValue[i], 10)

		rawValue_bigint := new(big.Int)
		rawValue_bigint.SetInt64(v)
		isvalid, sumIllegal = hmEncryption.DestinationVerify(illegalHmAmount, encryVlue_bigint.Bytes(),
			rawValue_bigint.Bytes(), getPaillierPublickey(contract.config))
		illegalHmAmount = sumIllegal
		result[i] = isvalid
	}
	return &HmCheckResult{
		CheckResult:        result,
		SumIllegalHmAmount: new(big.Int).SetBytes(sumIllegal).String(),
	}, nil
}

// GetStorageByAddr returns the storage by given contract address and bock number.
// The method is offered for hyperchain internal test.
func (contract *Contract) GetStorageByAddr(addr common.Address) (map[string]string, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)

	if err != nil {
		return nil, err
	}
	mp := make(map[string]string)

	if obj := stateDb.GetAccount(addr); obj == nil {
		return nil, &common.AccountNotExistError{Message: addr.Hex()}
	} else {
		cb := func(key common.Hash, value []byte) bool {
			return true
		}
		storages := obj.ForEachStorage(cb)
		if len(storages) == 0 {
			return nil, nil
		}

		for k, v := range storages {
			mp[k.Hex()] = common.Bytes2Hex(v)
		}
	}
	return mp, nil
}

// GetDeployedList return all deployed contracts list (include suicided)
func (contract *Contract) GetDeployedList(addr common.Address) ([]string, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return nil, err
	}
	if obj := stateDb.GetAccount(addr); obj == nil {
		return nil, &common.AccountNotExistError{Message: addr.Hex()}
	} else {
		return stateDb.GetDeployedContract(addr), nil
	}
}

// GetCreator return creator address
func (contract *Contract) GetCreator(addr common.Address) (common.Address, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return common.Address{}, err
	}
	if obj := stateDb.GetAccount(addr); obj == nil {
		return common.Address{}, &common.AccountNotExistError{Message: addr.Hex()}
	} else {
		if !isContractAccount(stateDb, addr) {
			return common.Address{}, nil
		}
		return stateDb.GetCreator(addr), nil
	}
}

// GetStatus return contract status
func (contract *Contract) GetStatus(addr common.Address) (string, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return "", err
	}
	if obj := stateDb.GetAccount(addr); obj == nil {
		return "", &common.AccountNotExistError{Message: addr.Hex()}
	} else {
		status := stateDb.GetStatus(addr)
		if !isContractAccount(stateDb, addr) {
			return "non-contract", nil
		}
		switch status {
		case 0:
			return "normal", nil
		case 1:
			return "frozen", nil
		default:
			return "undefined", nil
		}
	}
}

// GetCreateTime return contract status
func (contract *Contract) GetCreateTime(addr common.Address) (string, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return "", err
	}
	if obj := stateDb.GetAccount(addr); obj == nil {
		return "", &common.AccountNotExistError{Message: addr.Hex()}
	} else {
		if !isContractAccount(stateDb, addr) {
			return "", nil
		}
		blkNum := stateDb.GetCreateTime(addr)
		blk, err := edb.GetBlockByNumber(contract.namespace, blkNum)
		if err != nil {
			return "", &common.LeveldbNotFoundError{Message: "create block doesn't exist"}
		}
		return time.Unix(blk.Timestamp/1e9, blk.Timestamp%1e9).String(), nil
	}
}

// Deprecated
func (contract *Contract) GetArchive(addr common.Address, date string) (map[string]map[string]string, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return nil, err
	}

	if obj := stateDb.GetAccount(addr); obj == nil {
		return nil, &common.AccountNotExistError{Message: addr.Hex()}
	}

	return stateDb.ShowArchive(addr, date), nil
}

func getBlockStateDb(namespace string, config *common.Config) (vm.Database, error) {
	log := common.GetLogger(namespace, "api")
	stateDB, err := NewStateDb(config, namespace)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}
	return stateDB, nil
}

func isContractAccount(stateDb vm.Database, addr common.Address) bool {
	code := stateDb.GetCode(addr)
	return code != nil
}

func parseVmType(vmType string) types.TransactionValue_VmType {
	switch strings.ToLower(vmType) {
	case "jvm":
		return types.TransactionValue_JVM
	default:
		return types.TransactionValue_EVM
	}
}
