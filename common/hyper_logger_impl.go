//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"sync"
	"os"
	"path"
	"time"
	"github.com/spf13/cast"
	"fmt"
	"errors"
	"strconv"
	"github.com/op/go-logging"
)

//hyperLoggerMgrImpl implementation of HyperLoggerMgr interface.
type hyperLoggerMgrImpl struct {
	hyperLoggers map[string]*HyperLogger
	rwMutex sync.RWMutex
}

func (hmi *hyperLoggerMgrImpl) newHyperLoggerMgrImpl()  {
	
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
	conf               *Config
	moduleLoggers      map[string]*moduleLogger
	closeLogFile       chan struct{}
	moduleLoggersMutex sync.RWMutex
	writeToFile        bool
	currentFile        *os.File
	baseLevel          string
	fileFormat         string
	consoleFormat      string
	logDir             string
}

func (hl *HyperLogger) init() {
	conf := hl.conf

	// read all configs needed
	hl.writeToFile = conf.GetBool(LOG_DUMP_FILE)

	baseLevel := conf.GetString(LOG_BASE_LOG_LEVEL)
	if baseLevel == "" {
		commonLogger.Noticef("Invalid logging level: %s, using %s as default!", baseLevel, defaultLogLevel)
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

	ns := conf.GetString(NAMESPACE)
	mm := conf.GetStringMap(LOG_MODULE_KEY)

	// construct module loggers according to configs
	for m, l := range mm {
		compositeName := getCompositeModuleName(ns, m)
		_, err := hl.addNewLogger(compositeName, file, fileFormat, consoleFormat,
			cast.ToString(l), hl.writeToFile)
		if err != nil {
			commonLogger.Critical("init error")
		}
	}

	// generate new log file every time interval
	if hl.writeToFile {
		go hl.newLogFileByInterval(loggerDir, conf)
	}
}

func (hl *HyperLogger) addNewLogger(compositeName string, file *os.File,
fileFormat string, consoleFormat string, logLevel string, writeFile bool) (
ml *moduleLogger, err error) {
	if hl.moduleLoggers == nil {
		err = errors.New("addNewLogger error: moduleLoggers nil")
		return nil, err
	}
	ml = newModuleLogger(compositeName, file, fileFormat, consoleFormat,
		cast.ToString(logLevel), writeFile)
	hl.moduleLoggersMutex.Lock()
	hl.moduleLoggers[compositeName] = ml
	hl.moduleLoggersMutex.Unlock()
	return ml, nil
}

//newLogFileByInterval set new log file for hyperchain
func (hl *HyperLogger) newLogFileByInterval(loggerDir string, conf *Config) {
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

	if hl.moduleLoggers == nil {
		commonLogger.Critical("moduleLoggers nil")
		return
	}
	hl.moduleLoggersMutex.RLock()
	for _, ml := range hl.moduleLoggers {
		ml.setNewLogFile(file, hl.fileFormat)
	}
	hl.moduleLoggersMutex.RUnlock()

	hl.currentFile.Close()
	hl.currentFile = file

	for {
		select {
		case <-time.After(conf.GetDuration(LOG_NEW_FILE_INTERVAL)):
			fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
				time.Now().Format("-2006-01-02-15:04:05 PM")+".log")
			file, _ := os.Create(fileName)
			hl.moduleLoggersMutex.RLock()
			for _, ml := range hl.moduleLoggers {
				ml.setNewLogFile(file, hl.fileFormat)
			}
			hl.moduleLoggersMutex.RUnlock()
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
