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
	"hyperchain/p2p"
	"github.com/terasum/viper"
	"hyperchain/p2p/ipc"
	"fmt"
)

type hyperchain struct {
	nsMgr       namespace.NamespaceManager
	hs          jsonrpc.HttpServer
	p2pmgr 	    p2p.P2PManager
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
	logger = common.GetLogger(common.DEFAULT_LOG, "main")
	//P2P module MUST Start before namespace server
	vip := viper.New()
	vip.SetConfigFile(hp.args.ConfigPath)
	err := vip.ReadInConfig()
	if err != nil{
		panic(err)
	}
	p2pManager,err  := p2p.GetP2PManager(vip)
	if err != nil{
		panic(err)
	}
	hp.p2pmgr = p2pManager

	//
	//common.InitLog(globalConfig)

	httpPort := vip.GetInt("global.jsonrpc_port")
	hp.nsMgr = namespace.GetNamespaceManager(globalConfig)
	hp.hs = jsonrpc.GetHttpServer(hp.nsMgr, hp.stopFlag, hp.restartFlag,httpPort)

	return hp
}

func (h *hyperchain) start() {
	logger.Critical("Hyperchain server starting...")
	go h.nsMgr.Start()
	go h.hs.Start()
	go CheckLicense(h.stopFlag)
	logger.Critical("Hyperchain server started successful")
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
	ConfigPath  string `cli:"c,conf" usage:"config file path" dft:"./global.yaml"`
	IPCEndpoint string `cli:"ipc" usage:"ipc interactive shell attach endpoint" dft:"./hpc.ipc"`
	Shell       bool `cli:"s,shell" usage:"start interactive shell" dft:"false"`
}

var (
	logger *logging.Logger
)

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Pannic: ", r)
			}
		}()

		argv := ctx.Argv().(*argT)
		if argv.Shell {
			fmt.Println("Start hypernet interactive shell: ",argv.IPCEndpoint)
			ipc.IPCShell(argv.IPCEndpoint)
			return nil
		}
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