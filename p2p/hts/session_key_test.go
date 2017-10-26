package hts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSessionKey(t *testing.T) {
	key := NewSessionKey([]byte("123456"))

	// test SessionKey.IsValid()
	isValid := key.IsValid()
	assert.True(t, isValid, "the session key should be valid")

	// test SessionKey.GetKey()
	assert.Equal(t, []byte("123456"), key.GetKey())

	// test SessionKey.SetOff()
	key.SetOff()
	isValid = key.IsValid()
	assert.False(t, isValid, "the session key should be invalid")

	// test SessionKey.SetOn()
	key.SetOn()
	isValid = key.IsValid()
	assert.True(t, isValid, "the session key should be valid")

	// test SessionKey.Update()
	key.Update([]byte("7890"))
	assert.Equal(t, []byte("7890"), key.GetKey(), "the seesion key should be changed")
}
