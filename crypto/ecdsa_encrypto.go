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

type EcdsaEncrypto struct{
	name string
}

func NewEcdsaEncrypto(name string) *EcdsaEncrypto  {
	ee := &EcdsaEncrypto{name:name}
	return ee
}

func (ee *EcdsaEncrypto)GenerateKey()(*ecdsa.PrivateKey,error)  {
	return ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
}

func (ee *EcdsaEncrypto)Sign(hash []byte,  prv interface{})(sig []byte, err error)  {
	privateKey := prv.(ecdsa.PrivateKey)
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}

	seckey := common.LeftPadBytes(privateKey.D.Bytes(), privateKey.Params().BitSize/8)
	defer zeroBytes(seckey)
	sig, err = secp256k1.Sign(hash, seckey)
	return
}

//UnSign recovers commom.Address from txhash and signature
func (ee *EcdsaEncrypto)UnSign(args ...interface{})(common.Address, error)  {
	if len(args)!=2{
		err :=errors.New("paramas invalid")
		return common.Address{},err
	}
	hash := args[0].([]byte)
	sig := args[1].([]byte)
	pubBytes,err := secp256k1.RecoverPubkey(hash, sig)
	if err!=nil{
		return common.Address{},err
	}
	var addr common.Address
	copy(addr[:],ee.Keccak256(pubBytes[1:])[12:])
	return addr,nil
}

// LoadECDSA loads a secp256k1 private key from the given file.
// The key data is expected to be hex-encoded.
func (ee *EcdsaEncrypto)LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
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

	return ee.ToECDSA(key), nil
}
// New methods using proper ecdsa keys from the stdlib
func (ee *EcdsaEncrypto)ToECDSA(prv []byte) *ecdsa.PrivateKey {
	if len(prv) == 0 {
		return nil
	}

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp256k1.S256()
	priv.D = common.BigD(prv)
	priv.PublicKey.X, priv.PublicKey.Y = secp256k1.S256().ScalarBaseMult(prv)
	return priv
}

func (ee *EcdsaEncrypto)FromECDSA(prv *ecdsa.PrivateKey) []byte {
	if prv == nil {
		return nil
	}
	return prv.D.Bytes()
}
// SaveECDSA saves a secp256k1 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func (ee *EcdsaEncrypto)SaveECDSA(file string, key *ecdsa.PrivateKey) error {
	k := hex.EncodeToString(ee.FromECDSA(key))
	return ioutil.WriteFile(file, []byte(k), 0600)
}
//SaveNodeInfo saves the info of node into local file
//ip addr and pri
func (ee *EcdsaEncrypto)SaveNodeInfo(file string, ip string ,addr common.Address, pri *ecdsa.PrivateKey) error {
	prikey := hex.EncodeToString(ee.FromECDSA(pri))
	content := ip +" "+addr.Hex()+" "+prikey+" \n"

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
func (ee *EcdsaEncrypto)PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := ee.FromECDSAPub(&p)
	return common.BytesToAddress(ee.Keccak256(pubBytes[1:])[12:])
}
func (ee *EcdsaEncrypto)Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}
func (ee *EcdsaEncrypto)FromECDSAPub(pub *ecdsa.PublicKey) []byte {
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