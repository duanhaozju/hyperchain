//author:zsx
//data:2016-11-3

package api

import (
	"fmt"
	"github.com/hyperchain/hyperchain/accounts"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/hyperchain/hyperchain/manager"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p"
	"strconv"
	"testing"
	"time"
)

func Test_CompileContract(t *testing.T) {

	cAPI := NewPublicContractAPI(nil, nil, nil, false, 1, 1)

	_, err := cAPI.CompileContract("")
	fmt.Println(err)

	_, err = cAPI.CompileContract(`
		// it is a accumulator
		contract Accumulator{
		    uint32 sum = 0;
		    bytes32 hello = "abcdefghijklmnopqrstuvwxyz";
		
		    function increment(){
		        sum = sum + 1;
		    }
		
		    function getSum() returns(uint32){
		        return sum;
		    }
		
		    function getHello() constant returns(bytes32){
		        return hello;
		    }
		
		    function add(uint32 num1,uint32 num2) {
		        sum = sum+num1+num2;
		    }
		}
		`)
	if err != nil {
		//t.Errorf("cAPI.CompileContract fail ")
	}

}

func Test_DeployContract(t *testing.T) {
	//单例数据库状态设置为close
	hyperdb.Setclose()
	//初始化数据
	core.InitDB("./build/keystore", 8001)

	//初始化AccountManager
	keydir := "./build/keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GenerateNodeKey(strconv.Itoa(1), "./build/keynodes")
	am := accounts.NewAccountManager(keydir, encryption)

	Peermanager1 := &p2p.GRPCPeerManager{
		NodeID: uint64(1),
	}
	eventMux1 := new(event.TypeMux)

	expiredTime := time.Time{}
	//初始化pm
	pm := manager.NewEventHub(nil, Peermanager1, eventMux1, nil, am, nil, 0, true, nil, expiredTime)

	//获取db句柄
	db, _ := hyperdb.GetDBDatabase()

	cAPI := NewPublicContractAPI(eventMux1, pm, db, true, 1, 3000)

	cAPI2 := NewPublicContractAPI(eventMux1, pm, db, false, 1, 3000)

	from := common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89")
	to := common.HexToAddress("b18c8575e3284e79b92100025a31378feb8100d6")
	var args2 = SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      NewInt64ToNumber(1000),
		GasPrice: NewInt64ToNumber(1000),
		Value:    NewInt64ToNumber(1000),
		Payload:  "",
	}

	var args3 = SendTxArgs{
		From:     from,
		Gas:      NewInt64ToNumber(1000),
		GasPrice: NewInt64ToNumber(1000),
		Value:    NewInt64ToNumber(1000),
		Payload:  "",
	}

	var args4 = SendTxArgs{
		From:      from,
		Gas:       NewInt64ToNumber(1000),
		GasPrice:  NewInt64ToNumber(1000),
		Value:     NewInt64ToNumber(1000),
		Signature: "123",
		Payload:   "",
	}

	_, err := cAPI.DeployContract(args2)
	fmt.Println(" cAPI.DeployContract(args2):" + err.Error())
	if err == nil {
		t.Errorf("DeployContract lock fail")
	}
	_, err = cAPI.DeployContract(args2)
	fmt.Println(" cAPI.DeployContract(args2):" + err.Error())
	if err == nil {
		t.Errorf("DeployContract 流量控制有问题")
	}

	_, err = cAPI2.DeployContract(args2)
	if err == nil {
		t.Errorf("DeployContract  fail 未解锁账户交易成功")
	}

	_, err = cAPI2.DeployContract(args4)
	if err == nil {
		fmt.Println(err.Error())
		t.Errorf("DeployContract  fail 无效签名通过")
	}

	//解锁账户
	am.UnlockAllAccount("./build/keystore")
	fmt.Println("success unlock all account")
	_, err = cAPI2.DeployContract(args2)
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("DeployContract  fail cAPI2.DeployContract(args2)")
	}

	_, err = cAPI2.DeployContract(args3)
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("DeployContract  fail 部署合约失败")
	}

	pm2 := manager.NewEventHub(nil, Peermanager1, nil, nil, am, nil, 0, true, nil, expiredTime)
	cAPI3 := NewPublicContractAPI(eventMux1, pm2, db, false, 1, 3000)
	_, err = cAPI3.DeployContract(args2)
	if err == nil {
		t.Errorf("DeployContract  fail 未判断eventMux为空")
	}
}
func Test_InvokeContract(t *testing.T) {

	core.InitDB("./build/keystore", 8001)

	//初始化AccountManager
	keydir := "./build/keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GenerateNodeKey(strconv.Itoa(1), "./build/keynodes")
	am := accounts.NewAccountManager(keydir, encryption)

	Peermanager1 := &p2p.GRPCPeerManager{
		NodeID: uint64(1),
	}
	eventMux1 := new(event.TypeMux)

	expiredTime := time.Time{}
	//初始化pm
	pm := manager.NewEventHub(nil, Peermanager1, eventMux1, nil, am, nil, 0, true, nil, expiredTime)

	//获取db句柄
	db, _ := hyperdb.GetDBDatabase()

	cAPI := NewPublicContractAPI(eventMux1, pm, db, true, 1, 3000)

	from := common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89")
	to := common.HexToAddress("b18c8575e3284e79b92100025a31378feb8100d6")
	var args2 = SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      NewInt64ToNumber(1000),
		GasPrice: NewInt64ToNumber(1000),
		Value:    NewInt64ToNumber(1000),
		Payload:  "",
	}

	_, err := cAPI.InvokeContract(args2)
	fmt.Println(" cAPI.InvokeContract(args2):" + err.Error())
	if err == nil {
		t.Errorf("InvokeContract lock fail")
	}

	_, err = cAPI.InvokeContract(args2)
	fmt.Println(" cAPI.DInvokeContract(args2):" + err.Error())
	if err == nil {
		t.Errorf("InvokeContract 流量控制有问题")
	}
}

