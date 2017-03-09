//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

//initTimers creates timers when start up
func (pbft *pbftImpl) initTimers() {
	pbft.pbftTimerMgr.newTimer(VC_RESEND_TIMER, pbft.config.GetDuration(PBFT_RESEND_VIEWCHANGE_TIMEOUT))
	pbft.pbftTimerMgr.newTimer(NULL_REQUEST_TIMER, pbft.config.GetDuration(PBFT_NULLREQUEST_TIMEOUT))
	pbft.pbftTimerMgr.newTimer(NEW_VIEW_TIMER, pbft.config.GetDuration(PBFT_VIEWCHANGE_TIMEOUT))
	pbft.pbftTimerMgr.newTimer(FIRST_REQUEST_TIMER, pbft.config.GetDuration(PBFT_FIRST_REQUEST_TIMEOUT))
	pbft.pbftTimerMgr.newTimer(NEGO_VIEW_RSP_TIMER, pbft.config.GetDuration(PBFT_NEGOVIEW_TIMEOUT))
	pbft.pbftTimerMgr.newTimer(RECOVERY_RESTART_TIMER, pbft.config.GetDuration(PBFT_RECOVERY_TIMEOUT))
	//pbft.pbftTimerMgr.newTimer(ADD_NODE_TIMER, pbft.config.GetDuration(PBFT_ADD_NODE_TIMEOUT))
	//pbft.pbftTimerMgr.newTimer(DEL_NODE_TIMER, pbft.config.GetDuration(PBFT_DEL_NODE_TIMEOUT))
	//pbft.pbftTimerMgr.newTimer(UPDATE_TIMER, pbft.config.GetDuration(PBFT_UPDATE_TIMEOUT))
}

//initMsgEventMap construct consensus_message to real type map
func (pbft *pbftImpl) initMsgEventMap()  {
	eventCreators = make(map[ConsensusMessage_Type] func() interface{})

	eventCreators[ConsensusMessage_TRANSACTION] = func() interface{} {return &TransactionBatch{}}
	eventCreators[ConsensusMessage_PRE_PREPARE] = func() interface{} {return &PrePrepare{}}
	eventCreators[ConsensusMessage_PREPARE] = func() interface{} {return &Prepare{}}
	eventCreators[ConsensusMessage_COMMIT] = func() interface{} {return &Commit{}}

	eventCreators[ConsensusMessage_CHECKPOINT] = func() interface{} {return &Checkpoint{}}

	eventCreators[ConsensusMessage_VIEW_CHANGE] =  func() interface{} {return &ViewChange{}}
	eventCreators[ConsensusMessage_NEW_VIEW] =  func() interface{} {return &NewView{}}

	eventCreators[ConsensusMessage_FRTCH_REQUEST_BATCH] =  func() interface{} {return &FetchRequestBatch{}}
	eventCreators[ConsensusMessage_RETURN_REQUEST_BATCH] = func() interface{} {return &ReturnRequestBatch{}}

	eventCreators[ConsensusMessage_NEGOTIATE_VIEW] =  func() interface{} {return &NegotiateView{}}
	eventCreators[ConsensusMessage_NEGOTIATE_VIEW_RESPONSE] =  func() interface{} {return &NegotiateViewResponse{}}

	eventCreators[ConsensusMessage_RECOVERY_INIT] =  func() interface{} {return &RecoveryInit{}}
	eventCreators[ConsensusMessage_RECOVERY_RESPONSE] =  func() interface{} {return &RecoveryResponse{}}

	eventCreators[ConsensusMessage_RECOVERY_FETCH_QPC] =  func() interface{} {return &RecoveryFetchPQC{}}
	eventCreators[ConsensusMessage_RECOVERY_RETURN_QPC] =  func() interface{} {return &RecoveryReturnPQC{}}

	eventCreators[ConsensusMessage_ADD_NODE] =  func() interface{} {return &AddNode{}}
	eventCreators[ConsensusMessage_DEL_NODE] =  func() interface{} {return &DelNode{}}

	eventCreators[ConsensusMessage_READY_FOR_N] =  func() interface{} {return &ReadyForN{}}
	eventCreators[ConsensusMessage_UPDATE_N] =  func() interface{} {return &UpdateN{}}
	eventCreators[ConsensusMessage_AGREE_UPDATE_N] =  func() interface{} {return &AgreeUpdateN{}}
	eventCreators[ConsensusMessage_FINISH_VCRESET] =  func() interface{} {return &FinishVcReset{}}
}

//initStatus configs basic status when starts up
func (pbft *pbftImpl) initStatus(){
	pbft.status = make(status)
	pbft.status[BYZANTINE] = false
	pbft.status[IS_NEW_NODE] = false
	pbft.status[IN_ADDING_NODE] = false
	pbft.status[IN_DELETING_NODE] = false
	pbft.status[IN_VC_RESET] = false
	pbft.status[VC_HANDLED] = false
	pbft.status[NEW_NODE_READY] = false
	pbft.status[UPDATE_HANDLED] = false
	pbft.status[IN_NEGO_VIEW] = true
	pbft.status[IN_RECOVERY] = true
}
