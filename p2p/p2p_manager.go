package p2p

import (
	"github.com/abiosoft/ishell"
	"hyperchain/p2p/network"
	"hyperchain/common"
	"github.com/spf13/viper"
)

type P2PManager interface {
	Start() error
	//Stop stop services under this namespace.
	Stop() error
	//Restart restart services under this namespace.
	Restart() error

	Register(namespace string,identifier string,peerID string) error

	Produce(namespace string,identifier string,msg []byte) error

	consume() error
}

var p2pManager P2PManager
var logger = common.GetLogger(common.DEFAULT_LOG, "p2p")

type P2PManagerImpl struct {
	config *viper.Viper

}

func StartUpP2PManager(config *viper.Viper){
	if p2pManager != nil {
		logger.Fatal("P2P Manager Already been setuped.")
	}


}

func GetP2PManager() P2PManager{
	p2pManager = newP2PManager()
	return p2pManager
}

func newP2PManager(config *common.Config) P2PManager{
	return &p2pManagerImpl{

	}
}

type p2pManagerImpl struct {
	 shell *ishell.Shell
	 node *Node
	 peerMap PeerMap
	dnsResolver *network.DNSResolver

}

func(p2pmgr *p2pManagerImpl) Bind(namespace string,innerID string,peerid string,tmux *event.TypeMux){

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