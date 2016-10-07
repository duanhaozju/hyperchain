package accounts

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"errors"
	"path/filepath"
	"hyperchain/crypto"
	"sync"
	"hyperchain/common"
	"time"
	"math/big"
	"os"
	"bufio"
	"fmt"
	"reflect"
)

var (
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
	ErrLocked  = errors.New("account is locked")
)

// Account represents a stored key.
// When used as an argument, it selects a unique key file to act on.
type Account struct {
	Address common.Address // Ethereum account address derived from the key

		       // File contains the key file name.
		       // When Acccount is used as an argument to select a key, File can be left blank to
		       // select just by address or set to the basename or absolute path of a file in the key
		       // directory. Accounts returned by AccountManager will always contain an absolute path.
	File string
}

// AccountAccountManager manages a key storage directory on disk.
type AccountManager struct {
	KeyStore	keyStore
	Encryption	crypto.Encryption
	mu		sync.RWMutex
	Unlocked map[common.Address]*unlocked
}
type unlocked struct {
	*Key
	abort chan struct{}
}
// NewAccountManager creates a AccountManager for the given directory.
func NewAccountManager(keydir string,encryp crypto.Encryption) *AccountManager {
	keydir, _ = filepath.Abs(keydir)
	am := &AccountManager{
		KeyStore: &keyStorePassphrase{keydir, StandardScryptN, StandardScryptP},
		Unlocked:make(map[common.Address]*unlocked),
		Encryption:encryp,
	}

	//am.unlockAllAccount(keydir)
	return am
}
func (am *AccountManager)UnlockAllAccount(keydir string){
	var accounts []Account
	accounts = getAllAccount(keydir)
	for _,a := range accounts{
		am.Unlock(a,"123")
	}

}
func getAllAccount(keydir string) []Account {
	var accounts []Account
	addressdir := keydir+"addresses/address"
	fp, _ := os.Open(addressdir)
	scanner := bufio.NewScanner(fp)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		addrHex := scanner.Text()
		//fmt.Println(data)
		addr := common.HexToAddress(string(addrHex)[:40])
		account := Account{
			Address:addr,
			File:keydir+addrHex,
		}
		accounts = append(accounts,account)
	}
	fp.Close()
	return accounts
}

// Sign signs hash with an unlocked private key matching the given address.
func (am *AccountManager) Sign(addr common.Address, hash []byte) (signature []byte, err error) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	unlockedKey, found := am.Unlocked[addr]
	if !found {
		return nil, ErrLocked
	}
	return am.Encryption.Sign(hash, unlockedKey.PrivateKey)
}
// SignWithPassphrase signs hash if the private key matching the given address can be
// decrypted with the given passphrase.
func (am *AccountManager) SignWithPassphrase(addr common.Address, hash []byte, passphrase string) (signature []byte, err error) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	unlockedKey, found := am.Unlocked[addr]
	if !found {
		file := am.KeyStore.JoinPath(addr.Hex()[2:])
		key, err := am.GetDecryptedKey(Account{Address: addr,File:file}, passphrase)
		if err != nil {
			return nil, err
		}
		unlockedKey = &unlocked{Key: key, abort: make(chan struct{})}
		//switch key.PrivateKey.(type) {
		//case *ecdsa.PrivateKey:
		//	actualPriKey := key.PrivateKey.(*ecdsa.PrivateKey)
		//	unlockedKey.Key = &Key{
		//		Address:    crypto.PubkeyToAddress(actualPriKey.PublicKey),
		//		PrivateKey: actualPriKey,
		//	}
		//	defer zeroKey(actualPriKey)
		//}
	}

	return am.Encryption.Sign(hash, unlockedKey.PrivateKey)
}
// Unlock unlocks the given account indefinitely.
func (am *AccountManager) Unlock(a Account, passphrase string) error {
	duration := big.NewInt(36000)
	d := time.Duration(duration.Int64()) * time.Second
	return am.TimedUnlock(a, passphrase, d)
}

// Lock removes the private key with the given address from memory.
func (am *AccountManager) Lock(addr common.Address) error {
	am.mu.Lock()
	if unl, found := am.Unlocked[addr]; found {
		am.mu.Unlock()
		am.expire(addr, unl, time.Duration(0)*time.Nanosecond)
	} else {
		am.mu.Unlock()
	}
	return nil
}

// TimedUnlock unlocks the given account with the passphrase. The account
// stays unlocked for the duration of timeout. A timeout of 0 unlocks the account
// until the program exits. The account must match a unique key file.
//
// If the account address is already unlocked for a duration, TimedUnlock extends or
// shortens the active unlock timeout. If the address was previously unlocked
// indefinitely the timeout is not altered.
func (am *AccountManager) TimedUnlock(a Account, passphrase string, timeout time.Duration) error {
	key, err := am.GetDecryptedKey(a,passphrase)
	if err != nil {
		fmt.Println(err)
		return err
	}
	am.mu.Lock()
	defer am.mu.Unlock()
	u, found := am.Unlocked[a.Address]
	if found {
		if u.abort == nil {
			// The address was unlocked indefinitely, so unlocking
			// it with a timeout would be confusing.
			zeroKey(key.PrivateKey.(*ecdsa.PrivateKey))
			return nil
		} else {
			// Terminate the expire goroutine and replace it below.
			close(u.abort)
		}
	}
	if timeout > 0 {
		u = &unlocked{Key: key, abort: make(chan struct{})}
		go am.expire(a.Address, u, timeout)
	} else {
		u = &unlocked{Key: key}
	}
	am.Unlocked[a.Address] = u
	return nil
}
func (am *AccountManager) GetDecryptedKey(a Account,auth string) (*Key, error) {
	//am.mu.Lock()
	//defer am.mu.Unlock()
	key, err := am.KeyStore.GetKey(a.Address, a.File, auth)
	return key, err
}

func (am *AccountManager) expire(addr common.Address, u *unlocked, timeout time.Duration) {
	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case <-u.abort:
	// just quit
	case <-t.C:
		am.mu.Lock()
	// only drop if it's still the same key instance that dropLater
	// was launched with. we can check that using pointer equality
	// because the map stores a new pointer every time the key is
	// unlocked.
		if am.Unlocked[addr] == u {
			zeroKey(u.PrivateKey.(*ecdsa.PrivateKey))
			delete(am.Unlocked, addr)
		}
		am.mu.Unlock()
	}
}
// NewAccount generates a new key and stores it into the key directory,
// encrypting it with the passphrase.
func (am *AccountManager) NewAccount(passphrase string) (Account, error) {
	_, account, err := storeNewKey(am, crand.Reader, passphrase)
	if err != nil {
		return Account{}, err
	}
	//am.AddrPassMap[common.ToHex(account.Address)] = passphrase
	storeNewAddrToFile(account)
	//am.Unlock(account,passphrase)
	return account, nil
}

// zeroKey zeroes a private key in memory.
func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

func ValidateAddr(from []byte) bool {
	var accounts []Account
	keydir,_ := filepath.Abs("./keystore")
	accounts = getAllAccount(keydir)
	addr := common.HexToAddress(string(from))
	ac := Account{
		Address:addr,
		File:keydir +"/" +string(from),
	}
	for _,account := range accounts{
		if(reflect.DeepEqual(ac,account)){
			return true
		}
	}
	//if _,found := am.unlocked[addr];found{
	//	return true
	//}
	return false
}
