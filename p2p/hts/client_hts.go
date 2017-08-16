package hts

import (
	"crypto"
	"github.com/pkg/errors"
	"sync"
)

type ClientHTS struct {
	priKey_s     crypto.PrivateKey
	priKey    []byte
	pubKey_s     crypto.PublicKey
	pubKey  []byte
	sessionKey *SessionKey
	security   Security
	CG *CertGroup
	rwlock *sync.RWMutex
}

func NewClientHTS(sec Security,cg *CertGroup) (*ClientHTS,error) {
	chts := &ClientHTS{
		security:sec,
		CG:cg,
		rwlock:new(sync.RWMutex),
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
	ch.rwlock.Lock()
	ch.sessionKey = NewSessionKey(sk)
	ch.rwlock.Unlock()
	return nil
}

func (ch *ClientHTS) Encrypt(msg []byte) ([]byte, error) {
	ch.rwlock.RLock()
	defer ch.rwlock.RUnlock()
	sKey := ch.sessionKey.GetKey()
	if sKey == nil {
		return nil, errors.New("cannot get session Key,enc failed.")
	}
	b,e := ch.security.Encrypt(sKey, msg)
	if e != nil{
		ch.sessionKey.SetOff()
	}
	return b,e
}

func (ch *ClientHTS) Decrypt(msg []byte) ([]byte, error) {
	ch.rwlock.RLock()
	defer ch.rwlock.RUnlock()
	sKey := ch.sessionKey.GetKey()
	if sKey == nil {
		return nil, errors.New("cannot get session Key,enc failed.")
	}
	b,e :=  ch.security.Decrypt(sKey, msg)
	if e != nil{
		ch.sessionKey.SetOff()
	}
	return b,e
}

func(ch *ClientHTS)GetSK()[]byte{
	ch.rwlock.RLock()
	defer ch.rwlock.RUnlock()
	return ch.sessionKey.GetKey()
}
