//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/mkideal/cli"
	"github.com/op/go-logging"
	"hyperchain/api/jsonrpc/core"
	"hyperchain/common"
	"hyperchain/namespace"
	"time"
)

type hyperchain struct {
	nsMgr       namespace.NamespaceManager
	hs          jsonrpc.HttpServer
	stopFlag    chan bool
	restartFlag chan bool
	args        *argT
}

func newHyperchain(argV *argT, conf *common.Config) *hyperchain {
	hp := &hyperchain{
		stopFlag:    make(chan bool),
		restartFlag: make(chan bool),
		args:        argV,
	}

	common.InitHyperLoggerManager(conf)
	//
	//common.InitLog(globalConfig)

	hp.nsMgr = namespace.GetNamespaceManager(conf)
	hp.hs = jsonrpc.GetHttpServer(hp.nsMgr, hp.stopFlag, hp.restartFlag)

	logger = common.GetLogger(common.DEFAULT_LOG, "main")
	return hp
}

func (h *hyperchain) start() {
	logger.Critical("Hyperchain server start...")
	h.nsMgr.Start()
	go h.hs.Start()
	go CheckLicense(h.stopFlag)
	logger.Critical("Hyperchain server started")
}

func (h *hyperchain) stop() {
	logger.Critical("Hyperchain server stop...")
	h.nsMgr.Stop()
	time.Sleep(3 * time.Second)
	h.hs.Stop()
	logger.Critical("Hyperchain server stopped")
}

func (h *hyperchain) restart() {
	logger.Critical("Hyperchain server restart...")
	h.stop()
	h.start()
}

type argT struct {
	cli.Helper
	ConfigPath     string `cli:"c,conf" usage:"config file path" dft:"./global.yaml"`
	RestoreEnable  bool   `cli:"r,restore" usage:"enable restore system status from dumpfile"`
	SId            string `cli:"s,sid" usage:"use to specify snapshot" dft:""`
	Namespace      string `cli:"n,namespace" usage:"use to specify namspace" dft:"global"`
}

var (
	logger *logging.Logger
)

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		globalConfig := common.NewConfig(argv.ConfigPath)

		hp := newHyperchain(argv, globalConfig)
		switch {
		case argv.RestoreEnable:
			restore(globalConfig, argv.SId, argv.Namespace)
		default:
			run(hp)
		}
		return nil
	})
}

func run(inst *hyperchain) {
	inst.start()
	for {
		select {
		case <-inst.stopFlag:
			inst.stop()
		case <-inst.restartFlag:
			inst.restart()
		}
	}
}