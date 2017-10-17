package executor

import (
	"hyperchain/common"
	"hyperchain/core/executor"
)

type ExecutorManager interface {

	Start() error

	Stop() error

}

type ecManagerImpl struct {

	executors map[string]executor.Executor

	// conf is the global config file of the system, contains global configs
	// of the node
	conf *common.Config

	stopHp    chan bool
	restartHp chan bool
}

func newExecutorManager(conf *common.Config, stopHp chan bool, restartHp chan bool) *ecManagerImpl {
	em := &ecManagerImpl{
		executors:  make(map[string]executor.Executor),

		conf:        conf,
		stopHp:      stopHp,
		restartHp:   restartHp,
	}
	return em
}

func GetExecutorMgr(conf *common.Config, stopHp chan bool, restartHp chan bool) *ecManagerImpl{
	return newExecutorManager(conf, stopHp,restartHp)
}

func (em *ecManagerImpl) Start(){
	//TODO: 读配置文件,开启多个executor实例
}

func (em *ecManagerImpl) Stop(){

}