//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package accounts

import (
	"testing"
	"fmt"
	"hyperchain/crypto"
	"hyperchain/common"
	"sync/atomic"
	"math/big"
	"time"
	//"hyperchain/hyperdb"
	//"hyperchain/core"
	//"hyperchain/core"
	//"hyperchain/hyperdb"
)

type Transaction struct {
	data txdata
	// caches
	from atomic.Value
}
type txdata struct  {
	Recipient *common.Address
	Amount *big.Int
	signature []byte
}

func TestManager(t *testing.T)  {

	keydir := "../config/keystore/"

	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := NewAccountManager(keydir,encryption)

	//account,err := am.NewAccount("123")
	//if err!=nil{
	//	t.Error(err)
	//	t.FailNow()
	//}
	//fmt.Println("------new account------")
	//fmt.Println(account.Address)
	//fmt.Println(common.ToHex(account.Address[:]))
	address := common.HexToAddress("0x6201cb0448964ac597faf6fdf1f472edf2a22b89")
	fmt.Println("------get key according to the given account------")
	ac := Account{
		Address:address,
		File:am.KeyStore.JoinPath(KeyFileName(address[:])),
	}
	//fmt.Println(ac)
	time1 := time.Now()
	key,_ := am.GetDecryptedKey(ac,"123")
	fmt.Println(time.Since(time1))
	fmt.Println(key.Address)
	fmt.Println(common.ToHex(key.Address[:]))
	//fmt.Println(hex.EncodeToString(key.Address[:]))
	//fmt.Println(key.PrivateKey)
	//
}

//func TestNewAccount(t *testing.T) {
//	keydir := "../keystore/"
//
//	encryption := crypto.NewEcdsaEncrypto("ecdsa")
//	am := NewAccountManager(keydir,encryption)
//	am.NewAccount("123")
//}

//func TestSigntx(t *testing.T) {
//	keydir := "../keystore/"
//	encryption := crypto.NewEcdsaEncrypto("ecdsa")
//
//	am := NewAccountManager(keydir,encryption)
//
//	core.InitDB(8082)
//	db ,_ := hyperdb.GetLDBDatabase()
//	height := core.GetHeightOfChain()
//	block ,_ := core.GetBlockByNumber(db,height)
//	tx := block.Transactions[1]
//
//	kec256Hash := crypto.NewKeccak256Hash("keccak256")
//
//	hash := tx.SighHash(kec256Hash)
//	ac:=Account{
//		Address:common.HexToAddress(string(tx.From)),
//		File:am.KeyStore.JoinPath(string(tx.From)),
//	}
//
//	am.Unlock(ac,"123")
//	var set = []int{500,5000,10000}
//
//	for _,j:=range set{
//		start := time.Now()
//		for i:=0;i<j;i++{
//			am.SignWithPassphrase(common.HexToAddress(string(tx.From)),hash[:],"123")
//			//fmt.Println(signature)
//			//tx.ValidateSign(encryption,kec256Hash)
//		}
//		fmt.Printf("signtx test %dtxs: %s",j,time.Since(start))
//		fmt.Println()
//
//	}
//	for _,j:=range set{
//		start := time.Now()
//		for i:=0;i<j;i++{
//			//am.SignWithPassphrase(common.HexToAddress(string(tx.From)),hash[:],"123")
//			tx.ValidateSign(encryption,kec256Hash)
//		}
//		fmt.Printf("unsigntx test %dtxs: %s",j,time.Since(start))
//		fmt.Println()
//
//	}
//
//}