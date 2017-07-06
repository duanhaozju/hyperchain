package hts

import (
	"crypto"
	"crypto/rand"
	"github.com/pkg/errors"
)

type ClientHTS struct {
	priKey crypto.PrivateKey
	pubKey crypto.PublicKey
	sessionKey SessionKey
	security Security
}

func NewCLientHTS()*ClientHTS{
	return &ClientHTS{}
}

func(ch *ClientHTS)GenPriKey() error{
	pri,pub,err := ch.security.GeneratePrivateKey(rand.Reader)
	if err !=nil{
		return err
	}
	ch.priKey = pri
	ch.pubKey = pub
	return nil
}

func(ch *ClientHTS)GenSharedKey(serverPubKey []byte) error{
	pubKey,err := ch.security.UnMarshal(serverPubKey)
	if err != nil{
		return err
	}
	sharedKey, err := ch.security.GenerateSharedKey(ch.priKey,pubKey)
	if err != nil{
		return err
	}
	ch.sessionKey = NewSessionKey(sharedKey)
	return nil
}

func(ch *ClientHTS)Encrypt(msg []byte)([]byte,error){
	sKey := ch.sessionKey.GetKey()
	if sKey == nil{
		return nil,errors.New("cannot get session Key,enc failed.")
	}
	return ch.security.Encrypt(sKey,msg)
}

func(ch *ClientHTS)Decrypt(msg []byte)([]byte,error){
	sKey := ch.sessionKey.GetKey()
	if sKey == nil{
		return nil,errors.New("cannot get session Key,enc failed.")
	}
	return ch.security.Encrypt(sKey,msg)
}
