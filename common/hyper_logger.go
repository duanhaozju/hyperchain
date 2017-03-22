//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"github.com/op/go-logging"

	"fmt"
	"github.com/spf13/cast"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
	"github.com/pkg/errors"
	"strings"
)

var logger = logging.MustGetLogger("commonLogger")
var hyperLoggers map[string]*HyperLogger
var rwMutex sync.RWMutex
var once sync.Once

var defaultLogLevel = "INFO"
var defaultConsoleFormat = `[%{module}]%{color}[%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message} %{color:reset}`
var defaultFileFormat = `[%{module}][%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message}`

type HyperLogger struct {
	conf   		 *Config
	loggers          map[string]*logging.Logger  // todo: multilogger
	//logBackendFile 	 *logging.LeveledBackend
	//logBackendCons   *logging.LeveledBackend
	//fileFormat       string
	//consoleFormat    string
	fileBackend	 logging.LeveledBackend
	consoleBackend   logging.LeveledBackend
	writeToFile      bool
	closeLogFile 	 chan struct{}
	baseLevel	 logging.Level
}

//InitHyperLogger int the whole logging system.
func InitHyperLogger(conf *Config) (*HyperLogger, error) {
	once.Do(func() {
		hyperLoggers = make(map[string]*HyperLogger)
	})

	// first time init in main no namespace exist
	// so set it
	if !conf.ContainsKey(NAMESPACE) {
		conf.Set(NAMESPACE, "system")
	}
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

	return getLogger(namespace, module)
}

func SetLogLevel(namespace string, module string, level string) (string, error) {
	hl, err := getHyperlogger(namespace)
	if err != nil {
		err := errors.New("Error: " + namespace + " not exist cannot get hyperlogger")
		return "", err
	}
	hl.setModuleLogLevel(namespace, module, level)
	return "", nil
}

func Close(namespace string) error {
	hl, err := getHyperlogger(namespace)
	if err != nil {
		logger.Errorf("Close Namespace Error: %s", err.Error())
		return err
	}
	hl.closeLogFile <- struct {}{}
	return nil
}

func newHyperLogger(conf *Config) *HyperLogger {
	hl := &HyperLogger{
		conf:    conf,
		closeLogFile: make(chan struct{}),
		loggers: make(map[string]*logging.Logger),
	}
	return hl
}

func (hl *HyperLogger) init() {
	conf := hl.conf

	// whether need write to file
	hl.writeToFile = conf.GetBool(LOG_FUMP_FILE)

	// console fortmat backend
	consoleFmt := conf.GetString(LOG_CONSOLE_FORMAT)
	backendStderr := hl.initConsoleBackend(consoleFmt)
	hl.consoleBackend = backendStderr

	// file format backend
	fileFormat = conf.GetString(LOG_FILE_FORMAT)
	loggerDir := conf.GetString(LOG_FILE_DIR)
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
		tm.Format("-2006-01-02-15:04:05 PM")+".log")
	os.MkdirAll(loggerDir, 0777)
	fileBackend := hl.initFileBackend(fileName, fileFormat)
	hl.fileBackend = fileBackend

	baseLevel, err := logging.LogLevel(conf.GetString(LOG_BASE_LOG_LEVEL))
	if err != nil {
		logger.Noticef("Invalid logging level: %s, using %s as default!", baseLevel, defaultLogLevel)
		hl.baseLevel, _ = logging.LogLevel(defaultLogLevel)
	} else {
		hl.baseLevel = baseLevel
	}

	hl.initLoggerLevelByConfiguration(conf)

	if hl.writeToFile {
		go hl.newLogFileByInterval(loggerDir, conf) //split log by second
	}
}

