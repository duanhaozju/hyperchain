package rbft

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func TestStatusMgr(t *testing.T) {
	var status RbftStatus
	status.activeState(&status.byzantine)
	status.inActiveState(&status.isNewNode)

	fmt.Println(atomic.LoadInt32(&status.byzantine))
	fmt.Println(status.isNewNode)
	if status.getState(&status.byzantine) {
		fmt.Println("yes")
	}
}
