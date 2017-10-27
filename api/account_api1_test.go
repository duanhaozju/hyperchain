//author:zsx
//data:2016-11-10

package api

import (
	"fmt"
	"github.com/hyperchain/hyperchain/accounts"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core"

	"github.com/hyperchain/hyperchain/crypto"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/hyperchain/hyperchain/manager"
	"testing"
)

func Test_GetAccounts1(t *testing.T) {
	hyperdb.Setclose()

	core.InitDB("./build/keystore1", 8001)
	//初始化AccountManager
	keydir := "../config/keystore1/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := accounts.NewAccountManager(keydir, encryption)

	//初始化pm
	pm := &manager.EventHub{

		AccountManager: am,
	}

	//获取db句柄
	db, _ := hyperdb.GetDBDatabase()

	//初始化API
	publicAccountAPI := NewPublicAccountAPI(pm, db)

	//从无效的build找所有账户
	ref := publicAccountAPI.GetAccounts()
	if ref != nil {
		t.Errorf("publicAccountAPI.GetAccounts() fail 无效数据库返回账户信息不为空 ")
		fmt.Println("===========================")
		fmt.Println(ref)
	}
}

func Test_GetBalance(t *testing.T) {
	//初始化数据  keystore1无效数据库
	core.InitDB("./build/keystore1", 8001)

	//初始化AccountManager
	keydir := "../config/keystore1/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := accounts.NewAccountManager(keydir, encryption)

	//初始化pm
	pm := &manager.EventHub{

		AccountManager: am,
	}

	//获取db句柄
	db, _ := hyperdb.GetDBDatabase()

	//初始化API
	publicAccountAPI := NewPublicAccountAPI(pm, db)

	add := common.Address([20]byte{'1', '2', '3'})
	ref, err := publicAccountAPI.GetBalance(add)

	if err == nil {
		fmt.Println(ref)
		t.Errorf("publicAccountAPI.GetBalance fail 空地址未返回错误")
	}

}
