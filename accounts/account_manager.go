/**
 * Created by Meiling Hu on 8/31/16.
 */
package accounts

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"encoding/hex"
)

// Manager manages a key storage directory on disk.
type Manager struct {
	keyStore KeyStore
	encryption *crypto.EcdsaEncrypto
	addrPassMap map[string]string
}

const keydir = "/tmp/hyperchain/cache/keystore/"

func NewManager() *Manager {
	am := &Manager{
		keyStore: &keyStorePlain{},
		encryption: crypto.NewEcdsaEncrypto("ecdsa"),
		addrPassMap:make(map[string]string),
	}
	return am
}

func (am *Manager)NewAccount(passphrase string) (*Key,error) {
	priv,err := am.encryption.GenerateKey()
	if err!=nil{
		return nil,err
	}
	key := &Key{
		Auth:passphrase,
		Address:common.ToHex(am.encryption.PrivKeyToAddress(*priv)),
		Priv:hex.EncodeToString(priv.D.Bytes()),
		PrivateKey:priv,
	}
	if err :=am.keyStore.StoreKey(keydir,key);err!=nil{
		return key,err
	}
	am.addrPassMap[key.Address] = key.Auth
	return key,nil
}

//get the key matching the given address and passphrase
func (am *Manager)GetAccountKey(address string,auth string)(*Key,error)  {
	filename := keydir +address
	k,err :=am.keyStore.GetKey(filename,address,auth)
	if err!=nil{
		return nil,err
	}
	keybyte,_ := hex.DecodeString(k.Priv.(string))
	return &Key{
		Address:k.Address,
		Auth:k.Auth,
		Priv:k.Priv,
		PrivateKey:am.encryption.ToECDSA(keybyte),
	},nil
}
