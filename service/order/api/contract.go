//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/core/vm/evm/compiler"
	"github.com/hyperchain/hyperchain/crypto/hmEncryption"
	"github.com/hyperchain/hyperchain/manager"
	"github.com/juju/ratelimit"
	"time"
	"math/big"
	"strings"
	capi "github.com/hyperchain/hyperchain/api"
	"strconv"
)

// This file implements the handler of Contract service API which
// can be invoked by client in JSON-RPC request.

type Contract struct {
	namespace   string
	eh          *manager.EventHub
	tokenBucket *ratelimit.Bucket
	config      *common.Config
}

// NewPublicContractAPI creates and returns a new Contract instance for given namespace name.
func NewPublicContractAPI(namespace string, eh *manager.EventHub, config *common.Config) *Contract {
	log := common.GetLogger(namespace, "api")
	fillrate, err := capi.GetFillRate(namespace, config, capi.CONTRACT)
	if err != nil {
		log.Errorf("invalid ratelimit fill rate parameters.")
		fillrate = 10 * time.Millisecond
	}
	peak := capi.GetRateLimitPeak(namespace, config, capi.CONTRACT)
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

// DeployContract deploys contract.
func (contract *Contract) DeployContract(args SendTxArgs) (common.Hash, error) {
	if capi.GetRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}
	return contract.postContract(args, 1)
}

// InvokeContract invokes contract.
func (contract *Contract) InvokeContract(args SendTxArgs) (common.Hash, error) {
	if capi.GetRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}
	return contract.postContract(args, 2)
}

// MaintainContract maintains contract, including upgrade contract, freeze contract and unfreeze contract.
func (contract *Contract) MaintainContract(args SendTxArgs) (common.Hash, error) {
	if capi.GetRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}
	return contract.postContract(args, 4)
}

// postContract will create a new transaction instance and post a NewTxEvent event.
func (contract *Contract) postContract(args SendTxArgs, txType int) (common.Hash, error) {
	log := common.GetLogger(contract.namespace, "api")

	consentor := contract.eh.GetConsentor()
	normal, full := consentor.GetStatus()
	if !normal || full {
		return common.Hash{}, &common.SystemTooBusyError{}
	}

	// 1. create a new transaction instance
	tx, err := prepareTransaction(args, txType, contract.namespace, contract.eh)
	if err != nil {
		return common.Hash{}, err
	}

	// 2. post a event.NewTxEvent event
	log.Debugf("[ %v ] post event NewTxEvent", contract.namespace)
	err = postNewTxEvent(args, tx, contract.eh)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.GetHash(), nil

}

type CompileCode struct {
	Abi   []string `json:"abi"`
	Bin   []string `json:"bin"`
	Types []string `json:"types"`
}

// ComplieContract complies solidity contract. It will return abi, bin and contract name.
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

// EncryptoArgs specifies parameters for Contract.EncryptoMessage function call.
type EncryptoArgs struct {

	// The balance(plain text) of account A before transferring money to account B.
	Balance capi.Number `json:"balance"`

	// The amount(plain text) that account A will transfer to account B.
	Amount capi.Number `json:"amount"`

	// InvalidHmValue is optional. It represents invalid homomorphic encryption transaction amount of account A
	// when a person transfers money(amount homomorphic encryption) to A that can't be verified by account A.
	//
	// For example, account C transfers 10 dollars to account A, but this transaction fails to pass validation
	// of account A. Therefore, account A saves 10 value encrypted as invalid homomorphic encryption value.
	InvalidHmValue string `json:"invalidHmValue"`
}

type HmResult struct {

	// The homomorphic sum of the homomorphic encryption balance of account A after transferring money to
	// account B (balance - amount) and invalid homomorphic value of account A.
	NewBalance_hm string `json:"newBalance"`

	// The amount(homomorphic encryption text) that account A will transfer to account B.
	Amount_hm string `json:"amount"`
}

