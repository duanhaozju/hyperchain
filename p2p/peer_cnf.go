package p2p

import (
	"sync"
	"github.com/terasum/viper"
)

type peerCnf struct {
	rwLock *sync.RWMutex
	vip    *viper.Viper
}

func newPeerCnf(vip *viper.Viper) *peerCnf {
	return &peerCnf{
		rwLock:new(sync.RWMutex),
		vip:vip,
	}
}

func (cnf *peerCnf)viper() *viper.Viper {
	return cnf.vip
}

func (cnf *peerCnf)RLock() {
	cnf.rwLock.RLock()
}

func (cnf *peerCnf)RUnlock() {
	cnf.rwLock.RUnlock()
}

func (cnf *peerCnf)Lock() {
	cnf.rwLock.Lock()
}
func (cnf *peerCnf)Unlock() {
	cnf.rwLock.Unlock()
}
