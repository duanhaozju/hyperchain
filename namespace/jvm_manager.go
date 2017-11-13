package namespace

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/vm/jcee/go"
	ledger "github.com/hyperchain/hyperchain/core/vm/jcee/go/ledger"

	"github.com/op/go-logging"
)

// This file defines JVM related functions. In hyperchain, we use two
// kinds of executor, which EVM is the default executor, and JVM is
// an optional.

const (
	BinHome    = "hyperjvm/bin"
	StartShell = "start_hyperjvm.sh"
	StopShell  = "stop_hyperjvm.sh"
)

// JvmManager manages the JVM executor.
type JvmManager struct {
	// ledger server handler, use to support ledger service
	ledgerProxy *ledger.LedgerProxy

	// system jvm client, for health maintain
	jvmCli jvm.ContractExecutor

	// lsofPath is used to locate "lsof" as different systems may not
	// contain this executable file, we must locate it before using it
	// to check the monitoring of specific ports.
	lsofPath string

	logger *logging.Logger
	conf   *common.Config
	exit   chan bool
}

// NewJvmManager returns a JvmManager instance using given config.
func NewJvmManager(conf *common.Config) *JvmManager {
	return &JvmManager{
		ledgerProxy: ledger.NewLedgerProxy(conf),
		jvmCli:      jvm.NewContractExecutor(conf, common.DEFAULT_NAMESPACE),
		logger:      common.GetLogger(common.DEFAULT_LOG, "nsmgr"),
		conf:        conf,
		exit:        make(chan bool),
	}
}

// Start turns on jvm service, including setting up a ledger server, starting
// jvm and establishing a jvm connection for health check.
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

// Stop turns off jvm service.
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

// startLedgerServer starts the hyperchain ledger server.
func (mgr *JvmManager) startLedgerServer() error {
	err := mgr.ledgerProxy.Server()
	if err != nil {
		return err
	}
	return nil
}

// startJvmServer starts the Jvm server.
func (mgr *JvmManager) startJvmServer() error {
	mgr.logger.Noticef("Try to start jvm server")
	binHome, err := getBinDir()
	if err != nil {
		return err
	}

	if !common.FileExist(binHome) {
		mgr.logger.Errorf("Hyperjvm bin is not found, path: %s is not existed!", binHome)
		return ErrBinNotFound
	}

	cmd := exec.Command(path.Join(binHome, StartShell), strconv.Itoa(mgr.conf.GetInt(common.JVM_PORT)), strconv.Itoa(mgr.conf.GetInt(common.LEDGER_PORT)))
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Start()
	if err != nil {
		mgr.logger.Error(out.String())
		return err
	}
	mgr.logger.Info("Execute start hyperjvm command successfully")
	return nil
}

// stopLedgerServer stops the ledger server.
func (mgr *JvmManager) stopLedgerServer() error {
	mgr.ledgerProxy.StopServer()
	mgr.logger.Info("Stop ledger server successfully")
	return nil
}

// stopJvmServer stops the Jvm server.
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
	mgr.logger.Info("Execute stop hyperjvm successfully")
	return nil
}

// startJvmServerDaemon starts to check the heartbeat of Jvm server.
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

// restartJvmServer restarts the Jvm server.
func (mgr *JvmManager) restartJvmServer() {
	mgr.logger.Info("Try to restart jvm server")
	err := mgr.startJvmServer()
	if err != nil {
		mgr.logger.Errorf("Start jvm server failed: %s", err)
	}
}

func (mgr *JvmManager) startLedgerServerDaemon() {

}

// notifyToExit defines the exit interface of Jvm server.
func (mgr *JvmManager) notifyToExit() {
	mgr.exit <- true
}

// checkJvmExist checks if Jvm server is running.
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
			return false
		}
	} else {
		return true
	}
}

// findExecutable locates and checks if there exists the specific
// executable file.
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

// getBinDir returns the bin directory.
func getBinDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, BinHome), nil
}
