package peerComm

import "testing"

func TestNewConfigReader(t *testing.T) {
	NewConfigReader("/home/chenquan/Workspace/go/src/hyperchain/config/peerconfig.json");
}

//func TestConfigReader_SaveAddress(t *testing.T) {
//	var address = Address{
//		ID:5,
//		IP:"127.0.0.1",
//		Port:8001,
//		RPCPort:8081,
//
//	}
//	//conf := NewConfigReader("/home/chenquan/Workspace/go/src/hyperchain/config/peerconfig.json");
//
//	//t.Log(conf.SaveAddress(address));
//}
