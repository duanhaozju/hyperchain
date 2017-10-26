package hts

import (
	"github.com/stretchr/testify/assert"
	"hyperchain/manager/event"
	"hyperchain/p2p/hts/secimpl"
	"testing"
)

func initialServerHTS(t *testing.T) *ServerHTS {
	hts, err := NewHTS("", secimpl.NewSecuritySelector(nsCnfigPath), nsCnfigPath)
	if err != nil {
		t.Fatalf("NewHTS error: %v", err)
	}

	ev := new(event.TypeMux)
	shts, err := hts.GetServerHTS(ev)
	assert.Nil(t, err, "should create a new hts server successfully")

	return shts
}

func TestNewServerHTS(t *testing.T) {
	initialServerHTS(t)
}

func TestServerHTS_KeyExchange(t *testing.T) {
	shts := initialServerHTS(t)

	// create a session key for hts server
	err := shts.KeyExchange("hyperchain", []byte("123"), shts.CG.eCERT)
	assert.Nil(t, err, "should create shared session key successfully")

	err = shts.KeyExchange("hyperchain", []byte("123"), nil)
	assert.NotNil(t, err, "should not create shared session key")
}

func TestServerHTS_EncryptAndDecrypt(t *testing.T) {
	shts := initialServerHTS(t)
	identify := "hyperchain"
	data := []byte("123456")

	// without session key
	encryptData := shts.Encrypt(identify, data)
	assert.Nil(t, encryptData, "should not encrypt data, because no session key")

	_, err := shts.Decrypt(identify, data)
	assert.NotNil(t, err, "should not decrypt data, because no session key")

	// create a session key for hts server
	err = shts.KeyExchange(identify, []byte("123"), shts.CG.eCERT)
	assert.Nil(t, err, "should create shared session key successfully")

	// with valid session key to encrypt and decrypt
	encryptData = shts.Encrypt(identify, data)
	decryptData, err := shts.Decrypt(identify, encryptData)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}
	assert.Equal(t, data, decryptData, "should decrypt encrypted data successfully")

	// with invalid session key
	if k, ok := shts.sessionKeyPool.Get(identify); !ok {
		t.Fatalf("cann't get session key")
	} else {
		key := k.(*SessionKey)
		key.SetOff()

		encryptData := shts.Encrypt(identify, data)
		assert.Nil(t, encryptData, "should not encrypt data, because session key is invalid")

		_, err := shts.Decrypt(identify, data)
		assert.NotNil(t, err, "should not decrypt data, because session key is invalid")

	}
}

func TestServerHTS_GetSK(t *testing.T) {
	shts := initialServerHTS(t)
	identify := "hyperchain"

	err := shts.KeyExchange(identify, []byte("123"), shts.CG.eCERT)
	assert.Nil(t, err, "should create shared session key successfully")

	sk, err := shts.security.GenerateShareKey(shts.priKey, []byte("123"), shts.CG.eCERT)
	if err != nil {
		t.Fatalf("GenerateShareKey error: %v", err)
	}
	shtsKey := shts.GetSK(identify)
	assert.Equal(t, sk, shtsKey, "the session key should be equal")

}
