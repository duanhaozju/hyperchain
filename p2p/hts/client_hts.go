package hts

import (
	"crypto"
	"github.com/pkg/errors"
	"fmt"
)

type ClientHTS struct {
	priKey_s     crypto.PrivateKey
	priKey    []byte
	pubKey_s     crypto.PublicKey
	pubKey  []byte
	sessionKey *SessionKey
	security   Security
	CG *CertGroup
}

func NewClientHTS(sec Security,cg *CertGroup) (*ClientHTS,error) {
	chts := &ClientHTS{
		security:sec,
		CG:cg,
	}
	chts.priKey = cg.eCERTPriv
	chts.priKey_s = cg.eCERTPriv_S
	return chts,nil
}

func(ch *ClientHTS) VerifySign(sign ,data, rawcert []byte) (bool,error){
	return ch.security.VerifySign(sign,data,rawcert)
}

func(ch *ClientHTS)GenShareKey(rand,rawcert []byte) error{

	sk,err := ch.security.GenerateShareKey(ch.priKey,rand,rawcert)
	if err != nil{
		return err
	}
	fmt.Println("client key exange",sk)
	ch.sessionKey = NewSessionKey(sk)
	return nil
}

func (ch *ClientHTS) Encrypt(msg []byte) ([]byte, error) {
	sKey := ch.sessionKey.GetKey()
	if sKey == nil {
		return nil, errors.New("cannot get session Key,enc failed.")
	}
	return ch.security.Encrypt(sKey, msg)
}

func (ch *ClientHTS) Decrypt(msg []byte) ([]byte, error) {
	sKey := ch.sessionKey.GetKey()
	if sKey == nil {
		return nil, errors.New("cannot get session Key,enc failed.")
	}
	return ch.security.Encrypt(sKey, msg)
}
