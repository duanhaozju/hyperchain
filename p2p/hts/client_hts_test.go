package hts

import (
	"testing"
	"hyperchain/p2p/hts/secimpl"
	"github.com/stretchr/testify/assert"
	"hyperchain/crypto"
)

func initialClientHTS(t *testing.T) *ClientHTS{
	hts, err := NewHTS("", secimpl.NewSecuritySelector(nsCnfigPath), nsCnfigPath)
	if err != nil {
		t.Fatalf("NewHTS error: %v", err)
	}

	chts, err := hts.GetAClientHTS()
	if err != nil {
		t.Fatalf("GetAClientHTS error: %v", err)
	}
	return chts
}

func TestNewClientHTS(t *testing.T) {
	initialClientHTS(t)
}

func TestClientHTS_GenShareKey(t *testing.T) {
	chts := initialClientHTS(t)

	err := chts.GenShareKey([]byte("123"), chts.CG.eCERT)
	assert.Nil(t, err, "should create shared session key successfully")

	err = chts.GenShareKey([]byte("123"), nil)
	assert.NotNil(t, err, "should not create shared session key")
}

func TestClientHTS_EncryptAndDecrypt(t *testing.T) {
	chts := initialClientHTS(t)

	// no session key
	data := []byte("hyperchain")
	_, err := chts.Encrypt(data)
	assert.NotNil(t, err, "should encrypt failed, because no session key")

	_, err = chts.Decrypt(data)
	assert.NotNil(t, err, "should decrypt failed, because no session key")

	// with session key
	err = chts.GenShareKey([]byte("123"), chts.CG.eCERT)
	assert.Nil(t, err, "should create shared session key successfully")

	eData, err := chts.Encrypt(data)
	assert.Nil(t, err, "should encrypt data successfully")

	dData, err := chts.Decrypt(eData)
	assert.Nil(t, err, "err should be nil")
	assert.Equal(t, data, dData, "should decrypt successfully")

	// with invalid session key
	tmp_chts := initialClientHTS(t)
	if err != nil {
		t.Fatalf("GetAClientHTS error: %v", err)
	}

	err = tmp_chts.GenShareKey([]byte("123"), chts.CG.eCERT)
	assert.Nil(t, err, "should create shared session key successfully")
	tmp_chts.sessionKey.SetOff()
	_, err = tmp_chts.Encrypt(data)
	assert.NotNil(t, err, "should encrypt failed, because session key is invalid")

	_, err = tmp_chts.Decrypt(data)
	assert.NotNil(t, err, "should decrypt failed, because session key is invalid")
}

func TestClientHTS_VerifySign(t *testing.T) {

	chts := initialClientHTS(t)

	hash := crypto.Keccak256([]byte("sign the data"))
	sign, err := chts.CG.ESign(hash)
	if err != nil {
		t.Fatalf("ESign error: %v", err)
	}
	isSuccess, err := chts.VerifySign(sign, []byte("sign the data"), chts.CG.eCERT)
	if err != nil {
		t.Fatalf("VerifySign error: %v", err)
	}
	assert.True(t, isSuccess, "the client hts should verify the signature successfully")
}

func TestClientHTS_GetSK(t *testing.T) {

	chts := initialClientHTS(t)

	chts.GenShareKey([]byte("123"), chts.CG.eCERT)

	sk, err := chts.security.GenerateShareKey(chts.priKey, []byte("123"), chts.CG.eCERT)
	if err != nil {
		t.Fatalf("GenerateShareKey error: %v", err)
	}
	chtsKey := chts.GetSK()
	assert.Equal(t, sk, chtsKey, "the session key should be equal")
}