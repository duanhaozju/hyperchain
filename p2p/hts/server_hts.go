package hts

import (
	"github.com/orcaman/concurrent-map"
	"crypto"
	"crypto/rand"
	"github.com/pkg/errors"
	"fmt"
)

type ServerHTS struct {
	security Security
	priKey crypto.PrivateKey
	pubKey crypto.PublicKey
	sessionKeyPool cmap.ConcurrentMap
}

func NewServerHTS()*ServerHTS{
	sh := &ServerHTS{
		sessionKeyPool:cmap.New(),
	}
	sh.genPriKey()
	return sh
}


// this is the only way to change the serverHTS private key and public key
func(sh *ServerHTS)genPriKey()error{
	pri,pub,err := sh.security.GeneratePrivateKey(rand.Reader)
	if err !=nil{
		return errors.New(fmt.Sprintf("generate private key failed, reason %s",err.Error()))
	}
	sh.priKey = pri
	sh.pubKey = pub
	return nil
}
//GetPubKey get the public key (bytes) to key exchange
func(sh *ServerHTS)GetPubKey()[]byte{
	return sh.security.Marshal(sh.pubKey)
}

func(sh *ServerHTS)KeyExchange(identify string,remotePubKey []byte)error{
	remotePub,err := sh.security.UnMarshal(remotePubKey)
	if err != nil{
		return err
	}
	sharedKey,err := sh.security.GenerateSharedKey(sh.priKey,remotePub)
	if err != nil{
		return err
	}
	sessionKey := NewSessionKey(sharedKey)
	sh.sessionKeyPool.Set(identify,sessionKey)
	return nil
}

func(sh *ServerHTS)Encrypt(identify string,msg []byte)[]byte{
	if sessionKey,ok := sh.sessionKeyPool.Get(identify);ok{
		sKey := sessionKey.(*SessionKey)
		sharedKey := sKey.GetKey()
		if sharedKey == nil{
			fmt.Printf("this session key is expired, id: %s \n" ,identify)
			return nil
		}
		encMsg,err := sh.security.Encrypt(sharedKey,msg)
		if err != nil{
			return nil
		}
		return encMsg
	}
	return nil
}

func(sh *ServerHTS)Decrypt(identify string,msg []byte)[]byte{
	if sessionKey,ok := sh.sessionKeyPool.Get(identify);ok{
		sessionKey := sessionKey.(*SessionKey)
		sharedKey := sessionKey.GetKey()
		if sharedKey == nil{
			fmt.Printf("this session key is expired, id: %s \n" ,identify)
			return nil
		}
		decMsg,err := sh.security.Decrypt(sharedKey,msg)
		if err != nil{
			return nil
		}
		return decMsg
	}
	return nil
}
