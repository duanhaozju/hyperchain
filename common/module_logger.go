package common

import (
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"os"
)
//TODO: Check why we need this kind of logger
type moduleLogger struct {
	compositeName  string
	logger         *logging.Logger
	level          string
	writeFile      bool
	consoleBackend logging.Backend
	fileBackend    logging.Backend
	backend        logging.LeveledBackend
}

var (
	BackendNil = errors.New("setLogLevel Error: backend nil")
)

// newModuleLogger return a moduleLogger pointer with backend binded.
func newModuleLogger(compositeName string, file *os.File,
	fileFormat string, consoleFormat string, logLevel string, writeFile bool) *moduleLogger {
	ml := &moduleLogger{
		compositeName: compositeName,
		level:         logLevel,
		writeFile:     writeFile,
	}

	logger := logging.MustGetLogger(compositeName)
	ml.logger = logger

	fileBackend := logging.NewLogBackend(file, "", 0)
	fileFormatter := logging.MustStringFormatter(fileFormat)
	fileFormatterBackend := logging.NewBackendFormatter(fileBackend, fileFormatter)
	ml.fileBackend = fileFormatterBackend

	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)
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
func (ml *moduleLogger) setLogLevel(level string) error {
	if ml.level == level {
		return nil
	}
	ml.level = level
	l, err := logging.LogLevel(level)
	if err != nil {
		return err
	}
	if ml.backend == nil {
		commonLogger.Critical("setLogLevel Error: backend nil")
		return BackendNil
	}
	ml.backend.SetLevel(l, ml.compositeName)
	return nil
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
