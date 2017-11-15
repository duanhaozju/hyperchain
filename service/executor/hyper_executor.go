package main

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/rpc"
	"github.com/hyperchain/hyperchain/service/executor/admin"
	"github.com/hyperchain/hyperchain/service/executor/controller"
	"github.com/mkideal/cli"
	"github.com/op/go-logging"
)

type HyperExecutor struct {
	exeCtl      controller.ExecutorController
	admin       *admin.Administrator
	apiServer   jsonrpc.RPCServer
	stopFlag    chan bool
	restartFlag chan bool
	args        *argT
	logger      *logging.Logger
}

func newHyperExecutor(argV *argT) *HyperExecutor {
	he := &HyperExecutor{
		stopFlag:    make(chan bool),
		restartFlag: make(chan bool, 1),
		args:        argV,
	}
	gc := common.NewConfig(he.args.ConfigPath)

	gc.Set(common.GLOBAL_CONFIG_PATH, he.args.ConfigPath)
	common.InitHyperLoggerManager(gc)

	he.logger = common.GetLogger(common.DEFAULT_LOG, "hypexec")
	he.exeCtl = controller.GetExecutorCtl(gc, he.stopFlag, he.restartFlag)
	he.admin = admin.NewAdministrator(he.exeCtl, gc)

	//TODO: fix api compatible he.apiServer = jsonrpc.GetRPCServer(he.exeCtl, gc, true)
	return he
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		he := newHyperExecutor(argv)
		run(he, argv)
		return nil
	})
}

func (h *HyperExecutor) start() error {
	h.logger.Notice("try to start hyper-executor!")
	if err := h.exeCtl.Start(); err != nil {
		panic(err)
	}

	if err := h.admin.Start(); err != nil {
		panic(err)
	}

	go func() {
		//TODO: fix api server problem
		//err := h.apiServer.Start()
		//if err != nil {
		//	panic(err)
		//}
	}()
	h.logger.Notice("hyper-executor start successful!")
	return nil

}

func (h *HyperExecutor) stop() {
	go h.admin.Stop()
	go h.exeCtl.Stop()
	go h.apiServer.Stop()
}

func (h *HyperExecutor) restart() {
	h.logger.Critical("executor server restart...")
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
