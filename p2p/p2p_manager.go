package p2p

import (
	"hyperchain/p2p/network"
	"github.com/spf13/viper"
	"hyperchain/manager/event"
	"github.com/pkg/errors"
	"fmt"
	"hyperchain/common"
)

type P2PManager interface {
	Start() error
	//Stop stop services under this namespace.
	Stop() error
	//Restart restart services under this namespace.
	Restart() error

	Register(namespace string,identifier string,peerID string) error

	GetPeerManager(namespace string,conf *viper.Viper,eventMux *event.TypeMux)(PeerManager,error)
}


type p2pManagerImpl struct {
	hypernet *network.HyperNet
	conf *viper.Viper
}

var p2pManager *p2pManagerImpl
//var logger = common.GetLogger(common.DEFAULT_LOG, "p2p")


func GetP2PManager(vip *viper.Viper)(P2PManager,error){
	var err error
	if p2pManager == nil{
		p2pManager,err = newP2PManager(vip)
		if err != nil{
			return nil,errors.New(fmt.Sprintf("there something wrong when get p2pmanager: %s",err.Error()))
		}
	}
	return p2pManager,nil
}

//ClearP2PManager clear the global p2pmanger, this is for test
func ClearP2PManager()error{
	if p2pManager != nil{
		p2pManager = nil
	}
	return nil
}
func newP2PManager(vip *viper.Viper)(*p2pManagerImpl,error){
	if p2pManager != nil{
		return p2pManager,nil
	}
	net,err :=network.NewHyperNet(vip)
	if err !=nil{
		return nil,err
	}
	p2pmgr :=  &p2pManagerImpl{
		hypernet:net,
		conf:vip,
	}
	p2pmgr.Start()

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
	err = mgr.hypernet.InitClients()
	return
}

//GetPeerManager this function is global access available, every namespace
//should use this function to get a peer manager, and then, the peer manager
//can supply all the high level methods.
//the interface method are same as hyperchain version 1.2, so all the high level
//interface needn't modify.
func GetPeerManager(namespace string, peerConfpath string,eventMux *event.TypeMux) (PeerManager,error){
	if &p2pManager == nil{
		return nil,errors.New("the P2P manager hasn't been initlized, Fatal error")
	}

	if !common.FileExist(peerConfpath){
		return nil,errors.New(fmt.Sprintf("connot find the peer config file %s", peerConfpath))
	}

	peerConf := viper.New()
	peerConf.SetConfigFile(peerConfpath)
	err := peerConf.ReadInConfig()
	if err != nil{
		return nil,errors.New(fmt.Sprintf("connot readin the config file %s ,err: %s", peerConfpath,err.Error()))
	}

	return p2pManager.GetPeerManager(namespace, peerConf,eventMux)
}


//GetPeerManager get a peermanager instance, every namespace has an independent instance
func (p2pmgr *p2pManagerImpl) GetPeerManager(namespace string,peerConf *viper.Viper,eventMux *event.TypeMux)(PeerManager,error){
	return NewPeerManagerImpl(namespace,peerConf,eventMux)
}


//Stop stop services under this namespace.
func(p2pmgr *p2pManagerImpl)Stop() error{
	return nil
}

//Restart restart services under this namespace.
func(p2pmgr *p2pManagerImpl)Restart() error{
	return nil
}


func(p2pmgr *p2pManagerImpl)Register(namespace string,identifier string,peerID string) error{
	return nil

}

