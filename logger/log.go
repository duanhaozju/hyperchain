package logger

import (
	"os"
	"log"
	"strconv"
)

var O *log.Logger

func NewLogger(port int){
	fileName := strconv.Itoa(port) + ".log"
	logFile,err  := os.Create(fileName)
	//defer logFile.Close()
	if err != nil {
		log.Fatalln("open file error !")
	}
	O = log.New(logFile,"[INFO] ",log.Lshortfile)
}

