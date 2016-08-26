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
)

func TestGetIpLocalIpAddr(t *testing.T) {
	if assert.IsType(t, string(""), GetIpLocalIpAddr()) {

	} else {
		t.Fail()
	}
}

func TestGetConfig(t *testing.T) {
	configs := GetConfig("/home/chenquan/Workspace/IdeaProjects/hyperchain-go/src/hyperchain-alpha/p2p/peerconfig.json")
	log.Println(configs["port"])
	log.Println(configs["seednode"])

}
