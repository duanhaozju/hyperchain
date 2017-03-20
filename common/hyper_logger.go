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
)

var logger = logging.MustGetLogger("commmon")

var hyperLogger *HyperLogger
var once sync.Once

//InitHyperLogger int the whole logging system.
func InitHyperLogger(conf *Config) {
	once.Do(func() {
	hyperLogger = newHyperLogger(conf)
	})
}

type HyperLogger struct {
	conf    *Config
	close   chan bool
	loggers map[string]*logging.Logger
}

func newHyperLogger(conf *Config) *HyperLogger {
	hl := &HyperLogger{
		conf:    conf,
		close:   make(chan bool),
		loggers: make(map[string]*logging.Logger),
	}
	hl.init()
	return hl
}

//GetLogger getLogger with specific namespace and module.
func GetLogger(namespace, module string) *logging.Logger {
	compositeModuleName := getCompositeModuleName(namespace, module)
	logger.Infof("init log module:%s", compositeModuleName)
	if hyperLogger == nil {
		fmt.Println("null hyperLogger")
	}
	if _, ok := hyperLogger.loggers[compositeModuleName]; !ok {
		logger := logging.MustGetLogger(compositeModuleName)
		hyperLogger.loggers[compositeModuleName] = logger
		return logger
	} else {
		return hyperLogger.loggers[compositeModuleName]
	}
}

func (hl *HyperLogger) init() {
	conf := hl.conf
	consoleFormat = conf.GetString(LOG_CONSOLE_FORMAT)
	fileFormat = conf.GetString(LOG_FILE_FORMAT)

	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)

	loggerDir := conf.GetString(LOG_FILE_DIR) //TODO: Keep logs in different dirs by different namespace?

	_, error := os.Stat(loggerDir)
	if !(error == nil || os.IsExist(error)) {
		os.MkdirAll(loggerDir, 0777)
	}
	lv, err := logging.LogLevel(conf.GetString(LOG_BASE_LOG_LEVEL))

	if err != nil {
		commonLogger.Noticef("Invalid logging level: %s, using INFO as default!", conf.GetString(LOG_BASE_LOG_LEVEL))
		logDefaultLevel = logging.INFO
	} else {
		logDefaultLevel = lv
	}

	backendStderr := hl.initLogBackend()

	if !conf.GetBool(LOG_FUMP_FILE) {
		logging.SetBackend(backendStderr)
	} else {
		closeLogFile = make(chan struct{})
		fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
			tm.Format("-2006-01-02-15:04:05 PM")+".log")
		hl.setNewLogFile(fileName, backendStderr)
		go hl.newLogFileByInterval(loggerDir, conf, backendStderr) //split log by sencond
	}

	hl.initLoggerLevelByConfiguration(conf)
}

func getCompositeModuleName(namespace, module string) string {
	return namespace + "/" + module
}

// init logger level by configuration
func (hl *HyperLogger) initLoggerLevelByConfiguration(conf *Config) {
	mm := conf.GetStringMap(LOG_MODULE_KEY) // module to level map
	for m, l := range mm {
		hl.SetModuleLogLevel(getCompositeModuleName("", m), cast.ToString(l))
	}
}

// SetModuleLogLevel sets the logging level for the specified module
// this his can be called from anywhere to dynamically change the log
// level for the module.
func (hl *HyperLogger) SetModuleLogLevel(module string, logLevel string) (string, error) {
	level, err := logging.LogLevel(logLevel)
	logLevelString := level.String()

	if err != nil {
		commonLogger.Warningf("Invalid logging level: %s - ignored", logLevel)
		return logLevelString, err
	}

	logging.SetLevel(logging.Level(level), module)
	commonLogger.Infof("Module '%s' logger enabled for log level: %s", module, level)
	return logLevelString, err
}

//initLogBackend init the backend info for logging.
func (hl *HyperLogger) initLogBackend() logging.LeveledBackend {
	backend_stderr := logging.NewLogBackend(os.Stderr, "", 0)
	var format_stderr = logging.MustStringFormatter(consoleFormat)
	backendFormatter := logging.NewBackendFormatter(backend_stderr, format_stderr)
	backendStderr := logging.AddModuleLevel(backendFormatter)
	backendStderr.SetLevel(logDefaultLevel, "")
	return backendStderr
}

//setNewLogFile set new file on disk to store logs.
func (hl *HyperLogger) setNewLogFile(fileName string, backendStderr logging.LeveledBackend) {
	logFile, err := os.Create(fileName)

	if err != nil {
		commonLogger.Fatalf("open file: %s error !", fileName)
	}
	logBackend := logging.NewLogBackend(logFile, "", 0)
	var format = logging.MustStringFormatter(fileFormat)
	backendFileFormatter := logging.NewBackendFormatter(logBackend, format)

	lb := logging.AddModuleLevel(backendFileFormatter)
	lb.SetLevel(logDefaultLevel, "")
	logging.SetBackend(backendStderr, lb)
}

//newLogFileByInterval set new log file for hyperchain
func (hl *HyperLogger) newLogFileByInterval(loggerDir string, conf *Config, backendStderr logging.LeveledBackend) {
	tm := time.Now()
	sec := (24+3-tm.Hour())*3600 - tm.Minute()*60 - tm.Second()
	// first log split at 3:00 AM
	// then split byte the interval
	d, _ := time.ParseDuration(fmt.Sprintf("%ds", sec))
	time.Sleep(d)
	timestamp := time.Now().Unix()
	tm = time.Unix(timestamp, 0)
	fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
		tm.Format("-2006-01-02-15:04:05PM")+".log")
	setNewLogFile(fileName, backendStderr)

	for {
		select {
		case <-time.After(conf.GetDuration(LOG_NEW_FILE_INTERVAL)):
			timestamp := time.Now().Unix()
			tm := time.Unix(timestamp, 0)
			fileName := path.Join(loggerDir, "hyperchain_"+strconv.Itoa(conf.GetInt(C_GRPC_PORT))+
				tm.Format("-2006-01-02-15:04:05 PM")+".log")
			setNewLogFile(fileName, backendStderr)
			commonLogger.Infof("Change log file, new log file name: %s", fileName)
		case <-closeLogFile:
			commonLogger.Info("Close logger service")
			commonLogger.SetBackend(backendStderr)
		}
	}
}