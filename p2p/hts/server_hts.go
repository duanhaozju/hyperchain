package hts

import (
	"crypto"
	"fmt"
	"github.com/orcaman/concurrent-map"
	"hyperchain/manager/event"
	"hyperchain/p2p/peerevent"
)

type ServerHTS struct {
	security       Security
	priKey        []byte
	priKey_s crypto.PrivateKey
	sessionKeyPool cmap.ConcurrentMap
	ev *event.TypeMux
	CG *CertGroup
}

func NewServerHTS(sec Security,cg *CertGroup,ev *event.TypeMux)(*ServerHTS,error) {
	sh := &ServerHTS{
		sessionKeyPool: cmap.New(),
		security:sec,
		priKey:cg.eCERTPriv,
		priKey_s:cg.eCERTPriv_S,
		CG:cg,
		ev:ev,
	}
	return sh,nil
}



func (sh *ServerHTS) KeyExchange(idenHash string,rand []byte, rawcert []byte) error {
	sk ,err := sh.security.GenerateShareKey(sh.priKey,rand,rawcert)
	if err != nil {
		return err
	}
	sessionKey := NewSessionKey(sk)
	sh.sessionKeyPool.Set(idenHash, sessionKey)
	return nil
}

func (sh *ServerHTS) Encrypt(identify string, msg []byte) []byte {
	defer func(){
		rec := recover()
		if rec  != nil{
			fmt.Println("Encrypt failed, fatal error:",rec)
		}
	}()
	if sessionKey, ok := sh.sessionKeyPool.Get(identify); ok {
		sKey := sessionKey.(*SessionKey)
		sharedKey := sKey.GetKey()
		if sharedKey == nil {
			fmt.Printf("this session key is expired, id: %s ", identify)
			return nil
		}
		encMsg, err := sh.security.Encrypt(sharedKey, msg)
		if err != nil {
			fmt.Println("ENCRYPT err ",err.Error())
			return nil
		}
		return encMsg
	}
	return nil
}

func (sh *ServerHTS) Decrypt(identify string, msg []byte) []byte {

	defer func(){
		rec := recover()
		if rec  != nil{
			go sh.ev.Post(peerevent.S_UPDATE_SESSION_KEY{identify})
		}
	}()
	if sessionKey, ok := sh.sessionKeyPool.Get(identify); ok {
		sessionKey := sessionKey.(*SessionKey)
		sharedKey := sessionKey.GetKey()
		if sharedKey == nil {
			//TODO expired.
			fmt.Printf("this session key is expired, id: %s ", identify)
			return nil
		}
		decMsg, err := sh.security.Decrypt(sharedKey, msg)
		if err != nil {
			fmt.Printf("Dec err: %s ", err.Error())
			go sh.ev.Post(peerevent.S_UPDATE_SESSION_KEY{identify})
			return nil
		}
		return decMsg
	}
	fmt.Println("key is nil,search kkey for",identify)
	go sh.ev.Post(peerevent.S_UPDATE_SESSION_KEY{identify})
	return nil
}

func (sh *ServerHTS)GetSK(hash string) []byte {
	if sessionKey, ok := sh.sessionKeyPool.Get(hash); ok {
		sessionKey := sessionKey.(*SessionKey)
		sharedKey := sessionKey.GetKey()
		return sharedKey
	}
	return nil
}
