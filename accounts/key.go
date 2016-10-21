

package accounts

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"hyperchain/crypto"
	"hyperchain/common"
	"github.com/op/go-logging"
)

const (
	version = 3
)

type Key struct {
	Address		common.Address
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	PrivateKey	interface{}
}

type keyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr common.Address, filename string, auth string) (*Key, error)
	//GetKeyFromCache(addr common.Address,keyjson *[]byte,auth string)(*Key,error)
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
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("key")
}
func newKey(am *AccountManager,rand io.Reader) (*Key, error) {
	privKey, err := am.Encryption.GeneralKey()
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
	addr := hex.EncodeToString(a.Address[:])+"\n"
	dir := filepath.Dir(a.File)
	file := dir+"/addresses/address"
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
func storeNewKey(am *AccountManager, rand io.Reader, auth string) (*Key, Account, error) {
	key, err := newKey(am,rand)
	if err != nil {
		return nil, Account{}, err
	}
	switch key.PrivateKey.(type) {
	case *ecdsa.PrivateKey:
		a := Account{Address: key.Address, File:am.KeyStore.JoinPath(KeyFileName(key.Address[:]))}
		if err := am.KeyStore.StoreKey(a.File, key, auth); err != nil {
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
	const dirPerm = 0777
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
func KeyFileName(keyAddr []byte) string {
	//ts := time.Now().UTC()
	//return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr))
	return fmt.Sprintf("%s", hex.EncodeToString(keyAddr))
}

//func toISO8601(t time.Time) string {
//	//var tz string
//	//name, offset := t.Zone()
//	//if name == "UTC" {
//	//	tz = "Z"
//	//} else {
//	//	tz = fmt.Sprintf("%03d00", offset/3600)
//	//}
//	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
//}