func getCompositeModuleName(namespace, module string) string {
	return namespace + "::" + module
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
// init logger level by configuration
func (hl *HyperLogger) initLoggerLevelByConfiguration(conf *Config) {
	ns := conf.GetString(NAMESPACE)
	mm := conf.GetStringMap(LOG_MODULE_KEY) // module to level map
	for m, l := range mm {
		hl.setModuleLogLevel(ns, m, cast.ToString(l))
	}
}

// setModuleLogLevel sets the logging level for the specified module.
func (hl *HyperLogger) setModuleLogLevel(name string, module string, logLevel string) (string, error) {

	compositeModuleName := getCompositeModuleName(name, module)
	level, err := logging.LogLevel(logLevel)
	logLevelString := level.String()
	if err != nil {
		logger.Warningf("Invalid logging level: %s - ignored", logLevel)
		return logLevelString, err
	}

	hl.consoleBackend.SetLevel(level, compositeModuleName)
	hl.fileBackend.SetLevel(level, compositeModuleName)

	// get logger register if not exist
	_, ok := hl.loggers[compositeModuleName]
	if !ok {
		newLogger := logging.MustGetLogger(compositeModuleName)
		hl.loggers[compositeModuleName] = newLogger
	}
	logger := hl.loggers[compositeModuleName]

	if hl.writeToFile {
		multiBackend := logging.MultiLogger(hl.fileBackend, hl.consoleBackend)
		logger.SetBackend(multiBackend)
	} else {
		logger.SetBackend(hl.consoleBackend)
	}

	logger.Infof("Module '%s' logger enabled for log level: %s", compositeModuleName, level)
	return logLevelString, err
}

// getLogger get the logger for specified module, new one if not exist
func getLogger(namespace, module string) *logging.Logger {
	hl, err := getHyperlogger(namespace)
	if err != nil {
		logger.Errorf("GetLogger error: %", err.Error())
		return nil
	}

	compositeModuleName := getCompositeModuleName(namespace, module)
	logger.Infof("init log module:%s", compositeModuleName)
	if _, ok := hl.loggers[compositeModuleName]; !ok {
		logger := logging.MustGetLogger(compositeModuleName)
		hl.loggers[compositeModuleName] = logger
		hl.setModuleLogLevel(namespace, module, hl.baseLevel.String())
		return logger
	} else {
		return hl.loggers[compositeModuleName]
	}

}

//initConsoleBackend init the console backend info for logging.
func (hl *HyperLogger) initConsoleBackend(consoleFormat string) logging.LeveledBackend {
	var consoleFormatter logging.Formatter
	if strings.EqualFold(consoleFormat, "") {
		logger.Notice("consoleFormat not set use default")
		consoleFormatter = logging.MustStringFormatter(defaultConsoleFormat)
	} else {
		consoleFormatter = logging.MustStringFormatter(consoleFormat)
	}
	consoleBackend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(consoleBackend, consoleFormatter)
	leveledConsoleBackend := logging.AddModuleLevel(backendFormatter)
	return leveledConsoleBackend
}

//initFileBackend init the file backend for logging.
func (hl *HyperLogger) initFileBackend(fileName string, fileFormat string) logging.LeveledBackend {
	var fileFormatter logging.Formatter
	if strings.EqualFold(fileFormat, "") {
		logger.Notice("fileFormat not set use default")
		fileFormatter = logging.MustStringFormatter(defaultFileFormat)
	} else {
		fileFormatter = logging.MustStringFormatter(fileFormat)
	}
	logFile, err := os.Create(fileName)
	if err != nil {
		logger.Fatalf("open file: %s error !", fileName)
	}
	fileBackend := logging.NewLogBackend(logFile, "", 0)
	fileBackendFormatter := logging.NewBackendFormatter(fileBackend, fileFormatter)
	leveledFileBackend := logging.AddModuleLevel(fileBackendFormatter)
	return leveledFileBackend
}


//setNewLogFile set new file on disk to store logs.
func (hl *HyperLogger) setNewLogFile(fileName string, backendStderr logging.LeveledBackend) {
	logFile, err := os.Create(fileName)

	if err != nil {
		logger.Fatalf("open file: %s error !", fileName)
	}
	logBackend := logging.NewLogBackend(logFile, "", 0)
	var format = logging.MustStringFormatter(fileFormat)
	backendFileFormatter := logging.NewBackendFormatter(logBackend, format)

	lb := logging.AddModuleLevel(backendFileFormatter)
	lb.SetLevel(logDefaultLevel, "")
	// todo: can set dump file log level here...

	logging.SetBackend(backendStderr, lb)
}

//newLogFileByInterval set new log file for hyperchain
func (hl *HyperLogger) newLogFileByInterval(loggerDir string, conf *Config) {
	tm := time.Now()
	hour, min, sec := 3, 0, 0
	duration := (24+hour)*3600+min*60+sec - (tm.Hour()*3600+tm.Minute()*60+tm.Second())
	// first log split at 3:00 AM
	// then split byte the interval
	d, _ := time.ParseDuration(fmt.Sprintf("%ds", duration))
	time.Sleep(d)
	timestamp := time.Now().Unix()
	tm = time.Unix(timestamp, 0)
	fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
		tm.Format("-2006-01-02-15:04:05PM")+".log")
	//setNewLogFile(fileName, backendStderr)
	fileFormat = conf.GetString(LOG_FILE_FORMAT)
	hl.fileBackend = hl.initFileBackend(fileName, fileFormat)
	hl.initLoggerLevelByConfiguration(conf)

	loop: for {
		select {
		case <-time.After(conf.GetDuration(LOG_NEW_FILE_INTERVAL)):
			timestamp := time.Now().Unix()
			tm := time.Unix(timestamp, 0)
			fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
				tm.Format("-2006-01-02-15:04:05 PM")+".log")
			fileFormat = conf.GetString(LOG_FILE_FORMAT)
			hl.fileBackend = hl.initFileBackend(fileName, fileFormat)
			hl.initLoggerLevelByConfiguration(conf)
			logger.Infof("Change log file, new log file name: %s", fileName)
		case <-hl.closeLogFile:
			logger.Info("Close logger service")
			break loop
		}
	}
}
