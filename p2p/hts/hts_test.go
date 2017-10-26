package hts

import (
	"github.com/stretchr/testify/assert"
	"hyperchain/manager/event"
	"hyperchain/p2p/hts/secimpl"
	"hyperchain/p2p/utils"
	"testing"
)

var (
	nsCnfigPath = utils.GetProjectPath() + "/p2p/test/namespace2.toml"
)

func TestNewHTS(t *testing.T) {
	_, err := NewHTS("", nil, nsCnfigPath)
	assert.Nil(t, err, "err should be nil")
}

func TestHTS_GetAClientHTS(t *testing.T) {
	hts, err := NewHTS("", nil, nsCnfigPath)
	assert.Nil(t, err, "err should be nil")

	chts, err := hts.GetAClientHTS()
	sign, err := chts.CG.ESign([]byte("hyperchain"))
	t.Log(len(sign))
	assert.NotNil(t, sign)
	assert.Nil(t, err)
}

func TestHTS_GetServerHTS(t *testing.T) {
	hts, err := NewHTS("", secimpl.NewSecuritySelector(nsCnfigPath), nsCnfigPath)
	if err != nil {
		t.Fatalf("NewHTS error: %v", err)
	}
	assert.Nil(t, err, "err should be nil")

	chts, err := hts.GetAClientHTS()
	assert.Nil(t, err, "err should be nil")
	data := []byte("hyperchain")
	sign, err := chts.CG.ESign(data)
	assert.NotNil(t, sign)
	assert.Nil(t, err)

	ev := new(event.TypeMux)
	shts, err := hts.GetServerHTS(ev)
	assert.Nil(t, err)
	b, e := shts.CG.EVerify(shts.CG.eCERT, sign, data)
	assert.Nil(t, e)
	assert.True(t, b)
}
