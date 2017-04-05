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

var logger = logging.MustGetLogger("commonLogger")
var hyperLoggers map[string]*HyperLogger
var rwMutex sync.RWMutex
var once sync.Once
var defaultLogLevel = "INFO"
var defaultConsoleFormat = `[%{module}]%{color}[%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message} %{color:reset}`
var defaultFileFormat = `[%{module}][%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message}`

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

// InitHyperLoggerManager init the very first hyperlogger
func InitHyperLoggerManager(conf *Config) {
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

	logger = GetLogger(DEFAULT_NAMESPACE, "common")
}

//InitHyperLogger init the whole logging system.
func InitHyperLogger(conf *Config) (*HyperLogger, error) {
	//once.Do(func() {
	//	hyperLoggers = make(map[string]*HyperLogger)
	//})
	//if !conf.ContainsKey(NAMESPACE) {
	//	conf.Set(NAMESPACE, DEFAULT_NAMESPACE)
	//}
	hyperLogger := newHyperLogger(conf)
	if hyperLogger == nil {
		return nil, errors.New("Init Hyperlogger error: nil return")
	}

	// cast it into hyperLoggers
	name := conf.GetString(NAMESPACE)
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
	var logger *logging.Logger
	if ml == nil {
		// dynamically loaded module
		hl, err := getHyperlogger(namespace)
		if err != nil {
			return nil
		}

		// add new module logger
		compositeName := getCompositeModuleName(namespace, module)

		if err != nil {
			return nil
		}
		newMl, err := hl.addNewLogger(compositeName, hl.currentFile,
			hl.fileFormat, hl.consoleFormat, hl.baseLevel, hl.writeToFile)
		if err != nil {
			return nil
		}
		logger = newMl.logger
	} else {
		logger = ml.logger
	}

	return logger
}

//SetLogLevel set log level by specific namespace module and the level provided by user.
func SetLogLevel(namespace string, module string, level string) error {
	ml := getModuleLogger(namespace, module)
	if ml == nil {
		err := errors.New("SetLogLevel Error: " + namespace + "::" + module + " not exist")
		return err
	}
	ml.setLogLevel(level)
	return nil
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
		logger.Errorf("Close Namespace Error: %s", err.Error())
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

func (hl *HyperLogger) init() {
	conf := hl.conf

	// read all configs needed
	hl.writeToFile = conf.GetBool(LOG_DUMP_FILE)

	baseLevel := conf.GetString(LOG_BASE_LOG_LEVEL)
	if baseLevel == "" {
		logger.Noticef("Invalid logging level: %s, using %s as default!", baseLevel, defaultLogLevel)
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
			logger.Critical("init error")
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
		logger.Critical("moduleLoggers nil")
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
			logger.Infof("Change log file, new log file name: %s", fileName)
			hl.currentFile.Close()
			hl.currentFile = file
		case <-hl.closeLogFile:
			logger.Info("Close logger service")
			hl.currentFile.Close()
			return
		}
	}
}

// getModuleLogger get the logger for specified module.
func getModuleLogger(namespace, module string) *moduleLogger {
	hl, err := getHyperlogger(namespace)
	if err != nil {
		logger.Critical("GetLogger error: hyperloger nil")
		return nil
	}
	compositeName := getCompositeModuleName(namespace, module)

	if hl.moduleLoggers == nil {
		logger.Critical("getLogger error: moduleLoggers nil")
		return nil
	}

	hl.moduleLoggersMutex.RLock()
	ml, ok := hl.moduleLoggers[compositeName]
	if !ok {
		logger.Debugf("module %s not exist", compositeName)
	}
	hl.moduleLoggersMutex.RUnlock()
	return ml
}

func getCompositeModuleName(namespace, module string) string {
	return namespace + "::" + module
}
