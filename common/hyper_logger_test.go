//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	//"os"
	"testing"

	"fmt"
)

func TestGetLogger(t *testing.T) {
	conf := NewConfig("../configuration/namespaces/global/config/global.yaml")
	InitNsHyperLogger(conf)
	log := GetLogger("global", "consensus")
	SetLogLevel("global", "consensus", "NOTICE")
	fmt.Println(log)
	fmt.Println("aaa")
	log.Notice("+++++++NOTICE")
	log.Warning("WARNING")
	log.Debug("DEBUG")
	log.Info("INFO")

	SetLogLevel("global", "consensus", "DEBUG")
	log.Notice("+++++++NOTICE")
	log.Warning("WARNING")
	log.Debug("DEBUG")
	log.Info("INFO")
}

func TestLoggerInit(t *testing.T)  {



}
