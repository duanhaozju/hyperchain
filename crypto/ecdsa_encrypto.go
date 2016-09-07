package crypto

import (
	"fmt"
	"hyperchain/common"
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"os"
	"io"
	"encoding/hex"
	"io/ioutil"
	"crypto/elliptic"
	"hyperchain/crypto/sha3"
	"hyperchain/crypto/secp256k1"
)

const keystoredir  = "/tmp/hyperchain/cache/keystore/"
//
type EcdsaEncrypto struct{
	name string
	port string
}

func NewEcdsaEncrypto(name string) *EcdsaEncrypto  {
	ee := &EcdsaEncrypto{name:name}
	return ee
}


func GenerateKey()(*ecdsa.PrivateKey,error)  {
	return ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
}

func (ee *EcdsaEncrypto)Sign(hash []byte,  prv interface{})(sig []byte, err error)  {
	privateKey := prv.(*ecdsa.PrivateKey)
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}

	seckey := common.LeftPadBytes(privateKey.D.Bytes(), privateKey.Params().BitSize/8)
	defer zeroBytes(seckey)
	sig, err = secp256k1.Sign(hash, seckey)
	return
}

//UnSign recovers Address from txhash and signature
func (ee *EcdsaEncrypto)UnSign(args ...interface{})([]byte, error)  {
	if len(args)!=2{
		err :=errors.New("paramas invalid")
		return nil,err
	}
	hash := args[0].([]byte)
	sig := args[1].([]byte)
	pubBytes,err := secp256k1.RecoverPubkey(hash, sig)
	if err!=nil{
		return nil,err
	}
	addr := Keccak256(pubBytes[1:])[12:]
	return addr,nil
}
func (ee *EcdsaEncrypto)GeneralKey(port string)(interface{},error) {
	key,err := GenerateKey()
	if err!=nil{
		return nil,err
	}

	ee.port=port
	k := hex.EncodeToString(FromECDSA(key))
	_, error := os.Stat(keystoredir)
	if error == nil || os.IsExist(error){
		//("directory exists")

	}else {
		os.MkdirAll(keystoredir,0777)
	}
	file := keystoredir+port
	if err:=ioutil.WriteFile(file, []byte(k), 0600);err!=nil{
		return key,err
	}
	return key,nil

}
//load key by given port
func (ee *EcdsaEncrypto)GetKey() (interface{},error) {
	file := keystoredir+ee.port
	return LoadECDSA(file)
}

func (ee *EcdsaEncrypto)PrivKeyToAddress(prv interface {})[]byte  {
	p := prv.(ecdsa.PrivateKey)
	return PubkeyToAddress(p.PublicKey)
}
// LoadECDSA loads a secp256k1 private key from the given file.
// key data is expected to be hex-encoded.
func LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	return ToECDSA(key), nil
}
// New methods using proper ecdsa keys from the stdlib
func ToECDSA(prv []byte) *ecdsa.PrivateKey {
	if len(prv) == 0 {
		return nil
	}

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp256k1.S256()
	priv.D = common.BigD(prv)
	priv.PublicKey.X, priv.PublicKey.Y = secp256k1.S256().ScalarBaseMult(prv)
	return priv
}

func FromECDSA(prv *ecdsa.PrivateKey) []byte {
	if prv == nil {
		return nil
	}
	return prv.D.Bytes()
}
// SaveECDSA saves a secp256k1 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveECDSA(file string, key *ecdsa.PrivateKey) error {
	k := hex.EncodeToString(FromECDSA(key))
	return ioutil.WriteFile(file, []byte(k), 0600)
}
//SaveNodeInfo saves the info of node into local file
//ip addr and pri
func SaveNodeInfo(file string, port string ,addr []byte, pri *ecdsa.PrivateKey) error {
	prikey := hex.EncodeToString(FromECDSA(pri))
	content := port +" "+common.ToHex(addr)+" "+prikey+" \n"
	fmt.Println(content)

	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR,0600)
	if err != nil {
		return err
	}
	n, err := f.Write([]byte(content))
	if err == nil && n < len([]byte(content)) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err

}
func PubkeyToAddress(p ecdsa.PublicKey) []byte {
	pubBytes := FromECDSAPub(&p)
	return Keccak256(pubBytes[1:])[12:]
}

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}
func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}