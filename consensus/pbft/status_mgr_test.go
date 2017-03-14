package pbft

import (
	"testing"
	"fmt"
	"sync/atomic"
)

func TestStatusMgr(t *testing.T) {
	var status PbftStatus
	status.activeState(&status.byzantine)
	status.inActiveState(&status.isNewNode)

	fmt.Println(atomic.LoadInt32(&status.byzantine))
	fmt.Println(status.negCurrentState(&status.isNewNode))
	fmt.Println(status.isNewNode)
	if status.getState(&status.byzantine){
		fmt.Println("yes")
	}
}
