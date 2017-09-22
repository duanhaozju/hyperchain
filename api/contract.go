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
	log := common.GetLogger(namespace, "api")

	var tx *types.Transaction

	// 1. verify if the parameters are valid
	realArgs, err := prepareExcute(args, txType)
	if err != nil {
		return common.Hash{}, err
	}

	// 2. create a new transaction instance
	payload := common.FromHex(realArgs.Payload)
	txValue := types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS,
		realArgs.Value.ToInt64(), payload, args.Opcode, parseVmType(realArgs.VmType))

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
		if err := tx.SetNVPHash(hash); err != nil {
			log.Errorf("set NVP hash failed! err Msg: %v.", err.Error())
			return common.Hash{}, &common.CallbackError{Message: "marshal nvp hash error"}
		}
	}
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.Hash().Bytes()

	// 3. check if there is duplicated transaction
	var exist bool
	if err, exist = edb.LookupTransaction(contract.namespace, tx.GetHash()); err != nil || exist == true {
		if exist, _ = edb.JudgeTransactionExist(contract.namespace, tx.TransactionHash); exist {
			return common.Hash{}, &common.RepeadedTxError{Message: "repeated tx " + common.ToHex(tx.TransactionHash)}
		}
	}

	// 4. verify transaction signature
	if !tx.ValidateSign(contract.eh.GetAccountManager().Encryption, kec256Hash) {
		log.Errorf("invalid signature %v", common.ToHex(tx.TransactionHash))
		return common.Hash{}, &common.SignatureInvalidError{Message: "invalid signature, tx hash " + common.ToHex(tx.TransactionHash)}
	}

	// 5. post transaction event
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
			// nvp node fails to forward tx to vp node
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

// MaintainContract maintains contract, including upgrade contract, freeze contract and unfreeze contract.
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

// TODO add annotation
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

// GetDeployedList returns all deployed contracts list (including destroyed).
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

// GetCreator returns contract creator address.
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

// GetStatus returns current contract status.
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

// GetCreateTime returns contract creation time.
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
