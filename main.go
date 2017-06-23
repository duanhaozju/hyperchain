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
	"github.com/spf13/viper"
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
	//P2P module MUST Start before namespace server
	vip := viper.New()
	vip.SetConfigFile(hp.args.ConfigPath)
	err := vip.ReadInConfig()
	if err != nil{
		panic(err)
	}
	//vip.Set("global.p2p.hosts", utils.GetProjectPath()+"/p2p/test/hosts.yaml")
	//vip.Set("global.p2p.retrytime", "3s")
	//vip.Set("global.p2p.port",50019)
	p2pManager,err  := p2p.GetP2PManager(vip)
	if err != nil{
		panic(err)
	}else {
		fmt.Println("p2pmanager started.")
	}
	hp.p2pmgr = p2pManager

	//
	//common.InitLog(globalConfig)

	hp.nsMgr = namespace.GetNamespaceManager(globalConfig)
	hp.hs = jsonrpc.GetHttpServer(hp.nsMgr, hp.stopFlag, hp.restartFlag)

	logger = common.GetLogger(common.DEFAULT_LOG, "main")
	return hp
}

func (h *hyperchain) start() {
	logger.Critical("Hyperchain server start...")
	go h.nsMgr.Start()
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
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./global.yaml"`
}

var (
	logger *logging.Logger
)

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		//defer func() {
		//	if r := recover(); r != nil {
		//		fmt.Println("Recovered in f", r)
		//	}
		//}()
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