package p2p

import (
	"hyperchain/p2p/network"
	"github.com/terasum/viper"
	"hyperchain/manager/event"
	"github.com/pkg/errors"
	"hyperchain/common"
	"hyperchain/p2p/ipc"
	"fmt"
	"github.com/op/go-logging"
)



var (
	glogger *logging.Logger
	errP2PMGRNotInit = errors.New("The P2P manager hasn't been initlized, Fatal error")
)

type P2PManager interface {
	Start() error
	//Stop stop services under this namespace.
	Stop() error
	//Restart restart services under this namespace.
	Restart() error

	GetPeerManager(namespace string,conf *viper.Viper,eventMux *event.TypeMux,delChan chan bool)(PeerManager,error)
}


type p2pManagerImpl struct {
	hypernet *network.HyperNet
	conf *viper.Viper
	ipcShell *ipc.IPCServer
}

var globalP2PManager *p2pManagerImpl


func GetP2PManager(vip *viper.Viper)(P2PManager,error){
	glogger = common.GetLogger(common.DEFAULT_LOG,"p2p")
	if globalP2PManager == nil{
		p2pManager,err := newP2PManager(vip)
		if err != nil{
			glogger.Errorf("fatal error %s",err.Error())
			return nil,errors.New(fmt.Sprintf("there something wrong when get p2pmanager: %s",err.Error()))
		}
		globalP2PManager = p2pManager
	}
	return globalP2PManager,nil
}

//ClearP2PManager clear the global p2pmanger, this is for test
func ClearP2PManager()error{
	if globalP2PManager != nil{
		globalP2PManager.hypernet.Stop()
		globalP2PManager = nil
	}
	return nil
}

func newP2PManager(vip *viper.Viper)(*p2pManagerImpl,error){
	net,err :=network.NewHyperNet(vip)
	if err !=nil{
		return nil,err
	}
	p2pmgr :=  &p2pManagerImpl{
		hypernet:net,
		conf:vip,
		ipcShell:ipc.NEWIPCServer(vip.GetString(common.P2P_IPC)),
	}
	p2pmgr.Start()

	glogger.Notice("interactive ipc shell server listening...")
	rc := ipc.NewRemoteCall()
	ipc.RegisterFunc(rc,"network",p2pmgr.hypernet.Command)
	err = p2pmgr.ipcShell.Start(rc)
	if err != nil{
		glogger.Fatalf("cannot start ipc server, err: %s",err.Error())
	}
	return p2pmgr,nil
}

func (mgr *p2pManagerImpl)Start() (err error) {
	// if there are something wrong cause a panic,
	// here will recover
	defer func() {
		if r := recover();r != nil{
			err = r.(error)
		}
	}()
	err = mgr.hypernet.InitServer()
	if err != nil{
		return
	}
	err = mgr.hypernet.InitClients()
	return
}

//GetPeerManager this function is global access available, every namespace
//should use this function to get a peer manager, and then, the peer manager
//can supply all the high level methods.
//the interface method are same as hyperchain version 1.2, so all the high level
//interface needn't modify.
func GetPeerManager(namespace string, peerConfPath string,eventMux *event.TypeMux, delChan chan bool) (PeerManager,error){
	if globalP2PManager == nil{
		return nil,errP2PMGRNotInit
	}

	if !common.FileExist(peerConfPath){
		return nil,errors.New(fmt.Sprintf("connot find the peer config file %s", peerConfPath))
	}

	peerConf := viper.New()
	peerConf.SetConfigFile(peerConfPath)
	err := peerConf.ReadInConfig()
	if err != nil{
		return nil,errors.New(fmt.Sprintf("connot readin the config file %s ,err: %s", peerConfPath,err.Error()))
	}
	return globalP2PManager.GetPeerManager(namespace, peerConf,eventMux,delChan)
}


//GetPeerManager get a peermanager instance, every namespace has an independent instance
func (p2pmgr *p2pManagerImpl) GetPeerManager(namespace string,peerConf *viper.Viper,eventMux *event.TypeMux,delChan chan bool)(PeerManager,error){
	if p2pmgr == nil{
		return nil,errP2PMGRNotInit
	}
	return NewPeerManagerImpl(namespace,peerConf,eventMux,p2pmgr.hypernet,delChan)
}


//Stop stop services under this namespace.
func(p2pmgr *p2pManagerImpl)Stop() error{
	p2pmgr.hypernet.Stop()
	return nil
}

//Restart restart services under this namespace.
func(p2pmgr *p2pManagerImpl)Restart() error{
	return nil
}

