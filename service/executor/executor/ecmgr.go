package executor

import (
	"hyperchain/common"
	"hyperchain/core/executor"
	"hyperchain/common/service"
)

type ExecutorManager interface {

	Start() error

	Stop() error

}

type ecManagerImpl struct {

	executors map[string]executor.Executor

	// manager the connect, it can be included in the ExecutorRemote when it add
	services map[string]service.Service

	// conf is the global config file of the system, contains global configs
	// of the node
	conf *common.Config

	stopEm    chan bool
	restartEm chan bool
}

func newExecutorManager(conf *common.Config, stopEm chan bool, restartEm chan bool) *ecManagerImpl {
	em := &ecManagerImpl{
		executors:  make(map[string]executor.Executor),

		conf:        conf,
		stopEm:      stopEm,
		restartEm:   restartEm,
	}
	return em
}

func GetExecutorMgr(conf *common.Config, stopEm chan bool, restartEM chan bool) *ecManagerImpl{
	return newExecutorManager(conf, stopEm, restartEM)
}

func (em *ecManagerImpl) Start(){
	//TODO: 读配置文件,开启多个executor实例


}

func (em *ecManagerImpl) Stop(){

}