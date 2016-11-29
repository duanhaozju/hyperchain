//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package peerComm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetIpLocalIpAddr(t *testing.T) {
	if assert.IsType(t, string(""), GetLocalIp()) {
	} else {
		t.Fail()
	}
}
