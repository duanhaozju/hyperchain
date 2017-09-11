//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

// initTimers creates timers when start up
func (pbft *pbftImpl) initTimers() {
	pbft.timerMgr.newTimer(VC_RESEND_TIMER, pbft.config.GetDuration(PBFT_RESEND_VIEWCHANGE_TIMEOUT))
	pbft.timerMgr.newTimer(NULL_REQUEST_TIMER, pbft.config.GetDuration(PBFT_NULLREQUEST_TIMEOUT))
	pbft.timerMgr.newTimer(NEW_VIEW_TIMER, pbft.config.GetDuration(PBFT_VIEWCHANGE_TIMEOUT))
	pbft.timerMgr.newTimer(FIRST_REQUEST_TIMER, pbft.config.GetDuration(PBFT_FIRST_REQUEST_TIMEOUT))
	pbft.timerMgr.newTimer(NEGO_VIEW_RSP_TIMER, pbft.config.GetDuration(PBFT_NEGOVIEW_TIMEOUT))
	pbft.timerMgr.newTimer(RECOVERY_RESTART_TIMER, pbft.config.GetDuration(PBFT_RECOVERY_TIMEOUT))
	pbft.timerMgr.newTimer(BATCH_TIMER, pbft.config.GetDuration(PBFT_BATCH_TIMEOUT))
	pbft.timerMgr.newTimer(CLEAN_VIEW_CHANGE_TIMER, pbft.config.GetDuration(PBFT_BATCH_TIMEOUT))
	pbft.timerMgr.newTimer(VALIDATE_TIMER, pbft.config.GetDuration(PBFT_VALIDATE_TIMEOUT))
}

var eventCreators map[ConsensusMessage_Type]func() interface{}

// initMsgEventMap maps consensus_message to real consensus msg type which used to Unmarshal consensus_message's payload
// to actual consensus msg
func (pbft *pbftImpl) initMsgEventMap() {
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
func (pbft *pbftImpl) initStatus() {
	pbft.status.activeState(&pbft.status.inNegoView)
	pbft.status.activeState(&pbft.status.inRecovery)
}
