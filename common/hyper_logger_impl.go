//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/spf13/cast"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

const (
	default_logger_level  = "DEBUG"
	default_logger_format = "%{color}[%{module}][%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message} %{color:reset}"
)

//TODO: make the log prefix share same length ?

// HyperLoggerMgr manage all HyperLogger for different namespaces, a namespace will be allocate a HyperLogger.
type HyperLoggerMgr interface {
	//GetLogger get logger with specified namespace and module.
	getLogger(namespace, module string) *logging.Logger

	//SetLoggerLevel set logger level for specified namespace and module.
	setLoggerLevel(namespace, module, level string) error

	//GetLoggerLevel get logger level with specified namespace and module.
	getLoggerLevel(namespace, module string) (string, error)

	//addHyperLogger add A HyperLogger to this hyperLoggerMgr.
	addHyperLogger(hp *HyperLogger)

	//getHyperLogger get HyperLogger of namespace.
	getHyperLogger(namespace string) *HyperLogger
}

// hyperLoggerMgrImpl implementation of HyperLoggerMgr interface.
type hyperLoggerMgrImpl struct {
	hyperLoggers map[string]*HyperLogger
	rwMutex      sync.RWMutex
}

// newHyperLoggerMgrImpl new HyperLoggerMgr instance.
func newHyperLoggerMgrImpl() HyperLoggerMgr {
	hmi := &hyperLoggerMgrImpl{
		hyperLoggers: make(map[string]*HyperLogger),
	}
	return hmi
}

// addHyperLogger add hyperLogger to HyperLoggerMgr.
func (hmi *hyperLoggerMgrImpl) addHyperLogger(hl *HyperLogger) {
	hmi.rwMutex.Lock()
	hmi.hyperLoggers[hl.namespace] = hl
	hmi.rwMutex.Unlock()
}

// getHyperLogger get HyperLogger of namespace.
func (hmi *hyperLoggerMgrImpl) getHyperLogger(namespace string) (hl *HyperLogger) {
	hmi.rwMutex.RLock()
	hl = hmi.hyperLoggers[namespace]
	hmi.rwMutex.RUnlock()
	return hl
}

// GetLogger get logger with specified namespace and module.
func (hmi *hyperLoggerMgrImpl) getLogger(namespace, module string) *logging.Logger {
	hl := hyperLoggerMgr.getHyperLogger(namespace)
	if hl == nil {
		commonLogger.Warningf("No hyperlogger found for namespace: %s:%s "+
			"please init namespace level logger system before try to use it,", namespace, module)
		return logging.MustGetLogger(getCompositeModuleName(namespace, module))
	}
	return hl.getOrCreateLogger(module)
}

// SetLoggerLevel set logger level for specified namespace and module.
func (hmi *hyperLoggerMgrImpl) setLoggerLevel(namespace, module, level string) error {
	hl := hyperLoggerMgr.getHyperLogger(namespace)
	if hl == nil {
		return fmt.Errorf("namespace: %s logger system is not init yet!", namespace)
	}
	ml := hl.getModuleLogger(module)
	if ml == nil {
		hyperLoggerMgr.getLogger(namespace, module)
	}
	l, err := logging.LogLevel(level)
	if err != nil {
		return err
	}
	hl.backendLock.Lock()
	hl.backend.SetLevel(l, getCompositeModuleName(namespace, module))
	hl.backendLock.Unlock()
	return nil
}

// GetLoggerLevel get logger level with specified namespace and module.
func (hmi *hyperLoggerMgrImpl) getLoggerLevel(namespace, module string) (string, error) {
	hl := hyperLoggerMgr.getHyperLogger(namespace)
	if hl == nil {
		return "", fmt.Errorf("namespace: %s logger system is not init yet!", namespace)
	}
	hl.backendLock.RLock()
	defer hl.backendLock.RUnlock()
	return hl.backend.GetLevel(getCompositeModuleName(namespace, module)).String(), nil
}

// HyperLogger manage the logger by module for a specified namespace.
type HyperLogger struct {
	conf         *Config                    //config of this hyperlogger
	loggers      map[string]*logging.Logger //module name to logger map
	closeLogFile chan struct{}              //close dump log file flag channel
	rwLock       sync.RWMutex               //read write lock for loggers

	fileLock    sync.Mutex
	currentFile *os.File //current log file

	backendLock sync.RWMutex //readwrite lock for backend
	backend     logging.LeveledBackend
	namespace   string

	baseLevel string
}

// newHyperLogger new a HyperLogger instance.
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
	baseLevel := conf.GetString(LOG_BASE_LOG_LEVEL)

	if len(baseLevel) == 0 {
		hl.baseLevel = default_logger_level
	} else {
		hl.baseLevel = baseLevel
	}

	if hl.conf.GetBool(LOG_DUMP_FILE) {
		hl.newLoggerFile()
	}
	hl.newLeveledBackEnd()

	// construct module loggers according to configs
	mm := conf.GetStringMap(LOG_MODULE_KEY)
	for m, l := range mm {
		hl.createLogger(m, cast.ToString(l))
	}
	// generate new log file every time interval
	if hl.conf.GetBool(LOG_DUMP_FILE) {
		go hl.newLogFileByInterval(conf)
		go hl.newLogFileBySize(conf)
	}
}

