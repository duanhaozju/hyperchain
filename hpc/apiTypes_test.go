//author :zsx
//data :2016-11-3
package hpc

//import (
//	"hyperchain/accounts"
//	"hyperchain/core"
//	"hyperchain/crypto"
//	"hyperchain/event"
//	"hyperchain/manager"
//	"testing"
//)

//func Test_apiTypes(t *testing.T) {
//	eventMux := new(event.TypeMux)
//	core.InitDB("./build/keystore", 8023)
//	keydir := "../config/keystore/"
//	encryption := crypto.NewEcdsaEncrypto("ecdsa")
//	am := accounts.NewAccountManager(keydir, encryption)
//	pm := &manager.ProtocolManager{

//		AccountManager: am,
//	}
//	api := GetAPIs(eventMux, pm)
//	if api[0].Namespace != "tx" {
//		t.Errorf("apiTypes wrong")
//	}
//}
