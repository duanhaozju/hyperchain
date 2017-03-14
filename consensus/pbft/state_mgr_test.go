package pbft
//
//import (
//	"testing"
//	"fmt"
//)
//
//func TestState(t *testing.T) {
//	var status status
//	status = make(status)
//	status.activeState(IN_RECOVERY)
//	status.inActiveState(IN_RECOVERY)
//	//status["inRecovery"] = true
//	fmt.Println(status[IN_RECOVERY])
//	fmt.Println(status[ACTIVE_VIEW])
//
//	fmt.Println(status.checkStatesAnd(status[IN_RECOVERY],status[ACTIVE_VIEW]))
//	fmt.Println(status.checkStatesOr(status[IN_RECOVERY],status[ACTIVE_VIEW]))
//
//}
//
