//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"github.com/juju/ratelimit"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/manager"
	"strings"
	"time"
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

// DeployContract deploys contract.
func (contract *Contract) DeployContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}
	return contract.postContract(args, 1)
}

// InvokeContract invokes contract.
func (contract *Contract) InvokeContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}
	return contract.postContract(args, 2)
}

// MaintainContract maintains contract, including upgrade contract, freeze contract and unfreeze contract.
func (contract *Contract) MaintainContract(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(contract.config) && contract.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}
	return contract.postContract(args, 4)
}

// postContract will create a new transaction instance and post a NewTxEvent event.
func (contract *Contract) postContract(args SendTxArgs, txType int) (common.Hash, error) {
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
	err = postNewTxEvent(args, tx, contract.eh)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.GetHash(), nil

}

func parseVmType(vmType string) types.TransactionValue_VmType {
	switch strings.ToLower(vmType) {
	case "jvm":
		return types.TransactionValue_JVM
	default:
		return types.TransactionValue_EVM
	}
}
