package hts

import (
	"testing"
	"hyperchain/p2p/utils"
	"github.com/stretchr/testify/assert"
	"hyperchain/p2p/hts/secimpl"
)

func TestNewHTS(t *testing.T) {
	caconfigPath := utils.GetProjectPath() + "/p2p/test/cert/caconfig.yaml"
	_, err := NewHTS(nil, caconfigPath)
	assert.Nil(t, err, "err should be nil")
}

func TestHTS_GetAClientHTS(t *testing.T) {
	caconfigPath := utils.GetProjectPath() + "/p2p/test/cert/caconfig.yaml"
	hts, err := NewHTS(nil, caconfigPath)
	assert.Nil(t, err, "err should be nil")
	sign, err := hts.cg.ESign([]byte("hyperchain"))
	t.Log(len(sign))
	assert.NotNil(t, sign)
	assert.Nil(t, err)
}

func TestHTS_GetServerHTS(t *testing.T) {
	caconfigPath := utils.GetProjectPath() + "/p2p/test/cert/caconfig.yaml"
	hts, err := NewHTS(secimpl.NewECDHWithAES(), caconfigPath)
	assert.Nil(t, err, "err should be nil")
	chts, err := hts.GetAClientHTS()
	assert.Nil(t, err, "err should be nil")
	data := []byte("hyperchain")
	sign, err := chts.CG.ESign(data)
	assert.NotNil(t, sign)
	assert.Nil(t, err)
	shs, err := hts.GetServerHTS()
	assert.Nil(t, err)
	b, e := shs.CG.EVerify(sign, data)
	assert.Nil(t, e)
	assert.True(t, b)
}

func TestHTS_GetServerHTS2(t *testing.T) {
	caconfigPath := utils.GetProjectPath() + "/p2p/test/cert/caconfig.yaml"
	hts, err := NewHTS(secimpl.NewECDHWithAES(), caconfigPath)
	assert.Nil(t, err, "err should be nil")
	chts, err := hts.GetAClientHTS()
	assert.Nil(t, err, "err should be nil")
	data := []byte("hyperchain")
	sign, err := chts.CG.ESign(data)
	assert.NotNil(t, sign)
	assert.Nil(t, err)
	shs, err := hts.GetServerHTS()
	assert.Nil(t, err)
	b, e := shs.CG.EVerify(sign, data)
	assert.Nil(t, e)
	assert.True(t, b)

	pub, err := shs.CG.ParsePub(chts.CG.GetECert())
	assert.Nil(t, err)
	assert.NotNil(t, pub)
}
