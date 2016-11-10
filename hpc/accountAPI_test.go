//author:zsx
//data:2016-11-2
//GetAccounts需要写入区块，只测试了空的
//getbanlance 还未写入 只测试了空的，但是空的账户应该返回账户空或者余额0，现在出现了空指针异常，分析代码后暂认为是逻辑有一点问题
package hpc

//import (
//	//	"fmt"
//	//	"hyperchain/accounts"
//	"hyperchain/consensus"
//	"hyperchain/core"
//	"hyperchain/core/blockpool"
//	"hyperchain/core/types"
//	"hyperchain/crypto"
//	"hyperchain/event"
//	//	"hyperchain/hyperdb"
//	//	"hyperchain/manager"
//	//	"strings"
//	"testing"
//	"time"
//)

//func Test_AccountAPI(t *testing.T) {
//	core.InitDB("./build/keystore", 8023)
//	keydir := "../config/keystore/"
//	encryption := crypto.NewEcdsaEncrypto("ecdsa")
//	am := accounts.NewAccountManager(keydir, encryption)
//	pm := &manager.ProtocolManager{

//		AccountManager: am,
//	}
//	db, _ := hyperdb.GetLDBDatabase()
//	fmt.Println("-----------------start test----------------------")
//	publicAccountAPI := NewPublicAccountAPI(pm, db)
//	ad := publicAccountAPI.NewAccount("123456")
//	if len(ad) == 0 {
//		t.Errorf("newAccount wrong")
//	}
//	unlockParas := &UnlockParas{
//		Address:  ad,
//		Password: "123456",
//	}
//	ref, _ := publicAccountAPI.UnlockAccount(*unlockParas)
//	if !ref {

//		t.Errorf("UnlockParas  wrong")
//	}
//	unlockParas2 := &UnlockParas{
//		Address:  ad,
//		Password: "1234567",
//	}

//	_, err := publicAccountAPI.UnlockAccount(*unlockParas2)
//	if !strings.EqualFold("Incorrect address or password!", err.Error()) {
//		fmt.Println(err.Error())
//		t.Errorf("UnlockParas  wrong")
//	}

//	ref_getaccount := publicAccountAPI.GetAccounts()
//	if ref_getaccount != nil {
//		t.Errorf("GetAccounts  wrong")
//	}
//	fmt.Println("getbalance testing")
//	//	ref_getbanlance, _ := publicAccountAPI.GetBalance(ad)
//	//	if !strings.EqualFold(ref_getbanlance, "") {
//	//	}
//	//	t.Errorf("GetBalanc wrong")
//}
