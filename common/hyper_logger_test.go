//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"testing"
	"sync"
)

func TestGetLogLevel(t *testing.T) {
	conf := NewConfig("../configuration/namespaces/global/config/global.yaml")
	conf.Set(LOG_DUMP_FILE, false)
	namespace := "namespace1"
	InitHyperLogger(namespace, conf)

	err := SetLogLevel(namespace, "p2p", "debug")
	if err != nil {
		t.Error(err)
	}
	level, err := GetLogLevel(namespace, "p2p")
	if err != nil {
		 if level != "debug" {
			 t.Error("get logger level error")
		 }
	}
}

func TestSetLoggerLevel(t *testing.T) {
	conf := NewConfig("../configuration/namespaces/global/config/global.yaml")
	conf.Set(LOG_DUMP_FILE, false)
	InitHyperLogger("global", conf)
	log := GetLogger("global", "consensus")
	SetLogLevel("global", "consensus", "NOTICE")
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

func TestMultiNamespaceLogger(t *testing.T)  {
	//TODO: make the test env independent form the global.yaml

	conf := NewConfig("../configuration/namespaces/global/config/global.yaml")
	conf.Set(LOG_DUMP_FILE, false)

	namespaces := []string {"namespace1", "namespace2", "namespace3"}
	wg := &sync.WaitGroup{}
	wg.Add(len(namespaces))
	for _, ns := range namespaces {
		InitHyperLogger(ns, conf)
	}
	for _, ns := range namespaces {
		go func(ns string) {
			log1 := GetLogger(ns, "module1")
			log2 := GetLogger(ns, "module2")
			log3 := GetLogger(ns, "module3")
			for i := 0; i < 20; i ++ {
				log1.Warning("log1 ...")
				log2.Info("log2...")
				log3.Critical("log3...")

				//time.Sleep(1 * time.Second)
			}
			wg.Done()

		}(ns)
	}
	wg.Wait()

}

func TestMultiNamespaceLoggerWithDump(t *testing.T)  {
	//TODO: test more precisely by read the data in the logger dir
	conf := NewConfig("../configuration/namespaces/global/config/global.yaml")
	conf.Set(LOG_DUMP_FILE, true)
	namespaces := []string {"namespace1", "namespace2", "namespace3"}
	wg := &sync.WaitGroup{}
	wg.Add(len(namespaces))
	for _, ns := range namespaces {
		conf.Set(LOG_DUMP_FILE_DIR, "/tmp/hyperlogger/" + ns)
		InitHyperLogger(ns, conf)
	}
	for _, ns := range namespaces {
		go func(ns string) {
			log1 := GetLogger(ns, "module1")
			log2 := GetLogger(ns, "module2")
			log3 := GetLogger(ns, "module3")
			for i := 0; i < 20; i ++ {
				log1.Warning("log1 ...")
				log2.Info("log2...")
				log3.Critical("log3...")
			}
			wg.Done()

		}(ns)
	}
	wg.Wait()
}