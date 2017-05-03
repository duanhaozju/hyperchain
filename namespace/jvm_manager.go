package namespace

import (
	ledger "hyperchain/core/vm/jcee/go/ledger"
	"hyperchain/common"
	"hyperchain/core/vm/jcee/go/client"
	"os/exec"
	"github.com/op/go-logging"
	"strconv"
	"os"
	"path"
)
const (
	BinHome  = "hyperjvm/bin"
	StartShell = "start_hyperjvm.sh"
	StopShell =  "stop_hyperjvm.sh"
)

type JvmManager struct {
	ledgerProxy      *ledger.LedgerProxy     // ledger server handler, use to support ledger service
	jvmCli           jcee.ContractExecutor   // system jvm client, for health maintain
	startCmd         *exec.Cmd               // jvm start command
	logger           *logging.Logger	 // logger
	conf             *common.Config
}

func NewJvmManager(conf *common.Config) *JvmManager {
	return &JvmManager{
		ledgerProxy:     ledger.NewLedgerProxy(conf),
		jvmCli:          jcee.NewContractExecutor(conf, common.DEFAULT_NAMESPACE),
		logger:          common.GetLogger(common.DEFAULT_LOG, "nsmgr"),
		conf:            conf,
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
	return nil
}

// Start turn off jvm service.
func (mgr *JvmManager) Stop() error {
	// TODO
	return nil
}

func (mgr *JvmManager) startLedgerServer() error {
	go mgr.ledgerProxy.Server()
	return nil
}

func (mgr *JvmManager) startJvmServer() error {
	binHome, err := getBinDir()
	if err != nil {
		return err
	}
	cmd := exec.Command(path.Join(binHome, StartShell), strconv.Itoa(mgr.conf.GetInt(common.C_JVM_PORT)), strconv.Itoa(mgr.conf.GetInt(common.C_LEDGER_PORT)))
	mgr.startCmd = cmd
	err = mgr.startCmd.Run()
	if err != nil {
		mgr.logger.Error(err)
		return err
	}
	mgr.logger.Info("execute start hyperjvm command successful")
	return nil
}

func (mgr *JvmManager) backend() {

}

func getBinDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, BinHome), nil
}

