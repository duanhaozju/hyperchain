package csprng

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"agile/utils/common"
)

func TestCSPRNG(t *testing.T) {
	r,e := CSPRNG(32)
	assert.Nil(t,e)
	t.Log(common.ToHex(r))
}
