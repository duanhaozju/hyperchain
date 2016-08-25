package crypto

import (
	"testing"
	"math/big"
	"fmt"
	"crypto/elliptic"
	"hyperchain-alpha/common"
	"hyperchain-alpha/crypto/secp256k1"
)

func TestSigntx(t *testing.T)  {
	ee := NewEcdsaEncrypto("ECDSAEncryto")
	key, err := ee.GenerateKey()
	ee.SaveECDSA("./testFile",key)
	priv,err:=ee.LoadECDSA("./testFile")
	pub := key.PublicKey
	//priv := key


	var addr common.Address
	//addr = thecrypto.PubkeyToAddress(priv.PublicKey)
	pubBytes := elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
	addr = common.BytesToAddress(ee.Keccak256(pubBytes[1:])[12:])

	fmt.Println("public key is :")
	fmt.Println(pub)
	fmt.Println("private key is :")
	fmt.Println(key)

	//签名交易
	tx:= NewTransaction(common.Address{},big.NewInt(100))
	hash := tx.SigHash()
	signature,_ := ee.Sign(hash[:],*priv)
	signedTx, _ := tx.WithSignature(signature)

	//验证签名
	from, err := signedTx.From()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(from.Hex())

	fmt.Println(from)
	fmt.Println(addr)
	
	addr_from_unsign,_ := ee.UnSign(hash,signature)
	fmt.Println(addr_from_unsign)


}
