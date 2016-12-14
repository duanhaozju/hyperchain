//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package accounts

import (
	"testing"
	"hyperchain/crypto"
	"hyperchain/common"
	"sync/atomic"
	"math/big"
	"io/ioutil"
	"os"
	"time"
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

var testSigData = make([]byte, 32)

func tmpManager(t *testing.T) (string,*AccountManager) {

	encryption := crypto.NewEcdsaEncrypto("ecdsa")

	dir, err := ioutil.TempDir("/tmp","keystore-test")
	if err != nil {
		t.Fatal(
			err)
	}
	//fmt.Println(dir)
	am:= NewAccountManager(dir,encryption)
	return dir,am
}

func TestSign(t *testing.T) {
	dir, am := tmpManager(t)
	defer os.RemoveAll(dir)

	pass := "" // not used but required by API
	a1, err := am.NewAccount(pass)
	if err != nil {
		t.Fatal(err)
	}
	if err := am.Unlock(a1, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := am.Sign(a1.Address, testSigData); err != nil {
		t.Fatal(err)
	}
}

func TestSignWithPassphrase(t *testing.T) {
	dir, am := tmpManager(t)
	defer os.RemoveAll(dir)

	pass := "passwd"
	acc, err := am.NewAccount(pass)
	if err != nil {
		t.Fatal(err)
	}

	if _, unlocked := am.Unlocked[acc.Address]; unlocked {
		t.Fatal("expected account to be locked")
	}

	_, err = am.SignWithPassphrase(acc.Address,testSigData,pass)
	if err != nil {
		t.Fatal(err)
	}

	if _, unlocked := am.Unlocked[acc.Address]; unlocked {
		t.Fatal("expected account to be locked")
	}

	if _, err = am.SignWithPassphrase(acc.Address,testSigData, "invalid passwd"); err == nil {
		t.Fatal("expected SignHash to fail with invalid password")
	}
}

func TestTimedUnlock(t *testing.T) {
	dir, am := tmpManager(t)
	defer os.RemoveAll(dir)

	pass := "foo"
	a1, err := am.NewAccount(pass)

	// Signing without passphrase fails because account is locked
	_, err = am.Sign(a1.Address, testSigData)
	if err != ErrLocked {
		t.Fatal("Signing should've failed with ErrLocked before unlocking, got ", err)
	}

	// Signing with passphrase works
	if err = am.TimedUnlock(a1, pass, 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	// Signing without passphrase works because account is temp unlocked
	_, err = am.Sign(a1.Address, testSigData)
	if err != nil {
		t.Fatal("Signing shouldn't return an error after unlocking, got ", err)
	}

	// Signing fails again after automatic locking
	time.Sleep(250 * time.Millisecond)
	_, err = am.Sign(a1.Address, testSigData)
	if err != ErrLocked {
		t.Fatal("Signing should've failed with ErrLocked timeout expired, got ", err)
	}
}

func TestAccountManager_Lock(t *testing.T) {
	dir, am := tmpManager(t)
	defer os.RemoveAll(dir)

	pass := "foo"
	a1, err := am.NewAccount(pass)

	// Signing without passphrase fails because account is locked
	_, err = am.Sign(a1.Address, testSigData)
	if err != ErrLocked {
		t.Fatal("Signing should've failed with ErrLocked before unlocking, got ", err)
	}

	// Signing with passphrase works
	if err = am.TimedUnlock(a1, pass, 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	// Signing without passphrase works because account is temp unlocked
	_, err = am.Sign(a1.Address, testSigData)
	if err != nil {
		t.Fatal("Signing shouldn't return an error after unlocking, got ", err)
	}
	if err =am.Lock(a1.Address); err!=nil{
		t.Fatal(err)
	}
	_, err = am.Sign(a1.Address, testSigData)
	if err != ErrLocked {
		t.Fatal("Signing should've failed with ErrLocked timeout expired, got ", err)
	}
}
func TestAccountManager_UnlockAllAccount(t *testing.T) {
	keydir := "./config/keystore"
	_,am := tmpManager(t)
	am.UnlockAllAccount(keydir)

}
