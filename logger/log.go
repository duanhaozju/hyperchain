package myLogger

import (
	"os"
	"log"
	"strconv"
	"time"
)
//const logFileDir  = "/tmp/hyperchain/cache/logs/"
const logFileDir  = "./logs/"
var logger *log.Logger

func NewLogger(port int) *log.Logger{
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)

	_, error := os.Stat(logFileDir)
	if error == nil || os.IsExist(error){
		//fmt.Println("directory exists")

	}else {
		//fmt.Println("no")
		os.MkdirAll(logFileDir,0777)
	}

	fileName :=logFileDir+strconv.Itoa(port) +tm.Format("-2006-01-02-15:04:05 PM")+ ".log"
	logFile,err  := os.Create(fileName)
	//defer logFile.Close()
	if err != nil {
		log.Fatalln("open file error !")
	}
	logger = log.New(logFile,"[INFO] ",log.Lshortfile)
	return logger
}

func GetLogger() *log.Logger {
	return logger
}