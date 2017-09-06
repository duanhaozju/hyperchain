//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/mkideal/cli"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/namespace"
	"time"
	"hyperchain/p2p"
	"github.com/terasum/viper"
	"hyperchain/ipc"
	"fmt"
	_ "net/http/pprof"
	"hyperchain/rpc"
)

type hyperchain struct {
	nsMgr       namespace.NamespaceManager
	hs          jsonrpc.RPCServer
	ipcShell    *ipc.IPCServer
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
	globalConfig.Set(common.GLOBAL_CONFIG_PATH, hp.args.ConfigPath)
	common.InitHyperLoggerManager(globalConfig)
	logger = common.GetLogger(common.DEFAULT_LOG, "main")
	//P2P module MUST Start before namespace server
	vip := viper.New()
	vip.SetConfigFile(hp.args.ConfigPath)
	err := vip.ReadInConfig()
	if err != nil{
		panic(err)
	}

	hp.ipcShell	 = ipc.NEWIPCServer(globalConfig.GetString(common.P2P_IPC))
	p2pManager,err	:= p2p.GetP2PManager(vip)
	if err != nil{
		panic(err)
	}
	hp.p2pmgr 	= p2pManager
	hp.nsMgr 	= namespace.GetNamespaceManager(globalConfig)
	hp.hs 		= jsonrpc.GetRPCServer(hp.nsMgr, hp.stopFlag, hp.restartFlag)

	return hp
}

func (h *hyperchain) start() {
	logger.Notice("Hyperchain server starting...")
	go h.nsMgr.Start()
	go h.hs.Start()
	go CheckLicense(h.stopFlag)
	go h.ipcShell.Start()
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
	RestoreEnable bool     `cli:"r,restore" usage:"enable restore system status from dumpfile"`
	SId           string   `cli:"sid" usage:"use to specify snapshot" dft:""`
	Namespace     string   `cli:"n,namespace" usage:"use to specify namspace" dft:"global"`
	ConfigPath    string   `cli:"c,conf" usage:"config file path" dft:"./global.toml"`
	IPCEndpoint   string   `cli:"ipc" usage:"ipc interactive shell attach endpoint" dft:"./hpc.ipc"`
	Shell         bool     `cli:"s,shell" usage:"start interactive shell" dft:"false"`
	PProfEnable   bool     `cli:"pprof" usage:"use to specify whether to turn on pprof monitor or not"`
	PPort         string   `cli:"pport" usage:"use to specify pprof http port"`
}

var (
	logger *logging.Logger
)

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Panic: ", r)
			}
		}()

		argv := ctx.Argv().(*argT)
		globalConfig := common.NewConfig(argv.ConfigPath)

		switch {
		case argv.RestoreEnable:
			// Restore blockchain
			restore(globalConfig, argv.SId, argv.Namespace)
		case argv.Shell:
			// Start interactive shell
			fmt.Println("Start hypernet interactive shell: ", argv.IPCEndpoint)
			ipc.IPCShell(argv.IPCEndpoint)
		default:
			// Start hyperchain service
			hp := newHyperchain(argv)
			run(hp, argv)
		}
		return nil
	})
}

func run(inst *hyperchain, argv *argT) {
	if argv.PProfEnable {
		setupPProf(argv.PPort)
	}
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