// EncryptoMessage encrypts data by homomorphic encryption.
func (contract *Contract) EncryptoMessage(args EncryptoArgs) (*HmResult, error) {

	balance_bigint := new(big.Int)
	balance_bigint.SetInt64(args.Balance.Int64())

	amount_bigint := new(big.Int)
	amount_bigint.SetInt64(args.Amount.Int64())
	var isValid bool
	var newBalance_hm []byte
	var amount_hm []byte

	if args.InvalidHmValue == "" {
		isValid, newBalance_hm, amount_hm = hmEncryption.PreHmTransaction(balance_bigint.Bytes(),
			amount_bigint.Bytes(), nil, capi.GetPaillierPublickey(contract.config))
	} else {
		hmBalance_bigint := new(big.Int)
		hmBalance_bigint.SetString(args.InvalidHmValue, 10)
		isValid, newBalance_hm, amount_hm = hmEncryption.PreHmTransaction(balance_bigint.Bytes(),
			amount_bigint.Bytes(), hmBalance_bigint.Bytes(), capi.GetPaillierPublickey(contract.config))
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

// CheckArgs specifies parameters for Contract.CheckHmValue function call.
type CheckArgs struct {

	// All unverified transaction amount list (plain text).
	// For example, account A transfers 10 dollars to B twice, RawValue is [10,10].
	RawValue []int64 `json:"rawValue"`

	// All unverified transaction amount list (homomorphic encryption).
	// For example, account A transfers 10 dollars to B twice, EncryValue is [(encrypted 10), (encrypted 10)].
	EncryValue []string `json:"encryValue"`

	// Invalid homomorphic encryption value of account B.
	InvalidHmValue string `json:"invalidHmValue"`
}

type HmCheckResult struct {

	// Data validation results.
	// For example, account A transfers 10 dollars to B twice, but the first transfer validation fails,
	// CheckResult is [false, true]
	CheckResult []bool `json:"checkResult"`

	// The homomorphic sum of all invalid homomorphic encryption transaction amount of B.
	SumIllegalHmAmount string `json:"illegalHmAmount"`
}

// CheckHmValue returns verification result that account B verifies transaction amount that account A transfers to B.
func (contract *Contract) CheckHmValue(args CheckArgs) (*HmCheckResult, error) {
	if len(args.RawValue) != len(args.EncryValue) {
		return nil, &common.InvalidParamsError{Message: "invalid params, the length of rawValue is " +
			strconv.Itoa(len(args.RawValue)) + ", but the length of encryValue is " +
			strconv.Itoa(len(args.EncryValue))}
	}

	result := make([]bool, len(args.RawValue))

	illegalHmAmount_bigint := new(big.Int)
	illegalHmAmount := make([]byte, 16)
	sumIllegal := make([]byte, 16)

	if args.InvalidHmValue != "" {
		illegalHmAmount_bigint.SetString(args.InvalidHmValue, 10)
		illegalHmAmount = illegalHmAmount_bigint.Bytes()
	}
	var isvalid bool
	for i, v := range args.RawValue {
		encryVlue_bigint := new(big.Int)
		encryVlue_bigint.SetString(args.EncryValue[i], 10)

		rawValue_bigint := new(big.Int)
		rawValue_bigint.SetInt64(v)
		isvalid, sumIllegal = hmEncryption.DestinationVerify(illegalHmAmount, encryVlue_bigint.Bytes(),
			rawValue_bigint.Bytes(), capi.GetPaillierPublickey(contract.config))
		illegalHmAmount = sumIllegal
		result[i] = isvalid
	}
	return &HmCheckResult{
		CheckResult:        result,
		SumIllegalHmAmount: new(big.Int).SetBytes(sumIllegal).String(),
	}, nil
}

func parseVmType(vmType string) types.TransactionValue_VmType {
	switch strings.ToLower(vmType) {
	case "jvm":
		return types.TransactionValue_JVM
	default:
		return types.TransactionValue_EVM
	}
}
