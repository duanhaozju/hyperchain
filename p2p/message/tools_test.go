//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package message

import (
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetHash(t *testing.T) {
	// those test hash was generate from:
	// https://emn178.github.io/online-tools/keccak_256.html
	// 2017-02-10
	// for this test case change the GetHash real hash method as ByteHash

	//needhash 4b14b31acca9720d7ebcc6b35bfc18fb161cd1bae4e52c256dd57be880d4be7c
	//helloworld fa26db7ca85ead399216e7c6316bc50ed24393c3122b582735e7f3b0f91b93f0
	assert.Equal(t, GetHash("needhash"), "0x4b14b31acca9720d7ebcc6b35bfc18fb161cd1bae4e52c256dd57be880d4be7c")
	assert.Equal(t, GetHash("helloworld"), "0xfa26db7ca85ead399216e7c6316bc50ed24393c3122b582735e7f3b0f91b93f0")
}
