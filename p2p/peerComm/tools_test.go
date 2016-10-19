// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
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
