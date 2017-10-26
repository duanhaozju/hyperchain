package main

import (
	"fmt"
	"github.com/mkideal/cli"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"hyperchain/service/executor/admin"
	"hyperchain/service/executor/apiserver"
	"hyperchain/service/executor/manager"
)

type executorGlobal struct {
	exeMgr      manager.ExecutorManager
	admin       *admin.Administrator
	apiServer   apiserver.APIServer
	stopFlag    chan bool
	restartFlag chan bool
	args        *argT
}

func newExecutorGlobal(argV *argT) *executorGlobal {
	eg := &executorGlobal{
		stopFlag:    make(chan bool),
		restartFlag: make(chan bool),
		args:        argV,
	}
	globalConfig := common.NewConfig(eg.args.ConfigPath)
	globalConfig.Set(common.GLOBAL_CONFIG_PATH, eg.args.ConfigPath)
	common.InitHyperLoggerManager(globalConfig)
	logger = common.GetLogger(common.DEFAULT_LOG, "main")
	hyperdb.InitDBMgr(globalConfig)

	eg.exeMgr = manager.GetExecutorMgr(globalConfig, eg.stopFlag, eg.restartFlag)
	eg.admin = admin.NewAdministrator(eg.exeMgr, globalConfig)

	eg.apiServer = apiserver.GetAPIServer(eg.exeMgr, globalConfig)
	//TODO provides params to create a APIServer

	return eg
}

func main() {
	/**
	1. get config file

	2. get the executorManager instance and start all executor

	3. run APIServer to provide service

	*/

	cli.Run(new(argT), func(ctx *cli.Context) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Panic: ", r)
			}
		}()

		argv := ctx.Argv().(*argT)

		//start services, include executor manager and apiServer
		eg := newExecutorGlobal(argv)
		run(eg, argv)

		return nil
	})
}

func (h *executorGlobal) start() error {
	//err := h.exeMgr.Start()
	err := h.admin.Start()
	if err != nil {
		return err
	}

	//h.apiServer.Start()
	go h.apiServer.Start()
	return nil

}

func (h *executorGlobal) stop() {

}

func (h *executorGlobal) restart() {
	logger.Critical("executor server restart...")
	h.stop()
	h.start()
}

type argT struct {
	cli.Helper
	Version       bool   `cli:"v,version" usage:"get the version of hyperchain"`
	RestoreEnable bool   `cli:"r,restore" usage:"enable restore system status from dumpfile"`
	SId           string `cli:"sid" usage:"use to specify snapshot" dft:""`
	Namespace     string `cli:"n,namespace" usage:"use to specify namspace" dft:"global"`
	ConfigPath    string `cli:"c,conf" usage:"config file path" dft:"./global.toml"`
	IPCEndpoint   string `cli:"ipc" usage:"ipc interactive shell attach endpoint" dft:"./hpc.ipc"`
	Shell         bool   `cli:"s,shell" usage:"start interactive shell" dft:"false"`
	PProfEnable   bool   `cli:"pprof" usage:"use to specify whether to turn on pprof monitor or not"`
	PPort         string `cli:"pport" usage:"use to specify pprof http port"`
}

var (
	logger *logging.Logger
)

func run(inst *executorGlobal, argv *argT) {
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
