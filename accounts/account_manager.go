// Copyright 2015 The go-ethereum Authors
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

// Package accounts implements encrypted storage of secp256k1 private keys.
//
// Keys are stored as encrypted JSON files according to the Web3 Secret Storage specification.
// See https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition for more information.
package accounts

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"errors"
	"path/filepath"
	"hyperchain/crypto"
	"hyperchain/common"
)

var (
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

// Account represents a stored key.
// When used as an argument, it selects a unique key file to act on.
type Account struct {
	Address []byte // Ethereum account address derived from the key

	// File contains the key file name.
	// When Acccount is used as an argument to select a key, File can be left blank to
	// select just by address or set to the basename or absolute path of a file in the key
	// directory. Accounts returned by Manager will always contain an absolute path.
	File string
}

// Manager manages a key storage directory on disk.
type Manager struct {
	keyStore keyStore
	encryption crypto.Encryption
	addrPassMap map[string]string
}

// NewManager creates a manager for the given directory.
func NewManager(keydir string,encryp crypto.Encryption, scryptN, scryptP int) *Manager {
	keydir, _ = filepath.Abs(keydir)
	am := &Manager{
		keyStore: &keyStorePassphrase{keydir, scryptN, scryptP},
		addrPassMap:make(map[string]string),
		encryption:encryp,
	}
	return am
}

func (am *Manager) GetDecryptedKey(a Account) (*Key, error) {
	key, err := am.keyStore.GetKey(a.Address, a.File, am.addrPassMap[common.ToHex(a.Address)])
	return key, err
}

// NewAccount generates a new key and stores it into the key directory,
// encrypting it with the passphrase.
func (am *Manager) NewAccount(passphrase string) (Account, error) {
	_, account, err := storeNewKey(am, crand.Reader, passphrase)
	if err != nil {
		return Account{}, err
	}
	am.addrPassMap[common.ToHex(account.Address)] = passphrase
	storeNewAddrToFile(account)
	return account, nil
}

// zeroKey zeroes a private key in memory.
func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}