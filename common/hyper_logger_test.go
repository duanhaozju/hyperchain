//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	//"os"
	"testing"
)

func TestGetLogger(t *testing.T) {
	cp := "../configuration/configs/global/global.yaml"
	conf := NewConfig(cp)
	namespace := "gloable"
	logConfig := make(map[string]*Config)
	logConfig[namespace] = conf
	InitHyperLogger(logConfig)
	log := GetLogger(namespace, "common/hyper_logger")
	log.Notice("NOTICE")
	log.Warning("WARNING")
	log.Debug("DEBUG")
	log.Info("INFO")
}
