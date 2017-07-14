//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"fmt"
	"sync"
)

// Usage:
// Use GetLogger to get a logger
// Use SetLogLevel to set the logger's level

var commonLogger = logging.MustGetLogger("commonLogger")
var once sync.Once
var defaultLogLevel = "INFO"

//new code start here
var hyperLoggerMgr HyperLoggerMgr

type HyperLoggerMgr interface {
	//GetLogger get logger with specified namespace and module.
	GetLogger(namespace, module string) *logging.Logger

	//SetLoggerLevel set logger level for specified namespace and module.
	SetLoggerLevel(namespace, module, level string)

	//GetLoggerLevel get logger level with specified namespace and module.
	GetLoggerLevel(namespace, module string) string

	//addHyperLogger add A HyperLogger to this hyperLoggerMgr
	addHyperLogger(hp *HyperLogger)

	//getHyperLogger get HyperLogger of namespace
	getHyperLogger(namespace string) *HyperLogger
}

// InitHyperLoggerManager init the hyperlogger system
func InitHyperLoggerManager(conf *Config) {
	once.Do(func() {
		hyperLoggerMgr = newHyperLoggerMgrImpl(conf)
	})
}

//InitHyperLogger init hyperlogger for a namespace by namespace config.
func InitHyperLogger(nsConf *Config) (error) {
	hyperLogger := newHyperLogger(nsConf)
	if hyperLogger == nil {
		return fmt.Errorf("Init Hyperlogger error: nil return")
	}
	hyperLoggerMgr.addHyperLogger(hyperLogger)
	return nil
}

//GetLogger getLogger with specific namespace and module.
func GetLogger(namespace string, module string) *logging.Logger {
	ml := getModuleLogger(namespace, module)
	if ml == nil {
		// dynamically loaded module
		hl := hyperLoggerMgr.getHyperLogger(namespace)
		if hl == nil {
			//TODO: fix it
			commonLogger.Error(fmt.Errorf("No Namespace Logger found for %s using commonLogger instead", namespace))
			return commonLogger
		}
		// add new module logger
		compositeName := getCompositeModuleName(namespace, module)
		ml = newModuleLogger(compositeName, hl.currentFile, hl.fileFormat, hl.consoleFormat, hl.baseLevel, hl.dumpLog)

		err := hl.addNewLogger(ml)
		if err != nil {
			commonLogger.Error(fmt.Errorf("New logger for %s failed, using commonLogger instead", namespace))
			return commonLogger
		}
	}
	return ml.logger
}

//SetLogLevel set log level by specific namespace module and the level provided by user.
func SetLogLevel(namespace string, module string, level string) error {
	ml := getModuleLogger(namespace, module)
	if ml == nil {
		err := errors.New("SetLogLevel Error: " + namespace + "::" + module + " not exist")
		return err
	}
	return ml.setLogLevel(level)
}

//GetLogLevel get log level info by namespace and module.
func GetLogLevel(namespace, module string) (string, error) {
	ml := getModuleLogger(namespace, module)
	if ml == nil {
		err := errors.New("GetLogLevel Error: " + namespace + "::" + module + " not exist")
		return "", err
	}
	return ml.level, nil
}

//TODO: Fix it
//func CloseHyperlogger(namespace string) error {
//	hl, err := getHyperlogger(namespace)
//	if err != nil {
//		commonLogger.Errorf("Close Namespace Error: %s", err.Error())
//		return err
//	}
//	hl.closeLogFile <- struct{}{}
//
//	//TODO: Close logger by namespace
//
//	//rwMutex.Lock()
//	//delete(hyperLoggers, namespace)
//	//rwMutex.Unlock()
//
//	return nil
//}

// getModuleLogger get the logger for specified module.
func getModuleLogger(namespace, module string) *moduleLogger {

	hl := hyperLoggerMgr.getHyperLogger(namespace)

	if hl == nil {
		commonLogger.Criticalf("No hyperlogger found for namespace: %s", namespace)
		return nil
	}
	compositeName := getCompositeModuleName(namespace, module)


	hl.rwLock.RLock()
	ml, ok := hl.loggers[compositeName]
	if !ok {
		commonLogger.Debugf("module %s not exist", compositeName)
	}
	hl.rwLock.RUnlock()
	return ml
}

func getCompositeModuleName(namespace, module string) string {
	return namespace + "::" + module
}
