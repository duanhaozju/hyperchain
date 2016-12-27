//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"github.com/op/go-logging"

	"os"
	"path"
	"strconv"
	"time"
	"github.com/spf13/cast"
)

// A logger for this file.
var loggingLogger = logging.MustGetLogger("logging")
var logDefaultLevel logging.Level

//InitLog init the log by configuration from global.yaml
func InitLog(conf *Config) {
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)

	loggerDir := conf.GetString(LOG_FILE_DIR)
	_, error := os.Stat(loggerDir)
	if !(error == nil || os.IsExist(error)) {
		os.MkdirAll(loggerDir, 0777)
	}
	lv, err := logging.LogLevel(conf.GetString(LOG_BASE_LOG_LEVEL))
	if err != nil {
		loggingLogger.Warningf("Invalid logging level: %s, using INFO as default!", conf.GetString(LOG_BASE_LOG_LEVEL))
		logDefaultLevel = logging.INFO
	} else {
		logDefaultLevel = lv
	}

	backendStderr := initLogBackend()

	if !conf.GetBool(LOG_FUMP_FILE) {
		logging.SetBackend(backendStderr)
	} else {
		fileName := path.Join(loggerDir, "hyperchain_" + strconv.Itoa(conf.GetInt(GRPC_PORT)) +
			tm.Format("-2006-01-02-15:04:05 PM") + ".log")
		setNewLogFile(fileName, backendStderr)
	}

	initLoggerLevelByConfiguration(conf)
}

//setNewLogFile set new file on disk to store logs.
func setNewLogFile(fileName string, backendStderr logging.LeveledBackend) {
	logFile, err := os.Create(fileName)

	if err != nil {
		loggingLogger.Fatalf("open file: %s error !", fileName)
	}
	logBackend := logging.NewLogBackend(logFile, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}[%{level:.5s}] %{time:15:04:05.000} %{shortfile}[%{module}] %{shortfunc} -> %{color:reset}%{message}`,
	)
	backendFileFormatter := logging.NewBackendFormatter(logBackend, format)

	lb := logging.AddModuleLevel(backendFileFormatter)
	lb.SetLevel(logDefaultLevel, "")
	logging.SetBackend(backendStderr, lb)
}

//initLogBackend init the backend info for logging.
func initLogBackend() logging.LeveledBackend {
	backend_stderr := logging.NewLogBackend(os.Stderr, "", 0)
	var format_stderr = logging.MustStringFormatter(
		`%{color}[%{level:.5s}] %{time:15:04:05.000} %{shortfile}[%{module}] %{shortfunc} -> %{color:reset}%{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(backend_stderr, format_stderr)
	backendStderr := logging.AddModuleLevel(backendFormatter)
	backendStderr.SetLevel(logDefaultLevel, "")
	return backendStderr
}

// GetModuleLogLevel gets the current logging level for the specified module
func GetModuleLogLevel(module string) (string, error) {
	// logging.GetLevel() returns the logging level for the module, if defined.
	// otherwise, it returns the default logging level
	level := logging.GetLevel(module).String()

	loggingLogger.Debugf("Module '%s' logger enabled for log level: %s", module, level)

	return level, nil
}

// SetModuleLogLevel sets the logging level for the specified module
// this his can be called from anywhere to dynamically change the log
// level for the module.
func SetModuleLogLevel(module string, logLevel string) (string, error) {
	level, err := logging.LogLevel(logLevel)
	logLevelString := level.String()

	if err != nil {
		loggingLogger.Warningf("Invalid logging level: %s - ignored", logLevel)
		return logLevelString, err
	}

	logging.SetLevel(logging.Level(level), module)
	loggingLogger.Debugf("Module '%s' logger enabled for log level: %s", module, level)
	return logLevelString, err
}

// init logger level by configuration
func initLoggerLevelByConfiguration(conf *Config) {
	mm := conf.GetStringMap(LOG_MODUKE_KEY) // module to level map
	for m, l := range mm {
		SetModuleLogLevel(m, cast.ToString(l))
	}
}