func Test_GetCode(t *testing.T) {
	//单例数据库状态设置为close
	hyperdb.Setclose()

	core.InitDB("./build/keystore", 8003)
	//初始化AccountManager
	keydir := "./build/keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GenerateNodeKey(strconv.Itoa(1), "./build/keynodes")
	am := accounts.NewAccountManager(keydir, encryption)

	Peermanager1 := &p2p.GRPCPeerManager{
		NodeID: uint64(1),
	}
	eventMux1 := new(event.TypeMux)

	expiredTime := time.Time{}
	//初始化pm
	pm := manager.NewEventHub(nil, Peermanager1, eventMux1, nil, am, nil, 0, true, nil, expiredTime)

	//获取db句柄
	db, _ := hyperdb.GetDBDatabase()

	cAPI := NewPublicContractAPI(eventMux1, pm, db, false, 1, 3000)

	_, err := cAPI.GetCode(common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89"), 3)
	if err != nil {
		t.Error(" cAPI.GetCode(common.HexToAddress(6201cb0448964ac597faf6fdf1f472edf2a22b89), 3) fail")
	}

	_, err = cAPI.GetCode(common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2afffff"), 5)
	if err.Error() != "leveldb: not found" {
		t.Error("cAPI.GetCode(common.HexToAddress(6201cb0448964ac597faf6fdf1f472edf2afffff), 5) fail")
	}

}

func Test_GetContractCountByAddr(t *testing.T) {
	//初始化AccountManager
	keydir := "./build/keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GenerateNodeKey(strconv.Itoa(1), "./build/keynodes")
	am := accounts.NewAccountManager(keydir, encryption)

	Peermanager1 := &p2p.GRPCPeerManager{
		NodeID: uint64(1),
	}

	expiredTime := time.Time{}
	eventMux1 := new(event.TypeMux)
	//初始化pm
	pm := manager.NewEventHub(nil, Peermanager1, eventMux1, nil, am, nil, 0, true, nil, expiredTime)

	//获取db句柄
	db, _ := hyperdb.GetDBDatabase()

	cAPI := NewPublicContractAPI(eventMux1, pm, db, false, 1, 3000)

	ref, err := cAPI.GetContractCountByAddr(common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89"), 3)
	if *ref != Number(0) {
		t.Errorf("GetContractCountByAddr fail ")
	}

	_, err = cAPI.GetContractCountByAddr(common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89"), 8)

	if err == nil {
		t.Errorf("GetContractCountByAddr fail ")
	}

	_, err = cAPI.GetStorageByAddr(common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89"), 8)

	if err == nil {
		t.Errorf("cAPI.GetStorageByAddr fail ")
	}

	_, err = cAPI.GetStorageByAddr(common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2222222"), 3)

	if err != nil {
		t.Errorf("cAPI.GetStorageByAddr fail ")
	}

	_, err = cAPI.GetStorageByAddr(common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89"), 3)

	if err != nil {
		t.Errorf("cAPI.GetStorageByAddr fail ")
	}
}
