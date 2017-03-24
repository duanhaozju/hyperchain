//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	//"os"
	"testing"

	"fmt"
)

func TestGetLogger(t *testing.T) {
	//cp := "../configuration/configs/global/global.yaml"
	//conf := NewConfig(cp)
	//namespace := "gloable"
	//logConfig := make(map[string]*Config)
	//logConfig[namespace] = conf
	//InitHyperLogger(logConfig)
	//log := GetLogger(namespace, "common/hyper_logger")
	//log.Notice("NOTICE")
	//log.Warning("WARNING")
	//log.Debug("DEBUG")
	//log.Info("INFO")

	//config = NewConfig("../configuration/configs/global/global.yaml")

	conf := NewConfig("../configuration/namespaces/global/config/global.yaml")
	InitHyperLogger(conf)
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



