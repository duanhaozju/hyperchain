package p2p

import (
	"hyperchain/p2p/network"
	"github.com/spf13/viper"
	"hyperchain/manager/event"
	"github.com/pkg/errors"
	"fmt"
	"hyperchain/common"
	"sync"
)

type P2PManager interface {
	Start() error
	//Stop stop services under this namespace.
	Stop() error
	//Restart restart services under this namespace.
	Restart() error

	Register(namespace string,identifier string,peerID string) error

	GetPeerManager(conf *viper.Viper)(PeerManager,error)
}
type p2pManagerImpl struct {
	hypernet network.HyperNet
	conf *viper.Viper
}

var once sync.Once
var p2pManager *p2pManagerImpl
//var logger = common.GetLogger(common.DEFAULT_LOG, "p2p")

func GetP2PManager(conpath string)*PeerManager{
	if p2pManager == nil{
		once.Do(func() {
			p2pManager = newP2PManager(conpath)
		})
	}
	return p2pManager
}

func newP2PManager(conpath string)*p2pManagerImpl{
	if p2pManager != nil{
		return p2pManager
	}
	vip := viper.New()
	vip.SetConfigFile(conpath)
	err := vip.ReadInConfig()
	if err != nil{
		return err
	}
	return &p2pManagerImpl{
		hypernet:network.NewHyperNet(vip),
		conf:vip,
	}
	//TODO setup the p2pManager
	return nil
}

func (mgr *p2pManagerImpl)StartUp() (err error) {
	// if there are something wrong cause a panic,
	// here will recover
	defer func() {
		if r := recover();r != nil{
			err = r
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
func GetPeerManager(namespace string, config *common.Config,eventMux *event.TypeMux) (PeerManager,error){
	if &p2pManager == nil{
		return nil,errors.New("the P2P manager hasn't been initlized, Fatal error")
	}
	peerconf := viper.New()
	peerconf.SetConfigFile(config.GetString("global.p2p.hosts"))
	if err := peerconf.ReadInConfig(); err != nil{
		return nil,err
	}
	return p2pManager.GetPeerManager(namespace,peerconf)
}

func (p2pmgr *p2pManagerImpl) GetPeerManager(namespace string,conf *viper.Viper)(*peerManagerImpl,error){
	//TODO return a namespace's peer manager instance
	routers := conf.GetStringSlice("nodes")
	for route := range routers{
		fmt.Println(route)
	}
	return nil,nil
}

func(p2pmgr *p2pManagerImpl) bind(namespace string,innerID string,peerid string,tmux *event.TypeMux){

}

//func (p2pmgr *p2pManagerImpl) Register(namespace string,identifier string,peerID string) error{
//	hasher := sha3.New256()
//	hasher.Write([]byte(namespace))
//	hasher.Write([]byte(identifier))
//	//hash := hasher.Sum(nil)
//	//peer,err := p2pmgr.peerMap.Get(peerID)
//	//if err != nil{
//	//	return err
//	//}
//
//}

func (p2pmgr p2pManagerImpl) startNode(){

}