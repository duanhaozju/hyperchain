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
