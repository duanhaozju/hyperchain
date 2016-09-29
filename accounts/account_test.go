/**
 * Created by Meiling Hu on 9/2/16.
 */
package accounts

import (
	"testing"
	"fmt"
	"hyperchain/crypto"
	"hyperchain/common"
	"sync/atomic"
	"math/big"
	"time"
	"hyperchain/hyperdb"
	"hyperchain/core"
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
func NewTransaction(to common.Address,amount *big.Int) *Transaction {
	d:=txdata{
		Recipient:	&to,
		Amount:		new(big.Int),
	}
	if amount != nil{
		d.Amount.Set(amount)
	}
	return &Transaction{data:d}
}

func TestManager(t *testing.T)  {

	keydir := "../keystore/"

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
	//ac := Account{
	//	Address:account.Address,
	//	File:am.KeyStore.JoinPath(KeyFileName(account.Address)),
	//}
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
	////签名交易
	//tx:= NewTransaction(common.Address{},big.NewInt(100))
	//s256 := crypto.NewKeccak256Hash("Keccak256")
	//hash := s256.Hash([]interface{}{tx.data.Amount,tx.data.Recipient})
	//signature,err := am.Encryption.Sign(hash[:],key.PrivateKey)
	//
	//if err != nil {
	//	t.Error(err)
	//	t.FailNow()
	//
	//}
	////验证签名
	//from,err:= am.Encryption.UnSign(hash[:],signature)
	//if err != nil {
	//	t.Error(err)
	//	t.FailNow()
	//}
	//
	//fmt.Println(from)
	//fmt.Println(common.ToHex(from[:]))
	//fmt.Println(am.Unlocked[common.HexToAddress("6201cb0448964ac597faf6fdf1f472edf2a22b89")].PrivateKey)

}
func TestValidateAddr(t *testing.T) {
	//keydir := "../keystore/"
	//
	//encryption := crypto.NewEcdsaEncrypto("ecdsa")
	//am := NewAccountManager(keydir,encryption)
	start:=time.Now()
	from := []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89")
	fmt.Println(ValidateAddr(from))
	fmt.Println(time.Since(start))
}
func TestNewAccount(t *testing.T) {
	keydir := "../keystore/"

	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := NewAccountManager(keydir,encryption)
	am.NewAccount("123")
}
func TestSigntx(t *testing.T) {
	keydir := "./keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")

	am := NewAccountManager(keydir,encryption)

	core.InitDB(8082)
	db ,_ := hyperdb.GetLDBDatabase()
	height := core.GetHeightOfChain()
	block ,_ := core.GetBlockByNumber(db,height)
	tx := block.Transactions[1]

	kec256Hash := crypto.NewKeccak256Hash("keccak256")

	hash := tx.SighHash(kec256Hash)
	for i:=0;i<10;i++{
		start := time.Now()
		signature, _ := am.SignWithPassphrase(common.HexToAddress(string(tx.From)),hash[:],"123")
		fmt.Println(time.Since(start))
		fmt.Println(signature)
		tx.ValidateSign(encryption,kec256Hash)

	}
}