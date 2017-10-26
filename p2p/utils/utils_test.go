package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetProjectPath(t *testing.T) {
	t.Log(GetProjectPath())
	t.Log(GetLocalIP())
	t.Log(Sha3([]byte("hyperchain")))

	assert.Equal(t, "e080f59e1154d7a8e98a7006c93817beaa5159b9c3502e51a95b702abd2ef868", HashString("global"+"node1"))

	assert.Equal(t,
		"0x952ab6f8aad7f84a98fdfcabd2371f69d91d32b7fd577632f38d5caeb7f781d2",
		GetPeerHash("global", 2))

	assert.False(t, IPcheck("abcd"))
	assert.False(t, IPcheck("172.16.0"))
	assert.True(t, IPcheck("172.16.0.1"))
}
