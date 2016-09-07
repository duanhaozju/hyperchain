package core

import (
	"testing"
	"fmt"
	"hyperchain/crypto"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/accounts"
	"crypto/ecdsa"
)

func TestUnsign(t *testing.T) {

	encryption :=crypto.NewEcdsaEncrypto("ecdsa")
	scryptN := accounts.StandardScryptN
	scryptP := accounts.StandardScryptP
	keydir := "./keystore/"
	am := accounts.NewAccountManager(keydir,encryption, scryptN, scryptP)
	key1,_ := am.GetDecryptedKey(accounts.Account{Address:common.FromHex("0x6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		File:"/home/fox/gohome/src/hyperchain/keystore/6201cb0448964ac597faf6fdf1f472edf2a22b89"})
	key2,_ := am.GetDecryptedKey(accounts.Account{Address:common.FromHex("0xb18c8575e3284e79b92100025a31378feb8100d6"),
		File:"/home/fox/gohome/src/hyperchain/keystore/b18c8575e3284e79b92100025a31378feb8100d6"})
	prv1:=key1.PrivateKey.(*ecdsa.PrivateKey)
	prv2:=key2.PrivateKey.(*ecdsa.PrivateKey)
	fmt.Println(crypto.PubkeyToAddress(prv1.PublicKey))
	fmt.Println(crypto.PubkeyToAddress(prv2.PublicKey))

	from1:=common.BytesToAddress(common.FromHex("0x6201cb0448964ac597faf6fdf1f472edf2a22b89"))
	from2:=common.BytesToAddress(common.FromHex("0xb18c8575e3284e79b92100025a31378feb8100d6"))

	tx1 := types.NewTransaction(from1[:],[]byte{},[]byte{1})
	tx2 := types.NewTransaction(from2[:],[]byte{},[]byte{1})

	kec256Hash:=crypto.NewKeccak256Hash("keccak256")

	h1:=kec256Hash.Hash(tx1)
	h2:=kec256Hash.Hash(tx2)

	sign1,_ := encryption.Sign(h1[:],prv1)
	sign2,_:= encryption.Sign(h2[:],prv2)

	tx1.Signature = sign1
	tx2.Signature = sign2

	//f1,_:=encryption.UnSign(h1[:],sign1)
	tx2.From = tx1.From
	//h3:=kec256Hash.Hash(tx2)
	ff,_:=encryption.UnSign(h2[:],sign2)
	fmt.Println(ff)
}
