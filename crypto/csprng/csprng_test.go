package csprng

import (
	"github.com/stretchr/testify/assert"
	"hyperchain/common"
	"testing"
)

func TestCSPRNG(t *testing.T) {
	r, e := CSPRNG(32)
	assert.Nil(t, e)
	t.Log(common.ToHex(r))
}
