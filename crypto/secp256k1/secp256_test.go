package secp256k1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSign(t *testing.T) {
	_, e := Sign([]byte("hahah"), []byte("haahahah"))
	assert.Nil(t, e)
}
