// author: chenquan
// date: 16-9-2
// last modified: 16-9-2 13:59
// last Modified Author: chenquan
// change log: 
//		
package common

import (
	"github.com/op/go-logging"
	"os"
	"time"
	"strconv"
	"log"
)

 //This applied when test, no writing log


//func InitLog(level logging.Level,loggerDir string,port int){
//
//	_, error := os.Stat(loggerDir)
//	if error == nil || os.IsExist(error){
//		//fmt.Println("directory exists")
//
//	}else {
//		//fmt.Println("no")
//		os.MkdirAll(loggerDir,0777)
//	}
//
//	backend_stderr := logging.NewLogBackend(os.Stderr, "", 0)
//	var format_stderr = logging.MustStringFormatter(
//		`%{color}[%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message}%{color:reset}`,
//	)
//	backendFormatter := logging.NewBackendFormatter(backend_stderr, format_stderr)
//	backendStderr := logging.AddModuleLevel(backendFormatter)
//	backendStderr.SetLevel(level, "")
//	logging.SetBackend(backendStderr)
//}



func InitLog(level logging.Level,loggerDir string,port int){
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	_, error := os.Stat(loggerDir)
	if error == nil || os.IsExist(error){
		//fmt.Println("directory exists")
	}else {
		//fmt.Println("no")
		os.MkdirAll(loggerDir,0777)
	}
	fileName :=loggerDir+strconv.Itoa(port) +tm.Format("-2006-01-02-15:04:05 PM")+ ".log"
	logFile,err  := os.Create(fileName)
	//defer logFile.Close()
	if err != nil {
		log.Fatalln("open file error !")
	}
	backend_stderr := logging.NewLogBackend(os.Stderr, "", 0)
	backend_file := logging.NewLogBackend(logFile, "", 0)
	var format_stderr = logging.MustStringFormatter(
		`%{color}[%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message}%{color:reset}`,
	)
	var format_file = logging.MustStringFormatter(
		`{"level":"%{level}","time":"%{time:2006-01-02 15:04:05.000}","message":"%{message}"},`,
	)
	backendFormatter := logging.NewBackendFormatter(backend_stderr, format_stderr)
	backendStderr := logging.AddModuleLevel(backendFormatter)
	backendFileFormatter := logging.NewBackendFormatter(backend_file, format_file)
	backendFile := logging.AddModuleLevel(backendFileFormatter)
	backendStderr.SetLevel(level, "")
	backendFile.SetLevel(level, "")
	logging.SetBackend(backendStderr,backendFile)
}
