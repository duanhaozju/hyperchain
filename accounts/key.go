// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package accounts

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"hyperchain/crypto"
)

const (
	version = 3
)

type Key struct {
	Address []byte
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	PrivateKey interface{}
}

type keyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr []byte, filename string, auth string) (*Key, error)
	// Writes and encrypts the key.
	StoreKey(filename string, k *Key, auth string) error
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}

type plainKeyJSON struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privatekey"`
	Version    int    `json:"version"`
}

type encryptedKeyJSONV3 struct {
	Address string     `json:"address"`
	Crypto  cryptoJSON `json:"crypto"`
	Version int        `json:"version"`
}

type encryptedKeyJSONV1 struct {
	Address string     `json:"address"`
	Crypto  cryptoJSON `json:"crypto"`
	Version string     `json:"version"`
}

type cryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

type scryptParamsJSON struct {
	N     int    `json:"n"`
	R     int    `json:"r"`
	P     int    `json:"p"`
	DkLen int    `json:"dklen"`
	Salt  string `json:"salt"`
}

func newKey(am *Manager,rand io.Reader) (*Key, error) {
	privKey, err := am.encryption.GeneralKey("123")
	if err != nil {
		return nil, err
	}
	switch privKey.(type) {
	case *ecdsa.PrivateKey:
		actualKey := privKey.(*ecdsa.PrivateKey)
		key := &Key{
			Address:    crypto.PubkeyToAddress(actualKey.PublicKey),
			PrivateKey: actualKey,
		}
		return key,nil
	}
	return nil,nil
}
func storeNewAddrToFile(a Account) error {
	addr := hex.EncodeToString(a.Address)+"\n"
	dir := filepath.Dir(a.File)
	file := dir+"/address"
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR,0600)
	if err != nil {
		return err
	}
	n, err := f.Write([]byte(addr))
	if err == nil && n < len([]byte(addr)) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
func storeNewKey(am *Manager, rand io.Reader, auth string) (*Key, Account, error) {
	key, err := newKey(am,rand)
	if err != nil {
		return nil, Account{}, err
	}
	switch key.PrivateKey.(type) {
	case *ecdsa.PrivateKey:
		a := Account{Address: key.Address, File:am.keyStore.JoinPath(keyFileName(key.Address))}
		if err := am.keyStore.StoreKey(a.File, key, auth); err != nil {
			zeroKey(key.PrivateKey.(*ecdsa.PrivateKey))
			return nil, a, err
		}
		return key, a, err
	}
	return nil,Account{},nil
}

func writeKeyFile(file string, content []byte) error {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}

// keyFileName implements the naming convention for keyfiles:
// UTC--<created_at UTC ISO8601>-<address hex>
func keyFileName(keyAddr []byte) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr))
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}
