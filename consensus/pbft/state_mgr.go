//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

//stateManager manages the state of pbft
//exp:
//	1.get status: status["activeView"]
//	2.checkStatesAnd(status["activeView"],status["inRecovery"])
//	3.activeState("inRecovery")

type status map[string]bool

const (
	BYZANTINE = "byzantine"
	SKIP_IN_PROGRESS = "skipInProgress"
	STATE_TRANSFERRING = "stateTransferring"
	VALID = "valid"
	NEW_VIEW_TIMER_ACTIVE = "timerActive"
	IN_RECOVERY = "inRecovery"
	IN_NEGO_VIEW = "inNegoView"
	IS_NEW_NODE = "isNewNode"
	IN_ADDING_NODE = "inAddingNode"
	IN_DELETING_NODE = "inDeletingNode"
	IN_VC_RESET = "inVcReset"			// track if replica itself in vcReset
	VC_HANDLED = "vcHandled"			// track if replica handled the vc after receive newview
	NEW_NODE_READY = "newNodeReady"
	UPDATE_HANDLED = "updateHandled"
)

//checkStatesAnd checks the result of several status and each other
func (status status) checkStatesAnd(stateName... bool) bool {
	var rs bool = true
	for _, s := range stateName{
		rs = rs && s
	}
	return rs
}

//checkStatesAnd checks the result of several status or each other
func (status status) checkStatesOr(stateName... bool) bool {
	var rs bool = false
	for _, s := range stateName{
		rs = rs || s
	}
	return rs
}

//activeState sets the states to true
func (status status) activeState(stateName... string) {
	for _, s := range stateName{
		status[s] = true
	}
}

//inActiveState sets the states to false
func (status status) inActiveState(stateName... string)  {
	for _, s := range stateName{
		status[s] = false
	}
}

func (status status) print() {
	logger.Debug("┌─────────PBFT STATUS────────┐")

	for k, v := range status{
		if v == true{
			logger.Debugf("│         %-16s   │",k)
		}
	}
	logger.Debug("└────────────────────────────┘")
}
