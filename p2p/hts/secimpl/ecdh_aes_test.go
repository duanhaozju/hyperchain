package secimpl

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/csprng"
	"github.com/stretchr/testify/assert"
	"testing"
)

var pri1 = []byte(`
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICuXsSngHRi7HI7HbccKj3TOvFYM9TkEEYU+p+bot9iaoAoGCCqGSM49
AwEHoUQDQgAEEsyzz7Yxyqgrl4xOw+LuISXpP275a4HALrcwo1svLIOIFRiMN2Uj
a3irZzkYIoa+LxWJrh36+luHkJkikD33hA==
-----END EC PRIVATE KEY------`)

var rawcert1 = []byte(`
-----BEGIN CERTIFICATE-----
MIIB9zCCAZ6gAwIBAgIBATAKBggqhkjOPQQDAjBMMRMwEQYDVQQKEwpIeXBlcmNo
YWluMRYwFAYDVQQDEw1oeXBlcmNoYWluLmNuMRAwDgYDVQQqEwdEZXZlbG9wMQsw
CQYDVQQGEwJaSDAgFw0xNzA3MTcxMDE3NTNaGA8yMTE3MDYyMzExMTc1M1owTDET
MBEGA1UEChMKSHlwZXJjaGFpbjEWMBQGA1UEAxMNaHlwZXJjaGFpbi5jbjEQMA4G
A1UEKhMHRGV2ZWxvcDELMAkGA1UEBhMCWkgwWTATBgcqhkjOPQIBBggqhkjOPQMB
BwNCAAQSzLPPtjHKqCuXjE7D4u4hJek/bvlrgcAutzCjWy8sg4gVGIw3ZSNreKtn
ORgihr4vFYmuHfr6W4eQmSKQPfeEo28wbTAOBgNVHQ8BAf8EBAMCAgQwJgYDVR0l
BB8wHQYIKwYBBQUHAwIGCCsGAQUFBwMBBgIqAwYDgQsBMAwGA1UdEwEB/wQCMAAw
DQYDVR0OBAYEBAECAwQwFgYDKgMEBA9leHRyYSBleHRlbnNpb24wCgYIKoZIzj0E
AwIDRwAwRAIgTn7iBBVX4EpdbDxskIq5GLv3rx0I7iwR9jQGOkKVnWQCIFgffFic
Q/MhMLnuy5VDAM1ikNykC4XIMOgfL1JslD75
-----END CERTIFICATE------`)

var pri2 = []byte(`
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIJpCElLP0OHbpwgFmazbseo68Zbppl6d8bzZhaAclqlQoAoGCCqGSM49
AwEHoUQDQgAEeLmAGjS2AzlKm9gNMFQKJbAvxDXc5zjuxuwOmjcvz9MyOaxKnIm6
T/bMgqsl9pVS2KyiuWQubuvmUnbHrbUaOQ==
-----END EC PRIVATE KEY------`)

var rawcert2 = []byte(`
-----BEGIN CERTIFICATE-----
MIIB+TCCAZ6gAwIBAgIBATAKBggqhkjOPQQDAjBMMRMwEQYDVQQKEwpIeXBlcmNo
YWluMRYwFAYDVQQDEw1oeXBlcmNoYWluLmNuMRAwDgYDVQQqEwdEZXZlbG9wMQsw
CQYDVQQGEwJaSDAgFw0xNzA3MTcxMDE3NDlaGA8yMTE3MDYyMzExMTc0OVowTDET
MBEGA1UEChMKSHlwZXJjaGFpbjEWMBQGA1UEAxMNaHlwZXJjaGFpbi5jbjEQMA4G
A1UEKhMHRGV2ZWxvcDELMAkGA1UEBhMCWkgwWTATBgcqhkjOPQIBBggqhkjOPQMB
BwNCAAR4uYAaNLYDOUqb2A0wVAolsC/ENdznOO7G7A6aNy/P0zI5rEqcibpP9syC
qyX2lVLYrKK5ZC5u6+ZSdsettRo5o28wbTAOBgNVHQ8BAf8EBAMCAgQwJgYDVR0l
BB8wHQYIKwYBBQUHAwIGCCsGAQUFBwMBBgIqAwYDgQsBMAwGA1UdEwEB/wQCMAAw
DQYDVR0OBAYEBAECAwQwFgYDKgMEBA9leHRyYSBleHRlbnNpb24wCgYIKoZIzj0E
AwIDSQAwRgIhAIjutUGJCcWUM3BwE+wIn1zcwmasLoO7YAeaxQD6nKA9AiEA/gnX
3Znq9iqn3maDQCVVinfblLPQb9/lte3gOmYKA9s=
-----END CERTIFICATE------`)

func TestECDHWithAES_GenerateShareKey(t *testing.T) {
	ecaes := NewECDHWithAES()
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

func TestECDHWithAES_Encrypt(t *testing.T) {
	ecaes := NewECDHWithAES()
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
