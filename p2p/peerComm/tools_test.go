// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerComm

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"fmt"
	"time"
)

func TestGetIpLocalIpAddr(t *testing.T) {
	if assert.IsType(t, string(""), GetLocalIp()) {

	} else {
		t.Fail()
	}
}

func TestGetConfig(t *testing.T) {
	filepath,_ := os.Getwd()
	filepath = path.Join(filepath,"../peerconfig.json")
	log.Info(filepath)
	configs := GetConfig(filepath)
	log.Info(configs["port1"])
	log.Info(configs["node1"])
}