//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"crypto/rand"
	"errors"
	"github.com/hyperchain/hyperchain/common"
	"math/big"
)

var testAddrHex = "970e8128ab834e8eac17ab8e3812f010678cf791"
var testPrivHex = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
var encryption = NewEcdsaEncrypto("ecdsa")

// These tests are sanity checks.
// They should ensure that we don't e.g. use Sha3-224 instead of Sha3-256
// and that the sha3 library uses keccak-f permutation.

func BenchmarkSha3(b *testing.B) {
	a := []byte("hello world")
	amount := 1000000
	start := time.Now()
	for i := 0; i < amount; i++ {
		Keccak256(a)
	}

	fmt.Println(amount, ":", time.Since(start))
}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	if len(b) != 32 {
		return nil, errors.New("invalid length, need 256 bits")
	}
	return ToECDSA(b), nil
}

func TestSign(t *testing.T) {
	key, _ := HexToECDSA(testPrivHex)
	addr := common.HexToAddress(testAddrHex)

	msg := Keccak256([]byte("foo"))
	sig, err := encryption.Sign(msg, key)
	if err != nil {
		t.Errorf("Sign error: %s", err)
	}
	recoveredAddr, err := encryption.UnSign(msg, sig)
	if err != nil {
		t.Errorf("UnSign error: %s", err)
	}
	if addr != recoveredAddr {
		t.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr)
	}

	pri2Addr := encryption.PrivKeyToAddress(*key)
	if addr != pri2Addr {
		t.Errorf("Address mismatch: want: %x have: %x", addr, pri2Addr)
	}

}

func TestInvalidSign(t *testing.T) {
	_, err := encryption.Sign(make([]byte, 1), &ecdsa.PrivateKey{})
	if err == nil {
		t.Errorf("expected sign with hash 1 byte to error")
	}

	_, err = encryption.Sign(make([]byte, 33), &ecdsa.PrivateKey{})
	if err == nil {
		t.Errorf("expected sign with hash 33 byte to error")
	}
}
func TestEcdsaEncrypto_GenerateNodeKey(t *testing.T) {
	id := "id1"
	nodeDir := "/tmp/hyperchain/nodedir"
	defer os.Remove(nodeDir)
	err := encryption.GenerateNodeKey(id, nodeDir)
	if err != nil {
		t.Error("generate node key error")
	}
}

func TestLoadECDSAFile(t *testing.T) {
	keyBytes := common.FromHex(testPrivHex)
	fileName0 := "test_key0"
	fileName1 := "test_key1"
	checkKey := func(k *ecdsa.PrivateKey) {
		checkAddr(t, PubkeyToAddress(k.PublicKey), common.HexToAddress(testAddrHex))
		loadedKeyBytes := FromECDSA(k)
		if !bytes.Equal(loadedKeyBytes, keyBytes) {
			t.Fatalf("private key mismatch: want: %x have: %x", keyBytes, loadedKeyBytes)
		}
	}

	ioutil.WriteFile(fileName0, []byte(testPrivHex), 0600)
	defer os.Remove(fileName0)

	key0, err := LoadECDSA(fileName0)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key0)

	// again, this time with SaveECDSA instead of manual save:
	err = SaveECDSA(fileName1, key0)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fileName1)

	key1, err := LoadECDSA(fileName1)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key1)
}

func checkhash(t *testing.T, name string, f func([]byte) []byte, msg, exp []byte) {
	sum := f(msg)
	if bytes.Compare(exp, sum) != 0 {
		t.Fatalf("hash %s mismatch: want: %x have: %x", name, exp, sum)
	}
}

func checkAddr(t *testing.T, addr0, addr1 common.Address) {
	if addr0 != addr1 {
		t.Fatalf("address mismatch: want: %x have: %x", addr0, addr1)
	}
}

func TestSignTimeWithCryptoEcdsa(t *testing.T) {
	key, _ := GenerateKey()

	ee := NewEcdsaEncrypto("ecsda")
	var s256 = NewKeccak256Hash("Keccak256")

	for i := 1; i < 8; i++ {
		length, err := rand.Int(rand.Reader, big.NewInt(1024))
		if err != nil {
			t.Fatalf("Failed generating AES key [%s]", err)
		}
		msg, err := GetRandomBytes(int(length.Int64()) + 1)
		//fmt.Println("length of msg:",len(msg))
		if err != nil {
			t.Fatalf("Failed generating AES key [%s]", err)
		}
		hash := s256.Hash(msg)
		msg = hash[:]

		start := time.Now()
		sigma, err := ee.Sign(msg, key)
		if err != nil {
			t.Error("sign error")
		}
		fmt.Println("---sign with time---:", time.Since(start))
		address := ee.PrivKeyToAddress(*key)
		start = time.Now()
		check, err := ee.UnSign(msg, sigma)
		if err != nil {
			t.Error("unsign error")
		}
		if check == address {
			fmt.Println("---unsign with time---:", time.Since(start))
		}
		start = time.Now()
		r, s, err1 := ecdsa.Sign(rand.Reader, key, msg)
		if err1 != nil {
			t.Error("ecdsa sign error")
		}
		fmt.Println("***ecdsa sign time***:", time.Since(start))
		pubkey := key.PublicKey
		start = time.Now()
		check1 := ecdsa.Verify(&pubkey, msg, r, s)
		if check1 {
			fmt.Println("***ecdsa unsign time***:", time.Since(start))
		}
	}

}
func GetRandomBytes(len int) ([]byte, error) {
	key := make([]byte, len)

	// TODO: rand could fill less bytes then len
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}
