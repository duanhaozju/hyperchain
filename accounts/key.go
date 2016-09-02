/**
 * Created by Meiling Hu on 8/31/16.
 */
package accounts

import (
	"encoding/json"
	"os"
	"fmt"
	"path/filepath"
	"io/ioutil"
)

type Key struct {
	Auth string
	Address string
	Priv interface{}
	PrivateKey interface{}
}

type KeyStore interface {
	//loads the key from disk
	GetKey(file string,address string,auth string)(*Key,error)
	//stores the key to disk with auth together for binding
	StoreKey(dir string,k *Key) error
}
type keyStorePlain struct {
	keysDirPath string
}

func (ks keyStorePlain)GetKey(file string,address string,auth string)(*Key,error)  {
	fd,err := os.Open(file)
	if err!=nil{
		return nil,err
	}
	defer fd.Close()
	key := new(Key)
	if err := json.NewDecoder(fd).Decode(key); err != nil {
		return nil, err
	}
	if key.Address != address {
		return nil, fmt.Errorf("key content mismatch: have address %x, want %x", key.Address, address)
	}
	if key.Auth != auth {
		return nil, fmt.Errorf("key content mismath:  have auth %x, want %x",key.Auth,auth)
	}
	return key, nil
}

func (ks keyStorePlain)StoreKey(dir string,key *Key)error  {

	content,err := json.Marshal(key)
	if err != nil {
		return err
	}
	file := dir + key.Address
	return ks.writeKeyFile(file,content)
}

func (ks keyStorePlain)writeKeyFile(file string, content []byte) error {
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
	if _,err := f.Write(content);err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}