// newLoggerFile new logger dump file.
func (hl *HyperLogger) newLoggerFile() *os.File {
	hl.fileLock.Lock()
	preFile := hl.currentFile
	defer func() {
		if preFile != nil {
			preFile.Close()
		}
	}()
	var dir string
	if hl.namespace != DEFAULT_NAMESPACE {
		dir = GetPath(hl.namespace, hl.conf.GetString(LOG_DUMP_FILE_DIR))
	} else {
		dir = hl.conf.GetString(LOG_DUMP_FILE_DIR)
	}

	fileName := path.Join(dir, "hyperchain_"+strconv.Itoa(hl.conf.GetInt(P2P_PORT))+time.Now().Format("-2006-01-02-15:04:05 PM")+".log")
	os.MkdirAll(dir, 0777)
	file, err := os.Create(fileName)
	if err == nil {
		hl.currentFile = file
	} else {
		commonLogger.Error(err)
	}
	hl.fileLock.Unlock()
	return file
}

// newLeveledBackEnd new backend with new file.
func (hl *HyperLogger) newLeveledBackEnd() logging.LeveledBackend {
	hl.fileLock.Lock() //read file

	oldBackend := hl.backend
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)

	var consoleFormat string
	if len(hl.conf.GetString(LOG_CONSOLE_FORMAT)) == 0 {
		consoleFormat = default_logger_format
	} else {
		consoleFormat = hl.conf.GetString(LOG_CONSOLE_FORMAT)
	}
	consoleFormatter := logging.MustStringFormatter(consoleFormat)

	consoleFormatterBackend := logging.NewBackendFormatter(consoleBackend, consoleFormatter)

	hl.backendLock.Lock() // update backend
	if hl.conf.GetBool(LOG_DUMP_FILE) {
		fileBackend := logging.NewLogBackend(hl.currentFile, "", 0)
		fileFormatter := logging.MustStringFormatter(hl.conf.GetString(LOG_FILE_FORMAT))
		fileFormatterBackend := logging.NewBackendFormatter(fileBackend, fileFormatter)
		hl.backend = logging.MultiLogger(consoleFormatterBackend, fileFormatterBackend)
	} else {
		hl.backend = logging.MultiLogger(consoleFormatterBackend)
	}

	hl.rwLock.RLock() //read loggers
	if oldBackend != nil {
		for module := range hl.loggers {
			hl.backend.SetLevel(oldBackend.GetLevel(module), module)
		}
	}
	hl.rwLock.RUnlock()

	hl.backendLock.Unlock()

	hl.fileLock.Unlock()
	return hl.backend
}

// addNewLogger add new module logger for namespace logger.
func (hl *HyperLogger) addNewLogger(module string, logger *logging.Logger) error {
	hl.rwLock.Lock()
	hl.loggers[module] = logger
	hl.rwLock.Unlock()
	return nil
}

// getOrCreateLogger get logger by module, if module not existed create it.
func (hl *HyperLogger) getOrCreateLogger(module string) *logging.Logger {
	ml := hl.getModuleLogger(module)
	if ml != nil {
		return ml
	} else {
		return hl.createLogger(module, hl.baseLevel)
	}
}

// getOrCreateLogger get logger by module, if module not existed create it.
func (hl *HyperLogger) createLogger(module, lv string) *logging.Logger {
	compositeName := getCompositeModuleName(hl.namespace, module)
	logger := logging.MustGetLogger(compositeName)
	level, _ := logging.LogLevel(lv)

	hl.backendLock.Lock()
	logger.SetBackend(hl.backend)
	hl.backend.SetLevel(level, compositeName) //control level set
	hl.backendLock.Unlock()

	hl.addNewLogger(compositeName, logger)
	return logger
}

// getModuleLogger get moduleLogger by module name.
func (hl *HyperLogger) getModuleLogger(module string) *logging.Logger {
	var ml *logging.Logger
	hl.rwLock.RLock()
	ml = hl.loggers[getCompositeModuleName(hl.namespace, module)]
	hl.rwLock.RUnlock()
	return ml
}

// updateLogFileAndBackend when log file split.
func (hl *HyperLogger) updateLogFileAndBackend() {
	file := hl.newLoggerFile()
	hl.newLeveledBackEnd()
	hl.rwLock.RLock()      // read loggers
	hl.backendLock.RLock() // read backend
	for _, logger := range hl.loggers {
		logger.SetBackend(hl.backend)
	}
	hl.backendLock.RUnlock()
	hl.rwLock.RUnlock()
	commonLogger.Infof("New log file name: %s", file.Name())
}

// newLogFileByInterval set new log file by time for hyperchain
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

		if hl.conf.GetBool(LOG_DUMP_FILE) {
			hl.updateLogFileAndBackend()
		}
	}
	for {
		select {
		case <-time.After(conf.GetDuration(LOG_NEW_FILE_INTERVAL)):
			if hl.conf.GetBool(LOG_DUMP_FILE) {
				hl.updateLogFileAndBackend()
			}
		case <-hl.closeLogFile:
			hl.currentFile.Close()
			return
		}
	}
}

// newLogFileBySize set new log file by file size for hyperchain
func (hl *HyperLogger) newLogFileBySize(conf *Config) {
	for {
		select {
		case <-time.After(2 * time.Minute):
			commonLogger.Debug("check log file size")
			if hl.shouldSplitLog(conf) {
				hl.updateLogFileAndBackend()
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

func (hl *HyperLogger) shouldSplitLog(conf *Config) bool {
	const max_log_size = 199 * 1024 * 1024
	maxLogFileSize := conf.GetBytes(LOG_MAX_SIZE)
	if maxLogFileSize > max_log_size || maxLogFileSize <= 0 {
		maxLogFileSize = max_log_size
	}
	hl.fileLock.Lock()
	defer hl.fileLock.Unlock()
	fileInfo, err := os.Stat(hl.currentFile.Name())
	if err != nil {
		commonLogger.Error(err)
		return false
	}
	commonLogger.Debugf("log file size is %d ", fileInfo.Size())
	if fileInfo.Size() >= int64(maxLogFileSize) {
		return true
	}
	return false
}
