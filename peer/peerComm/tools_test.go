package peerComm

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetIpLocalIpAddr(t *testing.T) {
	if assert.IsType(t,string(""),GetIpLocalIpAddr()){

	}else{
		t.Fail()
	}
}
