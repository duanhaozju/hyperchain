//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"github.com/spf13/cast"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

var (
	ErrNoBackend = errors.New("common/hyper_logger_impl: backend is nil")
)

//HyperLoggerMgr manage all HyperLogger for different namespaces, a namespace will be allocate a HyperLogger.
type HyperLoggerMgr interface {
	//GetLogger get logger with specified namespace and module.
	getLogger(namespace, module string) *logging.Logger

	//SetLoggerLevel set logger level for specified namespace and module.
	setLoggerLevel(namespace, module, level string) error

	//GetLoggerLevel get logger level with specified namespace and module.
	getLoggerLevel(namespace, module string) (string, error)

	//addHyperLogger add A HyperLogger to this hyperLoggerMgr
	addHyperLogger(hp *HyperLogger)

	//getHyperLogger get HyperLogger of namespace
	getHyperLogger(namespace string) *HyperLogger
}

//hyperLoggerMgrImpl implementation of HyperLoggerMgr interface.
type hyperLoggerMgrImpl struct {
	hyperLoggers map[string]*HyperLogger
	rwMutex      sync.RWMutex
}

//newHyperLoggerMgrImpl new HyperLoggerMgr instance.
func newHyperLoggerMgrImpl() HyperLoggerMgr {
	hmi := &hyperLoggerMgrImpl{
		hyperLoggers: make(map[string]*HyperLogger),
	}
	return hmi
}

//addHyperLogger add hyperLogger to HyperLoggerMgr
func (hmi *hyperLoggerMgrImpl) addHyperLogger(hl *HyperLogger) {
	hmi.rwMutex.Lock()
	hmi.hyperLoggers[hl.namespace] = hl
	hmi.rwMutex.Unlock()
}

//getHyperLogger get HyperLogger of namespace
func (hmi *hyperLoggerMgrImpl) getHyperLogger(namespace string) (hl *HyperLogger) {
	hmi.rwMutex.RLock()
	hl = hmi.hyperLoggers[namespace]
	hmi.rwMutex.RUnlock()
	return hl
}

//GetLogger get logger with specified namespace and module.
func (hmi *hyperLoggerMgrImpl) getLogger(namespace, module string) *logging.Logger {
	hl := hyperLoggerMgr.getHyperLogger(namespace)
	if hl == nil {
		commonLogger.Errorf("No hyperlogger found for namespace: %s "+
			"please init namespace level logger system before try to use logger in it,", namespace)
		return nil
	}
	return hl.getOrCreateLogger(module)
}

//SetLoggerLevel set logger level for specified namespace and module.
func (hmi *hyperLoggerMgrImpl) setLoggerLevel(namespace, module, level string) error {
	hl := hyperLoggerMgr.getHyperLogger(namespace)
	if hl == nil {
		return fmt.Errorf("namespace: %s logger system is not init yet!", namespace)
	}
	l, err := logging.LogLevel(level)
	if err != nil {
		return err
	}
	hl.backend.SetLevel(l, getCompositeModuleName(namespace, module))
	return nil
}

//GetLoggerLevel get logger level with specified namespace and module.
func (hmi *hyperLoggerMgrImpl) getLoggerLevel(namespace, module string) (string, error) {
	hl := hyperLoggerMgr.getHyperLogger(namespace)
	if hl == nil {
		return "", fmt.Errorf("namespace: %s logger system is not init yet!", namespace)
	}
	ml := hl.getModuleLogger(module)
	if ml == nil {
		return "", fmt.Errorf("SetLogLevel Error: %s::%s not exist", namespace, module)
	}
	return hl.backend.GetLevel(getCompositeModuleName(namespace, module)).String(), nil
	//return ml.getLogLevel(), nil
}

//HyperLogger manage the logger by module for a specified namespace.
type HyperLogger struct {
	conf         *Config                    //config of this hyperlogger
	loggers      map[string]*logging.Logger //module name to logger map
	closeLogFile chan struct{}              //close dump log file flag channel
	rwLock       sync.RWMutex
	currentFile  *os.File //current log file
	backend      logging.LeveledBackend
	namespace    string

	//TODO: Remove the following variables
	dumpLog       bool
	baseLevel     string
	fileFormat    string
	consoleFormat string
	logDir        string
}

//newHyperLogger new a HyperLogger instance.
func newHyperLogger(conf *Config) *HyperLogger {
	hl := &HyperLogger{
		conf:         conf,
		closeLogFile: make(chan struct{}),
		loggers:      make(map[string]*logging.Logger),
	}
	hl.init()
	return hl
}

