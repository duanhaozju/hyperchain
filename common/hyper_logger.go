//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"fmt"
	"github.com/op/go-logging"
	"os"
	"sync"
)

var (
	commonLogger   *logging.Logger
	once           sync.Once
	hyperLoggerMgr HyperLoggerMgr
)

func init() {
	conf := defaultLoggerConfig()
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)
	consoleFormat := conf.GetString(LOG_CONSOLE_FORMAT)
	consoleFormatter := logging.MustStringFormatter(consoleFormat)
	consoleFormatterBackend := logging.NewBackendFormatter(consoleBackend, consoleFormatter)
	logging.SetBackend(consoleFormatterBackend)
	commonLogger = logging.MustGetLogger("commonLogger")
	newHyperLoggerMgr()
}

// InitHyperLoggerManager init the hyperlogger system
func InitHyperLoggerManager(conf *Config) {
	newHyperLoggerMgr()
	//init the system log
	conf.Set(NAMESPACE, DEFAULT_NAMESPACE)
	hl := newHyperLogger(conf)
	hyperLoggerMgr.addHyperLogger(hl)
	commonLogger = GetLogger(DEFAULT_NAMESPACE, "common")
}

// newHyperLoggerMgr init the HyperLoggerMgr only once
func newHyperLoggerMgr() {
	once.Do(func() {
		if hyperLoggerMgr == nil {
			hyperLoggerMgr = newHyperLoggerMgrImpl()
		}
	})
}

// InitRawHyperLogger init hyperlogger for a namespace by default setting.
func InitRawHyperLogger(namespace string) error {
	nsConf := NewRawConfig()
	nsConf.Set(NAMESPACE, namespace)
	newHyperLoggerMgr()
	hyperLogger := newHyperLogger(nsConf)
	if hyperLogger == nil {
		return fmt.Errorf("Init Hyperlogger error: nil return")
	}
	hyperLoggerMgr.addHyperLogger(hyperLogger)
	return nil
}

// InitHyperLogger init hyperlogger for a namespace by namespace config.
func InitHyperLogger(namespace string, nsConf *Config) error {
	nsConf.Set(NAMESPACE, namespace)
	newHyperLoggerMgr()
	hyperLogger := newHyperLogger(nsConf)
	if hyperLogger == nil {
		return fmt.Errorf("Init Hyperlogger error: nil return")
	}
	hyperLoggerMgr.addHyperLogger(hyperLogger)
	return nil
}

// GetLogger getLogger with specific namespace and module.
func GetLogger(namespace string, module string) *logging.Logger {
	if hyperLoggerMgr == nil {
		commonLogger.Warningf("Hyper logger sytem is not init, init it by default configs")
		InitHyperLoggerManager(defaultLoggerConfig())
	}
	return hyperLoggerMgr.getLogger(namespace, module)
}

// SetLogLevel set log level by specific namespace module and the level provided by user.
func SetLogLevel(namespace string, module string, level string) error {
	return hyperLoggerMgr.setLoggerLevel(namespace, module, level)
}

// GetLogLevel get log level info by namespace and module.
func GetLogLevel(namespace, module string) (string, error) {
	return hyperLoggerMgr.getLoggerLevel(namespace, module)
}

// defaultLoggerConfig return a default config
func defaultLoggerConfig() *Config {
	conf := NewRawConfig()
	conf.Set(LOG_BASE_LOG_LEVEL, default_logger_level)
	conf.Set(LOG_DUMP_FILE, false)
	conf.Set(LOG_CONSOLE_FORMAT, "%{color}[%{module}][%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message} %{color:reset}")
	return conf
}
