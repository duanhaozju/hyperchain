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
	stopFlag    chan bool
	restartFlag chan bool
	args        *argT
}

func newHyperchain(argV *argT) *hyperchain {
	hp := &hyperchain{
		stopFlag:    make(chan bool),
		restartFlag: make(chan bool),
		args:        argV,
	}

	globalConfig := common.NewConfig(hp.args.ConfigPath)
	common.InitHyperLogger(globalConfig)
	common.InitLog(globalConfig)

	hp.nsMgr = namespace.GetNamespaceManager(globalConfig)

	return hp
}

func (h *hyperchain) start() {
	logger.Critical("Hyperchan server start...")
	h.nsMgr.Start()
	go jsonrpc.Start(h.nsMgr, h.stopFlag, h.restartFlag)
	go CheckLicense(h.stopFlag)
	logger.Critical("Hyperchan server started")
}

func (h *hyperchain) stop() {
	logger.Critical("Hyperchan server stop...")
	h.nsMgr.Stop()
	time.Sleep(3 * time.Second)
	jsonrpc.Stop()
	logger.Critical("Hyperchan server stopped")
}

func (h *hyperchain) restart() {
	logger.Critical("Hyperchan server restart...")
	h.stop()
	h.start()
}

type argT struct {
	cli.Helper
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./global.yaml"`
}

var (
	logger *logging.Logger
)

func init() {
	logger = logging.MustGetLogger("main")
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		hp := newHyperchain(argv)
		hp.start()
		for {
			select {
			case <-hp.stopFlag:
				hp.stop()
				return nil
			case <-hp.restartFlag:
				hp.restart()
			}
		}
		return nil
	})
}
