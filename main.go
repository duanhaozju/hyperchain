//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/mkideal/cli"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/namespace"
	"time"
	"hyperchain/rpc"
)

type hyperchain struct {
	nsMgr       namespace.NamespaceManager
	rpcServer   jsonrpc.RPCServer
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
	common.InitHyperLoggerManager(globalConfig)
	//
	//common.InitLog(globalConfig)

	hp.nsMgr = namespace.GetNamespaceManager(globalConfig)
	hp.rpcServer = jsonrpc.GetRPCServer(hp.nsMgr, hp.stopFlag, hp.restartFlag)

	logger = common.GetLogger(common.DEFAULT_LOG, "main")
	return hp
}

func (h *hyperchain) start() {
	logger.Critical("Hyperchain server start...")
	h.nsMgr.Start()
	go h.rpcServer.Start()
	go CheckLicense(h.stopFlag)
	logger.Critical("Hyperchain server started")
}

func (h *hyperchain) stop() {
	logger.Critical("Hyperchain server stop...")
	h.nsMgr.Stop()
	time.Sleep(3 * time.Second)
	h.rpcServer.Stop()
	logger.Critical("Hyperchain server stopped")
}

func (h *hyperchain) restart() {
	logger.Critical("Hyperchain server restart...")
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