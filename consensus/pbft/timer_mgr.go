//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"hyperchain/consensus/events"
	"hyperchain/common"
	"time"
)

/**
	this file provide a mechanism to manage different timers.
 */

//timer wrapper for events timer.
type timer struct {
	timerName string
	eventTimer events.Timer //event timer
	timeout time.Duration    //timer timeout
	isActive bool
} 

//timerManager manage common used timer.
type timerManager struct {
	timers       map[string]*timer
	timerFactory events.TimerFactory //timer factory
	eventManager events.Manager      //manage pbft event
	requestTimeout time.Duration
}

//newTimerMgr init a instance of timerManager.
func newTimerMgr(conf *common.Config, pbft *pbftImpl, mt int) *timerManager {
	tm := &timerManager{}
	switch mt {
	case PBFT:
		tm.timers = make(map[string]*timer)
		tm.eventManager = events.NewManagerImpl()
		tm.eventManager.SetReceiver(pbft)
		tm.timerFactory = events.NewTimerFactoryImpl(tm.eventManager)
	case BATCH: //TODO: add timer
	default:
		logger.Errorf("Invalid timer manager type: %s", mt)
	}
	tm.requestTimeout = conf.GetDuration(PBFT_REQUEST_TIMEOUT)
	return tm
}

//newTimer new a pbft timer by timer name.
func (tm *timerManager) newTimer(tname string, timerOut time.Duration) {
	if tm.containsTimer(tname) {
		logger.Warningf("Timer %s has been created!!!", tname)
		return
	}
	tm.timers[tname] = &timer{
		timerName:tname,
		eventTimer:tm.timerFactory.CreateTimer(),
		timeout:timerOut,
		isActive:false,
	}
}

//startTimer start timer by timerName.
func (tm *timerManager) startTimer(timerName string, event events.Event) {
	tm.resetTimer(timerName, event)
}

//stopTimer stop timer by timerName.
func (tm *timerManager) stopTimer(timerName string) {
	if !tm.containsTimer(timerName) {
		logger.Errorf("Stop timer failed!, timer %s not created yet!", timerName)
	}
	tm.timers[timerName].eventTimer.Stop()
	tm.timers[timerName].isActive = false

}

func (tm *timerManager) softResetTimer(timerName string, event events.Event)  {
	if !tm.containsTimer(timerName) {
		logger.Errorf("SoftReset timer failed!, timer %s not created yet!", timerName)
	}
	timer := tm.timers[timerName]
	timer.eventTimer.SoftReset(timer.timeout, event)
	timer.isActive = true
}

//resetTimer reset timer by timerName.
func (tm *timerManager) resetTimer(timerName string, event events.Event) {
	if !tm.containsTimer(timerName) {
		logger.Errorf("Reset timer failed!, timer %s not created yet!", timerName)
	}
	timer := tm.timers[timerName]
	timer.eventTimer.Reset(timer.timeout, event)
	timer.isActive = true
}

//resetTimerWithNewTT reset timer with new timeout
func (tm *timerManager) resetTimerWithNewTT(timerName string, timeout time.Duration, event events.Event)  {
	if !tm.containsTimer(timerName) {
		logger.Errorf("Reset timer failed!, timer %s not created yet!", timerName)
	}
	timer := tm.timers[timerName]
	//timer.timeout = timeout
	timer.eventTimer.Reset(timeout, event)
}

func (tm *timerManager) containsTimer(timerName string) bool {
	_, ok := tm.timers[timerName]
	return ok
}

//Start start the event manager
func (tm *timerManager) Start()  {
	tm.eventManager.Start()
}

// getTimeoutValue get event timer timeout
func (tm *timerManager) getTimeoutValue(timerName string) time.Duration {
	if !tm.containsTimer(timerName) {
		logger.Warningf("Get tiemout failed!, timer %s not created yet! no time out", timerName)
		return 0 * time.Second
	}
	return tm.timers[timerName].timeout
}

//makeRequestTimeoutLegal if requestTimeout is not legal, make it legal
func (tm *timerManager) makeRequestTimeoutLegal() {
	nullRequestTimeout := tm.getTimeoutValue(NULL_REQUEST_TIMER)
	if tm.requestTimeout >= nullRequestTimeout && nullRequestTimeout != 0{
		tm.timers[NULL_REQUEST_TIMER].timeout = 3 * tm.requestTimeout / 2
		logger.Warningf("Configured null request timeout must be greater " +
			"than request timeout, setting to %v", tm.getTimeoutValue(NULL_REQUEST_TIMER))
	}
}