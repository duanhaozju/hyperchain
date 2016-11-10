//author:zsx
//data:2016-11-3
//CompileContract的错误返回有问题,compiler.CompileSourcefile()，编译出错err不为空没有返回

package hpc

//import (
//	"fmt"
//	"hyperchain/accounts"
//	//	"hyperchain/common"
//	"hyperchain/core"
//	"hyperchain/crypto"
//	"hyperchain/event"
//	"hyperchain/hyperdb"
//	"hyperchain/manager"
//	"hyperchain/p2p"
//	"testing"
//)

//func Test_PublicContractAPI(t *testing.T) {
//	core.InitDB("./build/keystore", 8023)
//	keydir := "../config/keystore/"
//	encryption := crypto.NewEcdsaEncrypto("ecdsa")
//	am := accounts.NewAccountManager(keydir, encryption)
//	Peermanager1 := &p2p.GrpcPeerManager{
//		NodeID: uint64(1),
//	}
//	pm := &manager.ProtocolManager{

//		AccountManager: am,
//		Peermanager:    Peermanager1,
//	}
//	db, _ := hyperdb.GetLDBDatabase()
//	eventMux := new(event.TypeMux)
//	cAPI := NewPublicContractAPI(eventMux, pm, db)

//	fmt.Println("-------------------------------------")
//	//	cpcode, _ := cAPI.CompileContract("")
//	//	if cpcode != nil {
//	//		t.Error("CompileContracttest1 wrong")
//	//	}

//	_, err := cAPI.CompileContract(`
//		contract Accumulator{
//		    uint32 sum = 0;
//		    bytes32 hello = "abcdefghijklmnopqrstuvwxyz";

//		    function increment(){
//		        sum = sum + 1;
//		    }

//		    function getSum() returns(uint32){
//		        return sum;
//		    }

//		    function getHello() constant returns(bytes32){
//		        return hello;
//		    }

//		    function add(uint32 num1,uint32 num2) {
//		        sum = sum+num1+num2;
//		    }
//		}
//		`)
//	if err != nil {
//		t.Error("CompileContractytest2 wrong")
//	}
//	//
//	//	var args2 = SendTxArgs{
//	//		From:     from,
//	//		To:       &to,
//	//		Gas:      NewInt64ToNumber(1000),
//	//		GasPrice: NewInt64ToNumber(1000),
//	//		Value:    NewInt64ToNumber(1000),
//	//		Payload:  "",
//	//	}
//	//	_, err2 := cAPI.DeployContract(args2)
//	//	if err2 == nil {
//	//		t.Errorf("DeployContract1 wrong")
//	//	}
//	//	_, err3 := cAPI.DeployContract(args)
//	//	if err3 == nil {
//	//		t.Errorf("DeployContract2 wrong")
//	//	}
//}
