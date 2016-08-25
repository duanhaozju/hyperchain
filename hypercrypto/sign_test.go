package hyperencrypt

import (
	"testing"
	"math/big"
	"fmt"
	"crypto/elliptic"
	"hyperchain-alpha/hypercrypto/secp256k1"
	"hyperchain-alpha/common"
)

func TestSigntx(t *testing.T)  {
	key, err := GenerateKey()
	//SaveECDSA("./testFile",key)
	//priv,err:=LoadECDSA("./testFile")
	pub := key.PublicKey
	priv := key

	var addr common.Address
	//addr = thecrypto.PubkeyToAddress(priv.PublicKey)
	pubBytes := elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
	addr = common.BytesToAddress(Keccak256(pubBytes[1:])[12:])

	fmt.Println("public key is :")
	fmt.Println(pub)
	fmt.Println("private key is :")
	fmt.Println(key)

	//签名交易
	tx,_ := NewTransaction(common.Address{},big.NewInt(100)).SignECDSA(priv)
	//验证签名
	from, err := tx.From()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(from.Hex())

	fmt.Println(from)
	fmt.Println(addr)


}
