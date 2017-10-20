package api

import (
	"hyperchain/core/executor"
	"hyperchain/api"
)

type ExecutorApi struct {
	ec			*executor.Executor
	namespace	string
	block		*api.Block
}

func NewExecutorApi(ec *executor.Executor, ns string)  *ExecutorApi{
	ea := &ExecutorApi{
		ec:			ec,
		namespace:	ns,
	}

	ea.block = api.NewPublicBlockAPI(ns)

	return ea
}

func(ea *ExecutorApi) Validate ()  {
	//ea.ec.Validate()
}

func (ea *ExecutorApi) GetChainHeight()  {
	ea.block.GetChainHeight()
}




