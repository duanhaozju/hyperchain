//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Usage:
// Use GetLogger to get a logger
// Use SetLogLevel to set the logger's level

var commonLogger = logging.MustGetLogger("commonLogger")
var hyperLoggers map[string]*HyperLogger
var rwMutex sync.RWMutex
var once sync.Once
var defaultLogLevel = "INFO"


//new code start here


type HyperLoggerMgr interface {
	//GetLogger get logger with specified namespace and module.
	GetLogger(namespace, module string) *logging.Logger

	//SetLoggerLevel set logger level for specified namespace and module.
	SetLoggerLevel(namespace, module, level string)

	//GetLoggerLevel get logger level with specified namespace and module.
	GetLoggerLevel(namespace, module string) string
}

// InitHyperLoggerManager init the very first hyperlogger
func InitHyperLoggerManager(conf *Config) {

	//new code start here


	/////////old code
	once.Do(func() {
		hyperLoggers = make(map[string]*HyperLogger)
	})

	conf.Set(NAMESPACE, DEFAULT_NAMESPACE)
	hl := newHyperLogger(conf)
	// read all configs needed
	hl.writeToFile = conf.GetBool(LOG_DUMP_FILE)

	baseLevel := conf.GetString(LOG_BASE_LOG_LEVEL)
	if baseLevel == "" {
		hl.baseLevel = defaultLogLevel
	} else {
		hl.baseLevel = baseLevel
	}

	fileFormat := conf.GetString(LOG_FILE_FORMAT)
	hl.fileFormat = fileFormat
	consoleFormat := conf.GetString(LOG_CONSOLE_FORMAT)
	hl.consoleFormat = consoleFormat
	loggerDir := conf.GetString(LOG_FILE_DIR)
	hl.logDir = loggerDir

	fileName := path.Join(loggerDir,
		"hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
			time.Now().Format("-2006-01-02-15:04:05 PM")+".log")
	os.MkdirAll(loggerDir, 0777)
	file, _ := os.Create(fileName)
	hl.currentFile = file

	name := conf.GetString(NAMESPACE)
	rwMutex.Lock()
	hyperLoggers[name] = hl
	rwMutex.Unlock()

	commonLogger = GetLogger(DEFAULT_NAMESPACE, "common")
}

//InitHyperLogger init hyperlogger for a namespace by namespace config.
func InitHyperLogger(nsConf *Config) (*HyperLogger, error) {
	hyperLogger := newHyperLogger(nsConf)
	if hyperLogger == nil {
		return nil, errors.New("Init Hyperlogger error: nil return")
	}

	// cast it into hyperLoggers
	name := nsConf.GetString(NAMESPACE)
	if strings.EqualFold(name, "") {
		return nil, errors.New("Init Hyperlogger error: nil namespace")
	}

	hyperLogger.init()
	rwMutex.Lock()
	hyperLoggers[name] = hyperLogger
	rwMutex.Unlock()

	return hyperLogger, nil
}

//GetLogger getLogger with specific namespace and module.
func GetLogger(namespace string, module string) *logging.Logger {
	ml := getModuleLogger(namespace, module)
	var tmpLogger *logging.Logger
	if ml == nil {
		// dynamically loaded module
		hl, err := getHyperlogger(namespace)
		if err != nil {
			commonLogger.Error(err)
			commonLogger.Errorf("%s namespace logger not initialized using common logger instead!")
			return commonLogger
		}

		// add new module logger
		compositeName := getCompositeModuleName(namespace, module)

		newMl, err := hl.addNewLogger(compositeName, hl.currentFile,
			hl.fileFormat, hl.consoleFormat, hl.baseLevel, hl.writeToFile)
		if err != nil {
			commonLogger.Error(err)
			commonLogger.Errorf("add new logger failed using common logger instead!")
			return commonLogger
		}
		tmpLogger = newMl.logger
	} else {
		tmpLogger = ml.logger
	}

	return tmpLogger
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

func CloseHyperlogger(namespace string) error {
	hl, err := getHyperlogger(namespace)
	if err != nil {
		commonLogger.Errorf("Close Namespace Error: %s", err.Error())
		return err
	}
	hl.closeLogFile <- struct{}{}
	rwMutex.Lock()
	delete(hyperLoggers, namespace)
	rwMutex.Unlock()

	return nil
}

func newHyperLogger(conf *Config) *HyperLogger {
	hl := &HyperLogger{
		conf:          conf,
		closeLogFile:  make(chan struct{}),
		moduleLoggers: make(map[string]*moduleLogger),
	}
	return hl
}

// getHyperlogger use RLock to read hyperloggers map
func getHyperlogger(namespace string) (*HyperLogger, error) {
	if hyperLoggers == nil {
		return nil, errors.New("getHyperlogger error: Hyperloggers nil")
	}

	var err error
	rwMutex.RLock()
	hl, ok := hyperLoggers[namespace]
	if !ok {
		err = errors.New("getHyperlogger error: namespace not exist")
	} else {
		err = nil
	}
	rwMutex.RUnlock()
	return hl, err
}

// getModuleLogger get the logger for specified module.
func getModuleLogger(namespace, module string) *moduleLogger {
	hl, err := getHyperlogger(namespace)
	if err != nil {
		commonLogger.Critical("GetLogger error: hyperloger nil")
		return nil
	}
	compositeName := getCompositeModuleName(namespace, module)

	if hl.moduleLoggers == nil {
		commonLogger.Critical("getLogger error: moduleLoggers nil")
		return nil
	}

	hl.moduleLoggersMutex.RLock()
	ml, ok := hl.moduleLoggers[compositeName]
	if !ok {
		commonLogger.Debugf("module %s not exist", compositeName)
	}
	hl.moduleLoggersMutex.RUnlock()
	return ml
}

func getCompositeModuleName(namespace, module string) string {
	return namespace + "::" + module
}
