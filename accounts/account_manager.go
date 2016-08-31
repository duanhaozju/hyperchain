/**
 * Created by Meiling Hu on 8/31/16.
 */
package accounts

import (
	"hyperchain/crypto"
	"hyperchain/common"
)

// Manager manages a key storage directory on disk.
type Manager struct {
	keyStore KeyStore
	encryption *crypto.EcdsaEncrypto
}

const keydir = "/tmp/hyperchain/cache/keystore/"

func NewManager() *Manager {
	am := &Manager{
		keyStore: &keyStorePlain{},
		encryption: crypto.NewEcdsaEncrypto("ecdsa"),
	}
	return am
}

func (am *Manager)NewAccount(passphrase string) (*Key,error) {
	priv,err := am.encryption.GenerateKey()
	if err!=nil{
		return nil,err
	}
	key := &Key{
		auth:passphrase,
		address:am.encryption.PrivKeyToAddress(*priv),
		priv:priv,
	}
	if err :=am.keyStore.StoreKey(keydir,key);err!=nil{
		return key,err
	}
	return key,nil
}

func (am *Manager)GetAccountKey(address []byte,auth string)(*Key,error)  {
	filename := keydir + common.ToHex(address)
	return am.keyStore.GetKey(filename,address,auth)
}
