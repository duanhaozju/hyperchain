//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"github.com/op/go-logging"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

func InitLog(level logging.Level, loggerDir string, gRPCport int, dumpFileFlag bool) {
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	_, error := os.Stat(loggerDir)
	if error == nil || os.IsExist(error) {
		//fmt.Println("directory exists")
	} else {
		//fmt.Println("no")
		os.MkdirAll(loggerDir, 0777)
	}

	backend_stderr := logging.NewLogBackend(os.Stderr, "", 0)
	var format_stderr = logging.MustStringFormatter(
		`%{color}[%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message}%{color:reset}`,
	)
	backendFormatter := logging.NewBackendFormatter(backend_stderr, format_stderr)
	backendStderr := logging.AddModuleLevel(backendFormatter)
	backendStderr.SetLevel(level, "")

	if !dumpFileFlag {
		logging.SetBackend(backendStderr)
	} else {
		fileName := path.Join(loggerDir, strconv.Itoa(gRPCport)+tm.Format("-2006-01-02-15:04:05 PM")+".log")
		logFile, err := os.Create(fileName)
		//defer logFile.Close()
		if err != nil {
			log.Fatalln("open file error !")
		}
		backend_file := logging.NewLogBackend(logFile, "", 0)
		var format_file = logging.MustStringFormatter(
			`{"level":"%{level}","time":"%{time:2006-01-02 15:04:05.000}","message":"%{message}"},`,
		)
		backendFileFormatter := logging.NewBackendFormatter(backend_file, format_file)
		backendFile := logging.AddModuleLevel(backendFileFormatter)
		backendFile.SetLevel(level, "")
		logging.SetBackend(backendStderr, backendFile)

	}
}
