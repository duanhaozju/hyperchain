package secimpl

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/csprng"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestECDHWithSm4_GenerateShareKey(t *testing.T) {
	ecaes := NewECDHWithSM4()
	r, err := csprng.CSPRNG(32)
	assert.Nil(t, err)

	sk1, e := ecaes.GenerateShareKey(pri1, r, rawcert2)
	assert.Nil(t, e)
	assert.NotNil(t, sk1)
	t.Log(sk1)

	sk2, e := ecaes.GenerateShareKey(pri2, r, rawcert1)
	assert.Nil(t, e)
	assert.NotNil(t, sk2)
	t.Log(sk2)
	assert.Equal(t, sk1, sk2)

}

func TestECDHWithSM4_Encrypt(t *testing.T) {
	ecaes := NewECDHWithSM4()
	r, err := csprng.CSPRNG(32)
	assert.Nil(t, err)

	sk1, e := ecaes.GenerateShareKey(pri1, r, rawcert2)
	assert.Nil(t, e)
	assert.NotNil(t, sk1)
	t.Log(common.ToHex(sk1))

	sk2, e := ecaes.GenerateShareKey(pri2, r, rawcert1)
	assert.Nil(t, e)
	assert.NotNil(t, sk2)
	t.Log(common.ToHex(sk2))
	assert.Equal(t, sk1, sk2)

	data := []byte("hyperchain")
	b1, e := ecaes.Encrypt(sk1, data)
	t.Log(e)
	assert.Nil(t, e)
	b2, e := ecaes.Decrypt(sk1, b1)
	assert.Equal(t, data, b2)
}
