package csprng

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCSPRNG(t *testing.T) {
	r, e := CSPRNG(32)
	assert.Nil(t, e)
	t.Log(common.ToHex(r))
}
