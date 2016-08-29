// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerComm

import (
	"log"
	"testing"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
)

func TestGetIpLocalIpAddr(t *testing.T) {
	if assert.IsType(t, string(""), GetIpLocalIpAddr()) {

	} else {
		t.Fail()
	}
}

func TestGetConfig(t *testing.T) {
	filepath,_ := os.Getwd()
	filepath = path.Join(filepath,"../peerconfig.json")
	log.Println(filepath)
	configs := GetConfig(filepath)
	log.Println(configs["port1"])
	log.Println(configs["node1"])
}
