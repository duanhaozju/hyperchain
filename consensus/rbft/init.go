//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

// initTimers creates timers when start up
func (rbft *rbftImpl) initTimers() {
	rbft.timerMgr.newTimer(VC_RESEND_TIMER, rbft.config.GetDuration(RBFT_RESEND_VIEWCHANGE_TIMEOUT))
	rbft.timerMgr.newTimer(NULL_REQUEST_TIMER, rbft.config.GetDuration(RBFT_NULLREQUEST_TIMEOUT))
	rbft.timerMgr.newTimer(NEW_VIEW_TIMER, rbft.config.GetDuration(RBFT_VIEWCHANGE_TIMEOUT))
	rbft.timerMgr.newTimer(FIRST_REQUEST_TIMER, rbft.config.GetDuration(RBFT_FIRST_REQUEST_TIMEOUT))
	rbft.timerMgr.newTimer(NEGO_VIEW_RSP_TIMER, rbft.config.GetDuration(RBFT_NEGOVIEW_TIMEOUT))
	rbft.timerMgr.newTimer(RECOVERY_RESTART_TIMER, rbft.config.GetDuration(RBFT_RECOVERY_TIMEOUT))
	rbft.timerMgr.newTimer(BATCH_TIMER, rbft.config.GetDuration(RBFT_BATCH_TIMEOUT))
	rbft.timerMgr.newTimer(REQUEST_TIMER, rbft.config.GetDuration(RBFT_REQUEST_TIMEOUT))
	rbft.timerMgr.newTimer(CLEAN_VIEW_CHANGE_TIMER, rbft.config.GetDuration(RBFT_CLEAN_VIEWCHANGE_TIMEOUT))
	rbft.timerMgr.newTimer(VALIDATE_TIMER, rbft.config.GetDuration(RBFT_VALIDATE_TIMEOUT))

	rbft.timerMgr.makeNullRequestTimeoutLegal()
	rbft.timerMgr.makeRequestTimeoutLegal()
	rbft.timerMgr.makeCleanVcTimeoutLegal()
}

var eventCreators map[ConsensusMessage_Type]func() interface{}

// initMsgEventMap maps consensus_message to real consensus msg type which used to Unmarshal consensus_message's payload
// to actual consensus msg
func (rbft *rbftImpl) initMsgEventMap() {
	eventCreators = make(map[ConsensusMessage_Type]func() interface{})

	eventCreators[ConsensusMessage_TRANSACTION] = func() interface{} { return &TransactionBatch{} }
	eventCreators[ConsensusMessage_PRE_PREPARE] = func() interface{} { return &PrePrepare{} }
	eventCreators[ConsensusMessage_PREPARE] = func() interface{} { return &Prepare{} }
	eventCreators[ConsensusMessage_COMMIT] = func() interface{} { return &Commit{} }
	eventCreators[ConsensusMessage_CHECKPOINT] = func() interface{} { return &Checkpoint{} }
	eventCreators[ConsensusMessage_VIEW_CHANGE] = func() interface{} { return &ViewChange{} }
	eventCreators[ConsensusMessage_NEW_VIEW] = func() interface{} { return &NewView{} }
	eventCreators[ConsensusMessage_FRTCH_REQUEST_BATCH] = func() interface{} { return &FetchRequestBatch{} }
	eventCreators[ConsensusMessage_RETURN_REQUEST_BATCH] = func() interface{} { return &ReturnRequestBatch{} }
	eventCreators[ConsensusMessage_NEGOTIATE_VIEW] = func() interface{} { return &NegotiateView{} }
	eventCreators[ConsensusMessage_NEGOTIATE_VIEW_RESPONSE] = func() interface{} { return &NegotiateViewResponse{} }
	eventCreators[ConsensusMessage_RECOVERY_INIT] = func() interface{} { return &RecoveryInit{} }
	eventCreators[ConsensusMessage_RECOVERY_RESPONSE] = func() interface{} { return &RecoveryResponse{} }
	eventCreators[ConsensusMessage_RECOVERY_FETCH_QPC] = func() interface{} { return &RecoveryFetchPQC{} }
	eventCreators[ConsensusMessage_RECOVERY_RETURN_QPC] = func() interface{} { return &RecoveryReturnPQC{} }
	eventCreators[ConsensusMessage_ADD_NODE] = func() interface{} { return &AddNode{} }
	eventCreators[ConsensusMessage_DEL_NODE] = func() interface{} { return &DelNode{} }
	eventCreators[ConsensusMessage_READY_FOR_N] = func() interface{} { return &ReadyForN{} }
	eventCreators[ConsensusMessage_UPDATE_N] = func() interface{} { return &UpdateN{} }
	eventCreators[ConsensusMessage_AGREE_UPDATE_N] = func() interface{} { return &AgreeUpdateN{} }
	eventCreators[ConsensusMessage_FINISH_VCRESET] = func() interface{} { return &FinishVcReset{} }
	eventCreators[ConsensusMessage_FINISH_UPDATE] = func() interface{} { return &FinishUpdate{} }
	eventCreators[ConsensusMessage_FETCH_MISSING_TRANSACTION] = func() interface{} { return &FetchMissingTransaction{} }
	eventCreators[ConsensusMessage_RETURN_MISSING_TRANSACTION] = func() interface{} { return &ReturnMissingTransaction{} }

}

// initStatus init basic status when starts up
func (rbft *rbftImpl) initStatus() {
	rbft.status.reset()
	rbft.on(inNegotiateView)
	rbft.on(inRecovery)

	rbft.setNormal()
	rbft.setNotFull()
}
