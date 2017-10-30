package main

import (
	"github.com/mkideal/cli"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"hyperchain/rpc"
	"hyperchain/service/hypexec/admin"
	"hyperchain/service/hypexec/controller"
)

type HyperExecutor struct {
	exeCtl      controller.ExecutorController
	admin       *admin.Administrator
	apiServer   jsonrpc.RPCServer
	stopFlag    chan bool
	restartFlag chan bool
	args        *argT
}

func newHyperExecutor(argV *argT) *HyperExecutor {
	he := &HyperExecutor{
		stopFlag:    make(chan bool),
		restartFlag: make(chan bool, 1),
		args:        argV,
	}
	globalConfig := common.NewConfig(he.args.ConfigPath)
	globalConfig.Set(common.GLOBAL_CONFIG_PATH, he.args.ConfigPath)
	common.InitHyperLoggerManager(globalConfig)

	logger = common.GetLogger(common.DEFAULT_LOG, "hypexec")
	hyperdb.InitDBMgr(globalConfig)

	he.exeCtl = controller.GetExecutorCtl(globalConfig, he.stopFlag, he.restartFlag)
	he.admin = admin.NewAdministrator(he.exeCtl, globalConfig)

	he.apiServer = jsonrpc.GetRPCServer(he.exeCtl, globalConfig, true)
	return he
}

func main() {
	/**
	1. get config file
	2. get the executorManager instance and start all executor
	3. run APIServer to provide open service
	*/
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		eg := newHyperExecutor(argv)
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

func (h *HyperExecutor) stop() {
	go h.admin.Stop()
	go h.exeCtl.Stop()
	go h.apiServer.Stop()
}

func (h *HyperExecutor) restart() {
	logger.Critical("executor server restart...")
	h.stop()
	h.start()
}

type argT struct {
	cli.Helper
	Version       bool   `cli:"v,version" usage:"get the version of executor"`
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

func run(inst *HyperExecutor, argv *argT) {
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
