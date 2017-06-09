package p2p

import (
	"github.com/abiosoft/ishell"
	"golang.org/x/crypto/sha3"
	"hyperchain/p2p/network"
	"github.com/ethereum/go-ethereum/event"
	"sync"
	"hyperchain/common"
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

var once sync.Once
var p2pManager P2PManager
func GetP2PManager(config *common.Config) P2PManager{
	once.Do(func(){
			p2pManager = newP2PManager(config)
		})
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

func (p2pmgr *p2pManagerImpl) Register(namespace string,identifier string,peerID string) error{
	hasher := sha3.New256()
	hasher.Write([]byte(namespace))
	hasher.Write([]byte(identifier))
	hash := hasher.Sum(nil)
	peer,err := p2pmgr.peerMap.Get(peerID)
	if err != nil{
		return err
	}

}

func (p2pmgr p2pManagerImpl) startNode(){

}