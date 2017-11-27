package api

import (
	"fmt"
	edb "github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/juju/ratelimit"
	"time"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/vm"
	capi "github.com/hyperchain/hyperchain/api"
)


// This file implements the handler of Contract service API which
// can be invoked by client in JSON-RPC request.

type Contract struct {
	namespace   string
	tokenBucket *ratelimit.Bucket
	config      *common.Config
}

// NewPublicContractAPI creates and returns a new Contract instance for given namespace name.
func NewPublicContractAPI(namespace string, config *common.Config) *Contract {
	log := common.GetLogger(namespace, "api")
	fillRate, err := capi.GetFillRate(namespace, config, capi.CONTRACT)
	if err != nil {
		log.Errorf("invalid ratelimit fill rate parameters.")
		fillRate = 10 * time.Millisecond
	}
	peak := capi.GetRateLimitPeak(namespace, config, capi.CONTRACT)
	if peak == 0 {
		log.Errorf("got invalid ratelimit peak parameters as 0. use default peak parameters 500")
		peak = 500
	}
	return &Contract {
		namespace:   namespace,
		tokenBucket: ratelimit.NewBucket(fillRate, peak),
		config:      config,
	}
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
		return "", &common.AccountNotExistError{Address: addr.Hex()}
	}

	return fmt.Sprintf(`0x%x`, stateDb.GetCode(addr)), nil
}

// GetContractCountByAddr returns the number of contract that has been deployed by given account address.
// If account doesn't exist, error will be returned.
func (contract *Contract) GetContractCountByAddr(addr common.Address) (*capi.Number, error) {

	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return nil, err
	}

	acc := stateDb.GetAccount(addr)
	if acc == nil {
		return nil, &common.AccountNotExistError{Address: addr.Hex()}
	}

	return capi.Uint64ToNumber(stateDb.GetNonce(addr)), nil

}

// GetDeployedList returns all deployed contracts list (including destroyed).
func (contract *Contract) GetDeployedList(addr common.Address) ([]string, error) {
	stateDb, err := getBlockStateDb(contract.namespace, contract.config)
	if err != nil {
		return nil, err
	}
	if obj := stateDb.GetAccount(addr); obj == nil {
		return nil, &common.AccountNotExistError{Address: addr.Hex()}
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
		return common.Address{}, &common.AccountNotExistError{Address: addr.Hex()}
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
		return "", &common.AccountNotExistError{Address: addr.Hex()}
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
		return "", &common.AccountNotExistError{Address: addr.Hex()}
	} else {
		if !isContractAccount(stateDb, addr) {
			return "", nil
		}
		blkNum := stateDb.GetCreateTime(addr)
		blk, err := edb.GetBlockByNumber(contract.namespace, blkNum)
		if err != nil {
			return "", &common.DBNotFoundError{Type: capi.BLOCK, Id: fmt.Sprintf("number %#x", blkNum)}
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
