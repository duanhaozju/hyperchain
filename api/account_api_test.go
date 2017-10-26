//author:zsx
//data:2016-11-2

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

//测试成功创建一个新的account
func Test_NewAccount1(t *testing.T) {
	//单例数据库状态设置为close
	hyperdb.Setclose()
	//初始化数据
	core.InitDB("./build/keystore", 8001)

	//初始化AccountManager
	keydir := "../config/keystore/"
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

	ad, err := publicAccountAPI.NewAccountAPI("123456")
	if err != nil {
		t.Errorf("NewAccountAPI(123456) fail")
	}

	if len(ad) != 20 {
		log.Noticef("publicAccountAPI.NewAccountAPI fail，创建一个新账户失败")
		log.Noticef("新账户地址：%d", len(ad))
		fmt.Println(ad)
	}

}
func Test_UnlockAccount(t *testing.T) {

	//初始化AccountManager
	keydir := "../config/keystore/"
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

	ad, _ := publicAccountAPI.NewAccountAPI("123456")

	unlockParas := &UnlockParas{
		Address:  ad,
		Password: "123456",
	}

	unlockParas2 := &UnlockParas{
		Address:  ad,
		Password: "1234567",
	}

	//用123456解锁
	ref, _ := publicAccountAPI.UnlockAccount(*unlockParas)
	if !ref {
		t.Errorf("publicAccountAPI.UnlockAccount fail 解锁失败")
	}

	//用错误密码解锁
	ref, _ = publicAccountAPI.UnlockAccount(*unlockParas2)
	if ref {
		t.Errorf("publicAccountAPI.UnlockAccount fail 错误密码解锁成功")
	}
}

func Test_GetAccounts(t *testing.T) {

	//初始化AccountManager
	keydir := "../config/keystore/"
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

	//从有效的build找所有账户
	ref := publicAccountAPI.GetAccounts()
	if ref == nil {
		t.Errorf("publicAccountAPI.GetAccounts() fail 不能获取账户 ")
	}
}

//测试成功获得账户余额
func Test_GetBalance1(t *testing.T) {

	//初始化AccountManager
	keydir := "../config/keystore/"
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

	add := common.HexToAddress("0x0ed0dd439e44c140ab9fa6fc39459eaf20b2f5a7")
	ref, err := publicAccountAPI.GetBalance(add)

	if err != nil {
		fmt.Println(err)
		t.Errorf("publicAccountAPI.GetBalance1 fail 未能获得账户余额")
	}

	if ref != "0x640" {
		fmt.Println(ref)
		t.Errorf("publicAccountAPI.GetBalance1 fail 账户余额错误")
	}

	//不存在的账户
	add1 := common.HexToAddress("0x0ed0dd439e44c140ab9fa6fc39459eaf20b2f444")

	ref, err = publicAccountAPI.GetBalance(add1)

	fmt.Printf(err.Error())
	if err == nil {
		t.Errorf("publicAccountAPI.GetBalance1 fail 空账户返回错误不对")
	}
}
