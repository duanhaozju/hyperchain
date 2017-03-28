package common

import (
	"github.com/op/go-logging"
	"os"
	"github.com/pkg/errors"
)

type moduleLogger struct {
	compositeName	string
	logger		*logging.Logger
	level		string
	writeFile	bool
	consoleBackend	logging.Backend
	fileBackend	logging.Backend
	backend		logging.LeveledBackend
}

// newModuleLogger return a moduleLogger pointer with backend binded.
func newModuleLogger(compositeName string, file *os.File,
fileFormat string, consoleFormat string, logLevel string, writeFile bool) *moduleLogger {
	ml := &moduleLogger{
		compositeName:	compositeName,
		level:		logLevel,
		writeFile:	writeFile,
	}

	logger := logging.MustGetLogger(compositeName)
	ml.logger = logger

	fileBackend := logging.NewLogBackend(file, "", 0)
	fileFormatter := logging.MustStringFormatter(fileFormat)
	fileFormatterBackend := logging.NewBackendFormatter(fileBackend, fileFormatter)
	ml.fileBackend = fileFormatterBackend

	consoleBackend := logging.NewLogBackend(os.Stderr, "", 0)
	consoleFormatter := logging.MustStringFormatter(consoleFormat)
	consoleFormatterBackend := logging.NewBackendFormatter(consoleBackend, consoleFormatter)
	ml.consoleBackend = consoleFormatterBackend

	// bind logger and backend
	if writeFile {
		ml.backend = logging.MultiLogger(ml.fileBackend, ml.consoleBackend)
	} else {
		ml.backend = logging.MultiLogger(ml.consoleBackend)
	}
	level, _ := logging.LogLevel(logLevel)
	ml.backend.SetLevel(level, compositeName)
	ml.logger.SetBackend(ml.backend)

	return ml
}

// setLogLevel set both file log and console log to given level.
func (ml *moduleLogger) setLogLevel(level string) {
	if ml.level == level {
		return
	}

	ml.level = level
	l, _ := logging.LogLevel(level)
	if ml.backend == nil {
		logger.Critical("setLogLevel Error: backend nil")
		return
	}
	ml.backend.SetLevel(l, ml.compositeName)
}

// getLogLevel get level string of current module.
func (ml *moduleLogger) getLogLevel() string {
	return ml.level
}

// setFileBackend set file backend with given fileName.
func (ml *moduleLogger) setNewLogFile(file *os.File, fileFormat string) error {
	if file == nil {
		return errors.New("setNewLogFile Error: file nil")
	}
	fileBackend := logging.NewLogBackend(file, "", 0)
	fileFormatter := logging.MustStringFormatter(fileFormat)
	fileFormatterBackend := logging.NewBackendFormatter(fileBackend, fileFormatter)
	ml.fileBackend = fileFormatterBackend

	ml.backend = logging.MultiLogger(ml.consoleBackend, ml.fileBackend)
	level, _ := logging.LogLevel(ml.level)
	ml.backend.SetLevel(level, ml.compositeName)
	ml.logger.SetBackend(ml.backend)
	return nil
}