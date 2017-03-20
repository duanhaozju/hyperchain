//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	//"os"
	"testing"
	"fmt"
	"github.com/op/go-logging"

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

	var names []string
	var names_count int
	var config *Config

	names_count = 4
	names = getNamespaces(names_count)
	//config = NewConfig("../configuration/configs/global/global.yaml")
	config = NewConfig("../configuration/global2.yaml")


	for i := 0; i < names_count; i++ {
		if i >= len(names) {
			fmt.Println("Error: index overflow")
			return
		}
		InitHL(names[i], config)
		run(names[i])
	}

}

// run set up every namespace logger to generate logs
// according to logs, respectively.
func run(ns string) {
	fmt.Println("running ns to generate logs", ns)


	//setModuleFormat(ns)

	setModuleLevel(lv)

	generateLogs(ns)

	return
}

func getNamespaces(count int) []string {
	nss := make([]string, 4, 4)
	for i := 0; i < count; i++ {
		nss[i] = "ns_" + fmt.Sprintf("%d", i+1)
	}
	return nss
}

//func setModuleFormat(ns string) {
//
//}

func setModuleLevel(ns string, lv string) {
	//backend_stderr := logging.NewLogBackend(os.Stderr, "", 0)
	//var format_stderr = logging.MustStringFormatter(consoleFormat)
	//backendFormatter := logging.NewBackendFormatter(backend_stderr, format_stderr)
	//backendStderr := logging.AddModuleLevel(backendFormatter)
	//backendStderr.SetLevel(logDefaultLevel, "")
	//return backendStderr

	logger := GetLogger(ns, "consensus")
	bkd := logging.NewLogBackend()
	logger.SetBackend(leveledBackend)
}

func generateLogs(ns string) {
	log := GetLogger(ns, "consensus")
	log.Warning("WARNING")
	log.Notice("NOTICE")
	log.Info("INFO")
	log.Debug("DEBUG")
}