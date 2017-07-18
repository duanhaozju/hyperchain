package hts

import (
	"crypto"
	"fmt"

	"github.com/orcaman/concurrent-map"
	"agile/utils/common"
)

type ServerHTS struct {
	security       Security
	priKey        []byte
	priKey_s crypto.PrivateKey
	sessionKeyPool cmap.ConcurrentMap
	CG *CertGroup
}

func NewServerHTS(sec Security,cg *CertGroup)(*ServerHTS,error) {
	sh := &ServerHTS{
		sessionKeyPool: cmap.New(),
		security:sec,
		priKey:cg.eCERTPriv,
		priKey_s:cg.eCERTPriv_S,
		CG:cg,
	}
	return sh,nil
}



func (sh *ServerHTS) KeyExchange(idenHash string,rand []byte, rawcert []byte) error {
	sk ,err := sh.security.GenerateShareKey(sh.priKey,rand,rawcert)
	if err != nil {
		return err
	}
	fmt.Printf("server key exchange: %s,%s\n",idenHash,common.ToHex(sk))
	sessionKey := NewSessionKey(sk)
	sh.sessionKeyPool.Set(idenHash, sessionKey)
	return nil
}

func (sh *ServerHTS) Encrypt(identify string, msg []byte) []byte {
	if sessionKey, ok := sh.sessionKeyPool.Get(identify); ok {
		sKey := sessionKey.(*SessionKey)
		sharedKey := sKey.GetKey()
		if sharedKey == nil {
			fmt.Printf("this session key is expired, id: %s \n", identify)
			return nil
		}
		encMsg, err := sh.security.Encrypt(sharedKey, msg)
		if err != nil {
			return nil
		}
		return encMsg
	}
	return nil
}

func (sh *ServerHTS) Decrypt(identify string, msg []byte) []byte {
	if sessionKey, ok := sh.sessionKeyPool.Get(identify); ok {
		sessionKey := sessionKey.(*SessionKey)
		sharedKey := sessionKey.GetKey()
		if sharedKey == nil {
			fmt.Printf("this session key is expired, id: %s \n", identify)
			return nil
		}
		decMsg, err := sh.security.Decrypt(sharedKey, msg)
		if err != nil {
			return nil
		}
		return decMsg
	}
	return nil
}
