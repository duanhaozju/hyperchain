package peerComm

import (
	"testing"
	"hyperchain/p2p/peermessage"
)

func TestNewConfigReader(t *testing.T) {
	NewConfigReader("/home/chenquan/Workspace/go/src/hyperchain/config/peerconfig.json");
}

func TestConfigReader_Persist(t *testing.T) {
	peerlist := make(map[string]peermessage.PeerAddr)
	p1 := peermessage.NewPeerAddr("127.0.0.1",8005,8085,5)
	p2 := peermessage.NewPeerAddr("127.0.0.1",8006,8086,6)
	p3 := peermessage.NewPeerAddr("127.0.0.1",8007,8087,7)
	p4 := peermessage.NewPeerAddr("127.0.0.1",8007,8087,7)
	peerlist[p1.Hash] = *p1
	peerlist[p2.Hash] = *p2
	peerlist[p3.Hash] = *p3
	peerlist[p4.Hash] = *p4
	conf := NewConfigReader("./test/local_peerconfig.json")
	//conf.AddNodesAndPersist(peerlist)
	conf.DelNodesAndPersist(peerlist)
}
