package secp256k1

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	_,e := Sign([]byte("hahah"),[]byte("haahahah"))
	assert.Nil(t,e)
}
