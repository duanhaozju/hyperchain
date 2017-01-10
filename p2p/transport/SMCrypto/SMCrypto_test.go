package SMCrypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var smc *SMCrypto

func init(){
	smc = NewSMCrypto()
}

func TestSMCrypto_Connect(t *testing.T) {
	ret := smc.Connect()
	assert.Equal(t,ret,0)
}