// author: Lizhong kuang
// date: 16-8-29
// last modified: 16-8-29

package manager

import (
	"testing"

	"fmt"
	"hyperchain/event"
	"time"
)


func TestEncryption2(t *testing.T){
	manager := &ProtocolManager{
		eventMux:    new(event.TypeMux),
		quitSync:    make(chan struct{}),



	}
	manager.aLiveSub = manager.eventMux.Subscribe(event.AliveEvent{})

	go newEvent(manager)

	for obj := range manager.aLiveSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.AliveEvent:
			fmt.Print(ev.Payload)
		}
	}


}
func newEvent(manager *ProtocolManager)  {
	fmt.Println("1")
	for i := 0; i < 100; i += 1 {
		fmt.Println(i)

		//eventmux := new(event.TypeMux)
		manager.eventMux.Post(event.AliveEvent{true})
		time.Sleep(20000002222)

	}

}


func TestDecodeTx(t *testing.T){
	manager := &ProtocolManager{
		eventMux:    new(event.TypeMux),
		quitSync:    make(chan struct{}),



	}
	manager.aLiveSub = manager.eventMux.Subscribe(event.AliveEvent{})

	go newEvent(manager)

	for obj := range manager.aLiveSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.AliveEvent:
			fmt.Print(ev.Payload)
		}
	}


}
