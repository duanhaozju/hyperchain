//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package peerComm

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigUtil(t *testing.T) {
	cu := NewConfigUtil("../../config/local_peerconfig.json")
	assert.Exactly(t,4,cu.GetMaxPeerNumber())
	assert.Exactly(t,"0.0.0.0",cu.GetIP(1))
	assert.Exactly(t,int64(8001),cu.GetPort(1))
}
