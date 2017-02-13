//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package transport

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDesEncrypt(t *testing.T) {
	key := []byte("TnrEP|N.*lAgy<Q&@lBPd@J/")
	result,err := TripleDesEncrypt([]byte("hyperchain"), key)
	assert.Nil(t,err)
	assert.NotEmpty(t,result)
}

func TestDesDecrypt(t *testing.T) {
	key := []byte("TnrEP|N.*lAgy<Q&@lBPd@J/")
	ori := []byte{15,59,116,170,98,81,15,91,105,37,126,229,180,145,149,108}
	result,err := TripleDesDecrypt(ori, key)
	assert.Nil(t,err)
	assert.NotEmpty(t,result)
	assert.Equal(t,result,[]byte("hyperchain"))
}

func BenchmarkDesEncrypt(b *testing.B) {
	key := []byte("TnrEP|N.*lAgy<Q&@lBPd@J/")
	for i := 0; i<b.N;i++{
		TripleDesEncrypt([]byte("hyperchain"), key)
	}
}

func BenchmarkDesDecrypt(b *testing.B) {
	key := []byte("TnrEP|N.*lAgy<Q&@lBPd@J/")
	ori := []byte{15,59,116,170,98,81,15,91,105,37,126,229,180,145,149,108}
	for i :=0;i<b.N;i++{
		_, err := TripleDesDecrypt(ori, key)
		if err != nil {
			panic(err)
		}
	}
}
