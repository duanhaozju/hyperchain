//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"github.com/op/go-logging"
	"fmt"
	"sync"
)

var(
	commonLogger = logging.MustGetLogger("commonLogger")
	once sync.Once
	hyperLoggerMgr HyperLoggerMgr
)

func init() {
	newHyperLoggerMgr()
}

// InitHyperLoggerManager init the hyperlogger system
func InitHyperLoggerManager(conf *Config) {
		newHyperLoggerMgr()
		//init the system log
		conf.Set(NAMESPACE, DEFAULT_NAMESPACE)
		hl := newHyperLogger(conf)
		hyperLoggerMgr.addHyperLogger(hl)
		commonLogger = GetLogger(DEFAULT_NAMESPACE, "common")
}

func newHyperLoggerMgr()  {
	once.Do(func() {
		if hyperLoggerMgr == nil {
			hyperLoggerMgr = newHyperLoggerMgrImpl()
		}
	})
}

//InitRawHyperLogger init hyperlogger for a namespace by default setting.
func InitRawHyperLogger(namespace string) error {
	nsConf := NewEmptyConfig()
	nsConf.Set(NAMESPACE, namespace)
	newHyperLoggerMgrImpl()
	hyperLogger := newHyperLogger(nsConf)
	if hyperLogger == nil {
		return fmt.Errorf("Init Hyperlogger error: nil return")
	}
	hyperLoggerMgr.addHyperLogger(hyperLogger)
	return nil
}

//InitHyperLogger init hyperlogger for a namespace by namespace config.
func InitHyperLogger(namespace string, nsConf *Config) error {
	nsConf.Set(NAMESPACE, namespace)
	newHyperLoggerMgrImpl()
	hyperLogger := newHyperLogger(nsConf)
	if hyperLogger == nil {
		return fmt.Errorf("Init Hyperlogger error: nil return")
	}
	hyperLoggerMgr.addHyperLogger(hyperLogger)
	return nil
}

//GetLogger getLogger with specific namespace and module.
func GetLogger(namespace string, module string) *logging.Logger {
	if hyperLoggerMgr == nil {
		panic(fmt.Sprintf("Hyper logger sytem is not init, please init it first"))
	}
	return hyperLoggerMgr.getLogger(namespace, module)
}

//SetLogLevel set log level by specific namespace module and the level provided by user.
func SetLogLevel(namespace string, module string, level string) error {
	return hyperLoggerMgr.setLoggerLevel(namespace, module, level)
}

//GetLogLevel get log level info by namespace and module.
func GetLogLevel(namespace, module string) (string, error) {
	return hyperLoggerMgr.getLoggerLevel(namespace, module)
}