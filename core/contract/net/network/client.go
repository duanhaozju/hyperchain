//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package network

import (
	//"golang.org/x/net/context"
	pb "hyperchain/core/contract/protos"
)

type ContractExecutor interface {
	//Execute execute the contract cmd.
	Execute(tx *Transaction) (interface{}, error)
	//Start start the contract executor.
	Start() error
	//Stop stop the contract executor.
	Stop() error
}

type contractExecutorImpl struct {
	client pb.ContractClient
}

func (cei *contractExecutorImpl) Executor(tx *Transaction) (interface{}, error) {
	//TODO: implement the executor logic
	return nil, nil
}

func (cei *contractExecutorImpl) Start() error {
	//TODO: start contract executor
	return nil
}

func (cei *contractExecutorImpl) Stop() error {
	//TODO: stop contract executor
	return nil
}

type Transaction struct {
	Msg string
}
