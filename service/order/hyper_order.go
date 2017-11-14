package main

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	res "github.com/hyperchain/hyperchain/core/executor/restore"
	"github.com/hyperchain/hyperchain/hyperdb"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/ipc"
	"github.com/hyperchain/hyperchain/namespace"
	"github.com/hyperchain/hyperchain/p2p"
	"github.com/hyperchain/hyperchain/rpc"
	"github.com/mkideal/cli"
	"github.com/op/go-logging"
	"github.com/terasum/viper"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var branch, commitID, date string

type HyperOrder struct {
	nsMgr       namespace.NamespaceManager
	hs          jsonrpc.RPCServer
	ipcShell    *ipc.IPCServer
	p2pmgr      p2p.P2PManager
	stopFlag    chan bool
	restartFlag chan bool
	args        *argT
}

func newHyperOrder(argV *argT) *HyperOrder {
	hp := &HyperOrder{
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
	if err != nil {
		panic(err)
	}

	hp.ipcShell = ipc.NEWIPCServer(globalConfig.GetString(common.P2P_IPC))
	p2pManager, err := p2p.GetP2PManager(vip)
	if err != nil {
		panic(err)
	}
	hp.p2pmgr = p2pManager
	hp.nsMgr = namespace.GetNamespaceManager(globalConfig, hp.stopFlag, hp.restartFlag)
	hp.hs = jsonrpc.GetRPCServer(hp.nsMgr, hp.nsMgr.GlobalConfig())

	return hp
}

func (h *HyperOrder) start() {
	logger.Notice("Hyperchain order starting...")
	go h.nsMgr.Start()
	go h.hs.Start()
	//go CheckLicense(h.stopFlag) TODO: add license check
	go h.ipcShell.Start()
}

func (h *HyperOrder) stop() {
	logger.Critical("Order stop...")
	h.nsMgr.Stop()
	time.Sleep(3 * time.Second)
	h.hs.Stop()
	logger.Critical("Hyperchain order stopped")
}

func (h *HyperOrder) restart() {
	logger.Critical("Hyperchain order restart...")
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

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Panic: ", r)
			}
		}()

		argv := ctx.Argv().(*argT)

		if argv.Version {
			if branch == "" || date == "" || commitID == "" {
				fmt.Println("Please run ./scripts/build.sh to build hyperchain with version information.")
				return nil
			}
			version := fmt.Sprintf("Hyperchain Version:\n%s-%s-%s", branch, date, commitID)
			fmt.Println(version)

			return nil
		}

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
			hp := newHyperOrder(argv)
			run(hp, argv)
		}
		return nil
	})
}

func run(inst *HyperOrder, argv *argT) {
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

//TODO: restore may not be here.
func restore(conf *common.Config, sid string, namespace string) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		fmt.Println("[RESTORE] init db failed.")
		fmt.Println("[RESTORE] detail reason: ", err.Error())
		return
	}
	handler := res.NewRestorer(conf, db, namespace)
	if err := handler.Restore(sid); err != nil {
		fmt.Printf("[RESTORE] restore from snapshot %s failed.\n", sid)
		fmt.Println("[RESTORE] detail reason: ", err.Error())
		return
	}
	fmt.Printf("[RESTORE] restore from snapshot %s success.\n", sid)
}

func setupPProf(port string) {
	addr := "0.0.0.0:" + port
	go func() {
		http.ListenAndServe(addr, nil)
	}()
}