func (hl *HyperLogger) init() {
	conf := hl.conf
	ns := conf.GetString(NAMESPACE)
	hl.namespace = ns
	// read all configs needed
	hl.dumpLog = conf.GetBool(LOG_DUMP_FILE)
	baseLevel := conf.GetString(LOG_BASE_LOG_LEVEL)

	if len(baseLevel) == 0 {
		hl.baseLevel = defaultLogLevel
	} else {
		hl.baseLevel = baseLevel
	}

	hl.fileFormat = conf.GetString(LOG_FILE_FORMAT)
	hl.consoleFormat = conf.GetString(LOG_CONSOLE_FORMAT)
	hl.logDir = conf.GetString(LOG_FILE_DIR)

	if hl.dumpLog {
		hl.newLoggerFile()
	}
	hl.newLeveledBackEnd()

	// construct module loggers according to configs
	mm := conf.GetStringMap(LOG_MODULE_KEY)
	for m, l := range mm {
		hl.createLogger(m, cast.ToString(l))
	}
	// generate new log file every time interval
	if hl.dumpLog {
		go hl.newLogFileByInterval(conf)
	}
}

//newLoggerFile new logger dump file
func (hl *HyperLogger) newLoggerFile() *os.File {
	preFile := hl.currentFile
	defer func() {
		if preFile != nil {
			preFile.Close()
		}
	}()

	fileName := path.Join(hl.logDir, "hyperchain_"+strconv.Itoa(hl.conf.GetInt(C_GRPC_PORT))+time.Now().Format("-2006-01-02-15:04:05 PM")+".log")
	os.MkdirAll(hl.logDir, 0777)
	file, err := os.Create(fileName)
	if err == nil {
		hl.currentFile = file
	}else {
		commonLogger.Error(err)
	}
	return file
}

//newLeveledBackEnd new backend with new file.
func (hl *HyperLogger) newLeveledBackEnd() logging.LeveledBackend {
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)
	consoleFormatter := logging.MustStringFormatter(hl.consoleFormat)
	consoleFormatterBackend := logging.NewBackendFormatter(consoleBackend, consoleFormatter)

	if hl.dumpLog {
		fmt.Println(hl.currentFile.Name())
		fileBackend := logging.NewLogBackend(hl.currentFile, "", 0)
		fileFormatter := logging.MustStringFormatter(hl.fileFormat)
		fileFormatterBackend := logging.NewBackendFormatter(fileBackend, fileFormatter)
		hl.backend = logging.MultiLogger(consoleFormatterBackend, fileFormatterBackend)
	} else {
		hl.backend = logging.MultiLogger(consoleFormatterBackend)
	}
	return hl.backend
}

//addNewLogger add new module logger for namespace logger
func (hl *HyperLogger) addNewLogger(module string, logger *logging.Logger) error {
	hl.rwLock.Lock()
	hl.loggers[module] = logger
	hl.rwLock.Unlock()
	return nil
}

//getOrCreateLogger get logger by module, if module not existed create it.
func (hl *HyperLogger) getOrCreateLogger(module string) *logging.Logger {
	ml := hl.getModuleLogger(module)
	if ml != nil {
		return ml
	} else {
		return hl.createLogger(module, hl.baseLevel)
	}
}

//getOrCreateLogger get logger by module, if module not existed create it.
func (hl *HyperLogger) createLogger(module, lv string) *logging.Logger {
	compositeName := getCompositeModuleName(hl.namespace, module)
	logger := logging.MustGetLogger(compositeName)
	level, _ := logging.LogLevel(lv)
	hl.backend.SetLevel(level, compositeName)
	logger.SetBackend(hl.backend)
	hl.addNewLogger(compositeName, logger)
	return logger
}

//getModuleLogger get moduleLogger by module name.
func (hl *HyperLogger) getModuleLogger(module string) *logging.Logger {
	var ml *logging.Logger
	hl.rwLock.RLock()
	ml = hl.loggers[getCompositeModuleName(hl.namespace, module)]
	hl.rwLock.RUnlock()
	return ml
}

//newLogFileByInterval set new log file for hyperchain
func (hl *HyperLogger) newLogFileByInterval(conf *Config) {
	du := conf.GetDuration(LOG_NEW_FILE_INTERVAL)
	if du.Hours() == 24 {
		tm := time.Now()
		hour, min, sec := 3, 0, 0
		duration := (24+hour)*3600 + min*60 + sec - (tm.Hour()*3600 + tm.Minute()*60 + tm.Second())
		// first log split at 3:00 AM
		// then split byte the interval
		d, _ := time.ParseDuration(fmt.Sprintf("%ds", duration))
		time.Sleep(d)

		if hl.dumpLog {
			hl.newLoggerFile()
			hl.newLeveledBackEnd()

			hl.rwLock.RLock()
			for _, logger := range hl.loggers {
				logger.SetBackend(hl.backend)
			}
			hl.rwLock.RUnlock()
		}
	}

	for {
		select {
		case <-time.After(conf.GetDuration(LOG_NEW_FILE_INTERVAL)):
			if hl.dumpLog {
				file := hl.newLoggerFile()
				hl.newLeveledBackEnd()
				hl.rwLock.RLock()
				for _, logger := range hl.loggers {
					logger.SetBackend(hl.backend)
				}
				hl.rwLock.RUnlock()

				//TODO: how to fix commmonLooger problem
				commonLogger.Infof("Split log file, new log file name: %s", file.Name())
			}
		case <-hl.closeLogFile:
			hl.currentFile.Close()
			return
		}
	}
}

func getCompositeModuleName(namespace, module string) string {
	return namespace + "::" + module
}
