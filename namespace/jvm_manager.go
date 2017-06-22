package namespace

import (
	ledger "hyperchain/core/vm/jcee/go/ledger"
	"hyperchain/common"
	"os/exec"
	"github.com/op/go-logging"
	"strconv"
	"path"
	"time"
	"fmt"
	"hyperchain/core/vm/jcee/go"
	"os"
)
const (
	BinHome  = "hyperjvm/bin"
	StartShell = "start_hyperjvm.sh"
	StopShell =  "stop_hyperjvm.sh"
)

type JvmManager struct {
	ledgerProxy      *ledger.LedgerProxy     // ledger server handler, use to support ledger service
	jvmCli           jvm.ContractExecutor   // system jvm client, for health maintain
	logger           *logging.Logger	 // logger
	conf             *common.Config
	exit             chan bool

}

func NewJvmManager(conf *common.Config) *JvmManager {
	return &JvmManager{
		ledgerProxy:     ledger.NewLedgerProxy(conf),
		jvmCli:          jvm.NewContractExecutor(conf, common.DEFAULT_NAMESPACE),
		logger:          common.GetLogger(common.DEFAULT_LOG, "nsmgr"),
		conf:            conf,
		exit:            make(chan bool),
	}
}

// Start turn on jvm service, include set up a ledger server, start jvm and establish a jvm connect for health check.
func (mgr *JvmManager) Start() error {
	if err := mgr.startLedgerServer(); err != nil {
		return err
	}
	if err := mgr.startJvmServer(); err != nil {
		return err
	}
	go mgr.startLedgerServerDaemon()
	go mgr.startJvmServerDaemon()
	return nil
}

// Start turn off jvm service.
func (mgr *JvmManager) Stop() error {
	if err := mgr.stopLedgerServer(); err != nil {
		return err
	}
	if err := mgr.stopJvmServer(); err != nil {
		return err
	}
	mgr.notifyToExit()
	return nil
}

func (mgr *JvmManager) startLedgerServer() error {
	err := mgr.ledgerProxy.Server()
	if err != nil {
		return err
	}
	return nil
}

func (mgr *JvmManager) startJvmServer() error {
	binHome, err := getBinDir()
	if err != nil {
		return err
	}
	cmd := exec.Command(path.Join(binHome, StartShell), strconv.Itoa(mgr.conf.GetInt(common.C_JVM_PORT)), strconv.Itoa(mgr.conf.GetInt(common.C_LEDGER_PORT)))
	err = cmd.Run()
	if err != nil {
		mgr.logger.Error(err)
		return err
	}
	mgr.logger.Info("execute start hyperjvm command successful")
	return nil
}

func (mgr *JvmManager) stopLedgerServer() error {
	mgr.ledgerProxy.StopServer()
	mgr.logger.Info("stop ledger server success")
	return nil
}

func (mgr *JvmManager) stopJvmServer() error {
	binHome, err := getBinDir()
	if err != nil {
		return err
	}
	cmd := exec.Command(path.Join(binHome, StopShell), strconv.Itoa(mgr.conf.GetInt(common.C_JVM_PORT)), strconv.Itoa(mgr.conf.GetInt(common.C_LEDGER_PORT)))
	err = cmd.Run()
	if err != nil {
		mgr.logger.Error(err)
		return err
	}
	mgr.logger.Info("execute stop hyperjvm successful")
	return nil
}

func (mgr *JvmManager) startJvmServerDaemon() {
	time.Sleep(10 * time.Second)
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <- mgr.exit:
			return
		case <- ticker.C:
			if !mgr.checkJvmExist() {
				mgr.restartJvmServer()
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func (mgr *JvmManager) restartJvmServer() {
	mgr.logger.Info("try to restart jvm server")
	err := mgr.startJvmServer()
	if err != nil {
		mgr.logger.Errorf("start jvm server failed. %v", err.Error())
	}
}

func (mgr *JvmManager) startLedgerServerDaemon() {

}

func (mgr *JvmManager) notifyToExit() {
	mgr.exit <- true
}


func (mgr *JvmManager) checkJvmExist() bool {
	subcmd := fmt.Sprintf("-i:%d", mgr.conf.GetInt(common.C_JVM_PORT))
	ret, err := exec.Command("/usr/sbin/lsof", subcmd).Output()
	if err != nil || len(ret) == 0 {
		if err == nil {
			return false
		}else {
			mgr.logger.Error(err.Error())
			return true
		}
	} else {
		return true
	}
}

func getBinDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, BinHome), nil
}