package p2p

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/terasum/viper"
	"hyperchain/common"
	"hyperchain/ipc"
	"hyperchain/manager/event"
	"hyperchain/p2p/network"
)

var (
	errP2PMGRNotInit = errors.New("The P2P manager hasn't been initlized, Fatal error")
	glogger          *logging.Logger
	globalP2PManager P2PManager
)

type P2PManager interface {

	// Start starts services under this namespace.
	Start() error

	// Stop stops services under this namespace.
	Stop() error

	// Restart restarts services under this namespace.
	Restart() error

	// GetPeerManager returns a PeerManager instance.
	GetPeerManager(namespace string, conf *viper.Viper, eventMux *event.TypeMux, delChan chan bool) (PeerManager, error)
}

// p2pManagerImpl implements the P2PManager interface.
type p2pManagerImpl struct {
	hypernet *network.HyperNet
	conf     *viper.Viper
	ipcShell *ipc.IPCServer
}

// GetP2PManager creates a new p2pManagerImpl instance and starts p2pManager service.
// If p2pManagerImpl instance existed, returns it.
func GetP2PManager(vip *viper.Viper) (P2PManager, error) {
	glogger = common.GetLogger(common.DEFAULT_LOG, "p2p")
	if globalP2PManager == nil {
		p2pManager, err := newP2PManager(vip)
		if err != nil {
			glogger.Errorf("fatal error %s", err.Error())
			return nil, errors.New(fmt.Sprintf("there are something wrong when get p2pmanager: %s", err.Error()))
		}
		globalP2PManager = p2pManager
	}
	return globalP2PManager, nil
}

func ClearP2PManager() {
	globalP2PManager = nil
}

// newP2PManager creates and returns a new p2pManagerImpl instance, starts p2pManager service.
func newP2PManager(vip *viper.Viper) (*p2pManagerImpl, error) {
	sname := vip.GetString(common.P2P_SERVERNAME)
	net, err := network.NewHyperNet(vip, sname)
	if err != nil {
		return nil, err
	}
	p2pmgr := &p2pManagerImpl{
		hypernet: net,
		conf:     vip,
	}

	if err := p2pmgr.Start(); err != nil {
		return nil, err
	}

	ipc.RegisterFunc("network", p2pmgr.hypernet.Command)
	return p2pmgr, nil
}

// Start initializes hypernet server and client.
func (mgr *p2pManagerImpl) Start() (err error) {
	// if there are something wrong cause a panic,
	// here will recover
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	err = mgr.hypernet.InitServer()
	if err != nil {
		return
	}
	err = mgr.hypernet.InitClients()
	return
}

// Stop stops hypernet services, including stop hypernet server and client.
func (p2pmgr *p2pManagerImpl) Stop() error {
	p2pmgr.hypernet.Stop()
	return nil
}

// Restart restarts services under this namespace.
// TODO
func (p2pmgr *p2pManagerImpl) Restart() error {
	return nil
}

// GetPeerManager creates and returns a new PeerManager instance, every namespace has an independent instance.
// Parameter peerConf is file peerconfig.toml.
func (p2pmgr *p2pManagerImpl) GetPeerManager(namespace string, peerConf *viper.Viper, eventMux *event.TypeMux, delChan chan bool) (PeerManager, error) {
	if p2pmgr == nil {
		return nil, errP2PMGRNotInit
	}
	return newPeerManagerImpl(namespace, peerConf, eventMux, p2pmgr.hypernet, delChan)
}

// GetPeerManager will call p2pManagerImpl.GetPeerManager to create and return a new peerManagerImpl instance
// implements PeerManager interface. Every namespace should use this function to get a PeerManager, and then,
// the PeerManager can supply all the high level methods.
// PeerManager interface method are same as hyperchain version 1.2, so all the high level interface needn't modify.
func GetPeerManager(namespace string, peerConfPath string, eventMux *event.TypeMux, delChan chan bool) (PeerManager, error) {
	if globalP2PManager == nil {
		return nil, errP2PMGRNotInit
	}

	if !common.FileExist(peerConfPath) {
		return nil, errors.New(fmt.Sprintf("connot find the peer config file %s", peerConfPath))
	}

	peerConf := viper.New()
	peerConf.SetConfigFile(peerConfPath)
	err := peerConf.ReadInConfig()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("connot readin the config file %s ,err: %s", peerConfPath, err.Error()))
	}
	return globalP2PManager.GetPeerManager(namespace, peerConf, eventMux, delChan)
}
