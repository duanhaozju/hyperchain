//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestRawLogger(t *testing.T) {
	conf := NewRawConfig()
	InitHyperLogger("test", conf)
	logger := GetLogger("test", "a")
	logger.Info("aaa")
	logger.Debug("sss")

	InitRawHyperLogger("test2")
	logger = GetLogger("test2", "abb")
	logger.Info("aaa")
	logger.Info("sss")
}

func TestGetLogLevel(t *testing.T) {
	conf := NewConfig("../configuration/namespaces/global/config/namespace.toml")
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
	conf := NewConfig("../configuration/namespaces/global/config/namespace.toml")
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

func TestMultiNamespaceLogger(t *testing.T) {
	//TODO: make the test env independent form the global.yaml

	conf := NewConfig("../configuration/namespaces/global/config/namespace.toml")
	conf.Set(LOG_DUMP_FILE, false)
	conf.Set(LOG_BASE_LOG_LEVEL, "DEBUG")

	namespaces := []string{"namespace1", "namespace2", "namespace3"}
	wg := &sync.WaitGroup{}
	wg.Add(len(namespaces))
	for _, ns := range namespaces {
		InitHyperLogger(ns, conf)
	}
	for _, namespace := range namespaces {
		go func(ns string) {
			log1 := GetLogger(ns, "module1")
			log2 := GetLogger(ns, "module2")
			log3 := GetLogger(ns, "module3")
			for i := 0; i < 2; i++ {
				if i == 1 {
					SetLogLevel(ns, "module1", "ERROR")
				}
				log1.Warning("log1 ...")
				log2.Info("log2...")
				log3.Critical("log3...")
			}
			wg.Done()

		}(namespace)
	}
	wg.Wait()

}

func TestMultiNamespaceLoggerWithDump(t *testing.T) {
	//TODO: test more precisely by read the data in the logger dir
	conf := NewConfig("../configuration/namespaces/global/config/namespace.toml")
	conf.Set(LOG_DUMP_FILE, true)
	conf.Set(LOG_BASE_LOG_LEVEL, "DEBUG")
	defer os.RemoveAll("./namespaces")

	namespaces := []string{"namespace1", "namespace2", "namespace3"}
	wg := &sync.WaitGroup{}
	wg.Add(len(namespaces))
	for _, ns := range namespaces {
		conf.Set(LOG_DUMP_FILE_DIR, "/tmp/hyperlogger/"+ns)
		InitHyperLogger(ns, conf)
	}
	for _, ns := range namespaces {
		go func(ns string) {
			log1 := GetLogger(ns, "module1")
			log2 := GetLogger(ns, "module2")
			log3 := GetLogger(ns, "module3")
			for i := 0; i < 2; i++ {
				if i == 1 {
					SetLogLevel(ns, "module1", "ERROR")
				}
				log1.Warning("log1 ...")
				log2.Info("log2...")
				log3.Critical("log3...")
			}
			wg.Done()

		}(ns)
	}
	wg.Wait()
}

func TestHyperLoggerSplitByInterval(t *testing.T) {
	conf := NewConfig("../configuration/namespaces/global/config/namespace.toml")
	conf.Set(LOG_DUMP_FILE, true)
	conf.Set(LOG_BASE_LOG_LEVEL, "DEBUG")
	conf.Set(LOG_NEW_FILE_INTERVAL, 4*time.Second)
	conf.Set(LOG_DUMP_FILE_DIR, "/tmp/hyperlogger/split")
	os.RemoveAll("/tmp/hyperlogger/split")
	InitHyperLogger("test", conf)
	defer os.RemoveAll("./namespaces")
	logger := GetLogger("test", "module1")
	for i := 1; i < 8; i++ {
		logger.Criticalf("test write log %d", i)
		time.Sleep(1 * time.Second)
	}
}
