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
	"git.hyperchain.cn/chenquan/confer/conf"
)

//hyperLoggerMgrImpl implementation of HyperLoggerMgr interface.
type hyperLoggerMgrImpl struct {
	hyperLoggers map[string]*HyperLogger
	rwMutex      sync.RWMutex
}

//newHyperLoggerMgrImpl new HyperLoggerMgr instance.
func newHyperLoggerMgrImpl(conf *Config) HyperLoggerMgr {
	hmi := &hyperLoggerMgrImpl{
		hyperLoggers: make(map[string]*HyperLogger),
	}
	conf.Set(NAMESPACE, DEFAULT_NAMESPACE)

	//TODO: remove the following code outside of this piece of code, make the code do what it really should do
	//init default HyperLogger
	hl := newHyperLogger(conf)
	hmi.addHyperLogger(hl)
	commonLogger = GetLogger(DEFAULT_NAMESPACE, "common")
	return hmi
}

//addHyperLogger add hyperLogger to HyperLoggerMgr
func (hmi *hyperLoggerMgrImpl) addHyperLogger(hl *HyperLogger)  {
	hmi.rwMutex.Lock()
	hmi.hyperLoggers[hl.conf.GetString(NAMESPACE)] = hl
	hmi.rwMutex.Unlock()
}

//GetLogger get logger with specified namespace and module.
func (hmi *hyperLoggerMgrImpl) GetLogger(namespace, module string) *logging.Logger {

	return nil
}

//SetLoggerLevel set logger level for specified namespace and module.
func (hmi *hyperLoggerMgrImpl) SetLoggerLevel(namespace, module, level string) {

	return nil
}

//GetLoggerLevel get logger level with specified namespace and module.
func (hmi *hyperLoggerMgrImpl) GetLoggerLevel(namespace, module string) string {

	return nil
}

//HyperLogger manage the logger by module for a specified namespace.
type HyperLogger struct {
	conf          *Config //config of this hyperlogger
	loggers       map[string]*moduleLogger //module name to moduleLogger map
	closeLogFile  chan struct{} //close dump log file flag channel
	rwLock        sync.RWMutex
	currentFile   *os.File //current log file

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
		conf:          conf,
		closeLogFile:  make(chan struct{}),
		loggers: make(map[string]*moduleLogger),
	}
	hl.init()
	return hl
}

func (hl *HyperLogger) init() {
	conf := hl.conf

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

	ns := conf.GetString(NAMESPACE)
	mm := conf.GetStringMap(LOG_MODULE_KEY)

	// construct module loggers according to configs
	for m, l := range mm {
		compositeName := getCompositeModuleName(ns, m)
		_, err := hl.addNewLogger(compositeName, hl.currentFile, hl.fileFormat, hl.consoleFormat,
			cast.ToString(l), hl.dumpLog)
		if err != nil {
			commonLogger.Critical("init error")
		}
	}

	// generate new log file every time interval
	if hl.dumpLog {
		go hl.newLogFileByInterval(conf)
	}
}

//newLoggerFile new logger dump file
func (hl *HyperLogger) newLoggerFile() {
	fileName := path.Join(hl.logDir, "hyperchain_"+strconv.Itoa(hl.conf.GetInt(C_GRPC_PORT))+ time.Now().Format("-2006-01-02-15:04:05 PM")+".log")
	os.MkdirAll(hl.logDir, 0777)
	file, err := os.Create(fileName)
	if err == nil {
		hl.currentFile = file
	}else {
		//TODO: we need a default log to handle this kind of error
	}
}

func (hl *HyperLogger) addNewLogger(compositeName string, file *os.File,
	fileFormat string, consoleFormat string, logLevel string, writeFile bool) (
	ml *moduleLogger, err error) {
	if hl.loggers == nil {
		err = errors.New("addNewLogger error: moduleLoggers nil")
		return nil, err
	}
	ml = newModuleLogger(compositeName, file, fileFormat, consoleFormat,
		cast.ToString(logLevel), writeFile)
	hl.rwLock.Lock()
	hl.loggers[compositeName] = ml
	hl.rwLock.Unlock()
	return ml, nil
}

//newLogFileByInterval set new log file for hyperchain
func (hl *HyperLogger) newLogFileByInterval(conf *Config) {
	loggerDir := hl.logDir
	tm := time.Now()
	hour, min, sec := 3, 0, 0
	duration := (24+hour)*3600 + min*60 + sec - (tm.Hour()*3600 + tm.Minute()*60 + tm.Second())
	// first log split at 3:00 AM
	// then split byte the interval
	d, _ := time.ParseDuration(fmt.Sprintf("%ds", duration))
	time.Sleep(d)

	fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
		time.Now().Format("-2006-01-02-15:04:05 PM")+".log")
	file, _ := os.Create(fileName)

	if hl.loggers == nil {
		commonLogger.Critical("moduleLoggers nil")
		return
	}
	hl.rwLock.RLock()
	for _, ml := range hl.loggers {
		ml.setNewLogFile(file, hl.fileFormat)
	}
	hl.rwLock.RUnlock()

	hl.currentFile.Close()
	hl.currentFile = file

	for {
		select {
		case <-time.After(conf.GetDuration(LOG_NEW_FILE_INTERVAL)):
			fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
				time.Now().Format("-2006-01-02-15:04:05 PM")+".log")
			file, _ := os.Create(fileName)
			hl.rwLock.RLock()
			for _, ml := range hl.loggers {
				ml.setNewLogFile(file, hl.fileFormat)
			}
			hl.rwLock.RUnlock()
			commonLogger.Infof("Change log file, new log file name: %s", fileName)
			hl.currentFile.Close()
			hl.currentFile = file
		case <-hl.closeLogFile:
			commonLogger.Info("Close logger service")
			hl.currentFile.Close()
			return
		}
	}
}
