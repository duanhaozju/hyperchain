package Server

import (
	"testing"
	"time"
	"fmt"
)

func TestNewChatServer(t *testing.T) {
	server := NewNode(8001)
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
