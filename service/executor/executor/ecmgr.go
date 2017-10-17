package executor

import (
	"hyperchain/common"
	"hyperchain/core/executor"
    "io/ioutil"
    "errors"
    "github.com/op/go-logging"
)

var logger *logging.Logger

const (
    DEFAULT_NAMESPACE  = "global"
    NS_CONFIG_DIR_ROOT = "namespace.config_root_dir"
)

type ExecutorManager interface {

	Start() error

	Stop() error

}

type ecManagerImpl struct {

	executors map[string]*executor.Executor

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
    logger = common.GetLogger(common.DEFAULT_LOG, "nsmgr")

    return newExecutorManager(conf, stopHp,restartHp)
}


func (em *ecManagerImpl) Start() error {
    //TODO: 读配置文件,开启多个executor实例
    configRootDir := em.conf.GetString(NS_CONFIG_DIR_ROOT)
    if configRootDir == "" {
        return errors.New("Namespace config root dir is not valid ")
    }
    dirs, err := ioutil.ReadDir(configRootDir)
    if err != nil {
        return err
    }
    for _, d := range dirs {
        if d.IsDir() {
            name := d.Name()
            start := em.conf.GetBool(common.START_NAMESPACE + name)
            if !start {
                continue
            }
            e, err := executor.NewExecutor(name, em.conf, nil, nil)
            if err != nil {
                logger.Errorf("NewExecutor is fault")
            }
            em.executors[name] = e
        } else {
            logger.Errorf("Invalid folder %v", d)
        }
    }
    return nil
}

func (em *ecManagerImpl) Stop() error {
    return nil
}