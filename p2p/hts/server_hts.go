package hts

import (
	"crypto"
	"fmt"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/hts/secimpl"
	"github.com/hyperchain/hyperchain/p2p/peerevent"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
)

type ServerHTS struct {
	security       secimpl.Security
	priKey         []byte
	priKey_s       crypto.PrivateKey
	sessionKeyPool cmap.ConcurrentMap // key -> peer hash, value -> session key
	ev             *event.TypeMux
	CG             *CertGroup
}

// NewServerHTS creates and returns a new ServerHTS instance.
func NewServerHTS(sec secimpl.Security, cg *CertGroup, ev *event.TypeMux) (*ServerHTS, error) {
	sh := &ServerHTS{
		sessionKeyPool: cmap.New(),
		security:       sec,
		priKey:         cg.eCERTPriv,
		priKey_s:       cg.eCERTPriv_S,
		CG:             cg,
		ev:             ev,
	}
	return sh, nil
}

// KeyExchange generates a shared session key.
func (sh *ServerHTS) KeyExchange(idenHash string, rand []byte, rawcert []byte) error {
	sk, err := sh.security.GenerateShareKey(sh.priKey, rand, rawcert)
	if err != nil {
		return err
	}
	sessionKey := NewSessionKey(sk)
	sh.sessionKeyPool.Set(idenHash, sessionKey)
	return nil
}

// Encrypt will use the shared session key to encrypt message.
func (sh *ServerHTS) Encrypt(identify string, msg []byte) []byte {
	defer func() {
		rec := recover()
		if rec != nil {
			fmt.Println("Encrypt failed, fatal error:", rec)
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
			fmt.Println("ENCRYPT err ", err.Error())
			return nil
		}
		return encMsg
	}
	return nil
}

// Decrypt will use the shared session key to decrypt message.
func (sh *ServerHTS) Decrypt(identify string, msg []byte) (b []byte, err error) {

	defer func() {
		rec := recover()
		if rec != nil {
			go sh.ev.Post(peerevent.S_UPDATE_SESSION_KEY{identify})
			err = errors.New(fmt.Sprintf("recovery from decrypt failed, %v", rec))
		}
	}()
	if sessionKey, ok := sh.sessionKeyPool.Get(identify); ok {
		sessionKey := sessionKey.(*SessionKey)
		sharedKey := sessionKey.GetKey()
		if sharedKey == nil {
			return nil, errors.New(fmt.Sprintf("this session key is expired, id: %s ", identify))
		}

		decMsg, err := sh.security.Decrypt(sharedKey, msg)
		if err != nil {
			fmt.Printf("Dec err: %s ", err.Error())
			go sh.ev.Post(peerevent.S_UPDATE_SESSION_KEY{identify})
			return nil, errors.New(fmt.Sprintf("descrypt keyt failed, reason: %s ", err.Error()))
		}
		return decMsg, nil
	}
	go sh.ev.Post(peerevent.S_UPDATE_SESSION_KEY{identify})
	return nil, errors.New(fmt.Sprintf("cannot find the session key of %s", identify))
}

// GetSK returns the shared session key that is negotiated with a peer.
func (sh *ServerHTS) GetSK(hash string) []byte {
	if sessionKey, ok := sh.sessionKeyPool.Get(hash); ok {
		sessionKey := sessionKey.(*SessionKey)
		sharedKey := sessionKey.GetKey()
		return sharedKey
	}
	return nil
}
