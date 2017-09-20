package namespace

import (
	"bytes"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/vm/jcee/go"
	ledger "hyperchain/core/vm/jcee/go/ledger"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

const (
	BinHome    = "hyperjvm/bin"
	StartShell = "start_hyperjvm.sh"
	StopShell  = "stop_hyperjvm.sh"
)

type JvmManager struct {
	ledgerProxy *ledger.LedgerProxy  // ledger server handler, use to support ledger service
	jvmCli      jvm.ContractExecutor // system jvm client, for health maintain
	logger      *logging.Logger      // logger
	conf        *common.Config
	exit        chan bool
	lsofPath    string
}

func NewJvmManager(conf *common.Config) *JvmManager {
	return &JvmManager{
		ledgerProxy: ledger.NewLedgerProxy(conf),
		jvmCli:      jvm.NewContractExecutor(conf, common.DEFAULT_NAMESPACE),
		logger:      common.GetLogger(common.DEFAULT_LOG, "nsmgr"),
		conf:        conf,
		exit:        make(chan bool),
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
	mgr.logger.Noticef("Try to start jvm server")
	binHome, err := getBinDir()
	if err != nil {
		return err
	}

	if !common.FileExist(binHome) {
		return fmt.Errorf("Hyperjvm bin is not found, path: %s is not existed!", binHome)
	}

	cmd := exec.Command(path.Join(binHome, StartShell), strconv.Itoa(mgr.conf.GetInt(common.JVM_PORT)), strconv.Itoa(mgr.conf.GetInt(common.LEDGER_PORT)))
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Start()
	if err != nil {
		mgr.logger.Error(out.String())
		return err
	}
	mgr.logger.Info("executor start hyperjvm command successful")
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
	cmd := exec.Command(path.Join(binHome, StopShell), strconv.Itoa(mgr.conf.GetInt(common.JVM_PORT)), strconv.Itoa(mgr.conf.GetInt(common.LEDGER_PORT)))
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
		case <-mgr.exit:
			return
		case <-ticker.C:
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
	noLsof := true
	if len(mgr.lsofPath) == 0 {
		path, err := exec.LookPath("lsof")
		logger.Debugf(path)
		if err != nil {
			paths := []string{"/usr/sbin/lsof", "/usr/bin/lsof"}
			for _, p := range paths {
				if err := findExecutable(p); err == nil {
					path = p
					noLsof = false
					break
				}
			}
		} else {
			noLsof = false
		}
		if len(path) == 0 || noLsof {
			logger.Errorf("No lsof command found")
			return true
		}
		mgr.lsofPath = path
	}

	subcmd := fmt.Sprintf("-i:%d", mgr.conf.GetInt(common.JVM_PORT))
	ret, err := exec.Command(mgr.lsofPath, subcmd).Output()
	if err != nil || len(ret) == 0 {
		if err == nil {
			return false
		} else {
			mgr.logger.Error(err.Error())
			return true
		}
	} else {
		return true
	}
}

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}

func getBinDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, BinHome), nil
}
