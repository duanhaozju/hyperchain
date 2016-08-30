// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:59
// last Modified Author: chenquan
// change log: add a header comment of this file
//

package Server

import (
	"testing"
	"time"
	"fmt"
	"hyperchain/event"

)

func TestNewChatServer(t *testing.T) {
	eventMux := new(event.TypeMux)
	server := NewNode(8001,true,eventMux)
	tickCount := 0
	for now:=range  time.Tick(3*time.Second){
		fmt.Println(now)
		tickCount +=1
		if tickCount >3{
			break
		}
	}

	server.StopServer()

}
