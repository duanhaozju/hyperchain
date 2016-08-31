/**
 * Created by Meiling Hu on 8/31/16.
 */
package accounts

import (
	"encoding/json"
	"os"
	"fmt"
	"hyperchain/common"
	"io"
	"path/filepath"
)

type Key struct {
	auth string
	address []byte
	priv interface{}
}

type KeyStore interface {
	//loads the key from disk
	GetKey(file string,address []byte,auth string)(*Key,error)
	//stores the key to disk with auth together for binding
	StoreKey(dir string,k *Key) error
}
type keyStorePlain struct {
	keysDirPath string
}

func (ks keyStorePlain)GetKey(file string,address []byte,auth string)(*Key,error)  {
	fmt.Println("address param")
	fmt.Println(address)
	fd,err := os.Open(file)
	if err!=nil{
		return nil,err
	}
	defer fd.Close()
	key := new(Key)
	if err := json.NewDecoder(fd).Decode(key); err != nil {
		return nil, err
	}
	fmt.Println("getkey-------------")
	fmt.Println(key.address)
	fmt.Println(key.auth)
	//if key.address != address {
	//	return nil, fmt.Errorf("key content mismatch: have address %x, want %x", key.address, address)
	//}
	if key.auth != auth {
		return nil, fmt.Errorf("key content mismath:  have auth %x, want %x",key.auth,auth)
	}
	return key, nil
}

func (ks keyStorePlain)StoreKey(dir string,key *Key)error  {

	content,err := json.Marshal(key)
	if err != nil {
		return err
	}
	result := make(map[string]interface{})
	json.Unmarshal(content,result)
	fmt.Println("~~~~~~~~~~~~~~~~~~~")
	fmt.Println(result)
	fmt.Println(result["auth"])
	address := common.ToHex(key.address)
	file := dir + address
	fmt.Println("============")
	fmt.Println(address)
	fmt.Println(key.auth)
	fmt.Println(key.priv)
	fmt.Println("content: ",content)
	return ks.writeKeyFile(file,content)
}

func (ks keyStorePlain)writeKeyFile(file string, content []byte) error {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	fmt.Println("write content:",content)
	const dirPerm = 0777
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return err
	}
	//// Atomic write: create a temporary hidden file first
	//// then move it into place. TempFile assigns mode 0600.
	//f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	//if err != nil {
	//	return err
	//}
	////if _, err := f.Write(content); err != nil {
	////	f.Close()
	////	os.Remove(f.Name())
	////	return err
	////}
	//if _,err := f.Write(content);err != nil {
	//	f.Close()
	//	os.Remove(f.Name())
	//	return err
	//}
	//f.Close()
	//return os.Rename(f.Name(), file)
	//_, error := os.Stat(filepath.Dir(file))
	//if error == nil || os.IsExist(error){
	//	fmt.Println("directory exists")
	//
	//}else {
	//	fmt.Println("no")
	//	os.MkdirAll(filepath.Dir(file),0777)
	//}
	//if err:=ioutil.WriteFile(file, content, 0600);err!=nil{
	//	fmt.Println("write error")
	//	return err
	//}
	//return nil
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR,0600)
	if err != nil {
		return err
	}
	n, err := f.Write(content)
	fmt.Println("length of content:",n)
	if err == nil && n < len(content) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}


