//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"time"
)

/**
	this file provide a mechanism to manage different timers.
 */

//titletimers manage timer with the same timer name
type titletimers struct {
	timerName string
	timeout   time.Duration
	counts    int
	isActive  []bool
}

//timerManager manage common used timer.
type timerManager struct {
	ttimers        map[string]*titletimers
	requestTimeout time.Duration
}

//newTimerMgr init a instance of timerManager.
func newTimerMgr(pbft *pbftImpl) *timerManager {
	tm := &timerManager{
		ttimers: make(map[string]*titletimers),
		requestTimeout: pbft.config.GetDuration(PBFT_REQUEST_TIMEOUT),
	}

	return tm
}

//newTimer new a pbft timer by timer name and append into correspond map
func (tm *timerManager) newTimer(tname string, d time.Duration) {
	tm.ttimers[tname] = &titletimers{
		timerName: tname,
		timeout:   d,
		counts:    0,
	}
}

//startTimer init and start a timer by name
func (tm *timerManager) startTimer(tname string, afterfunc func()) {
	logger.Errorf("Starting a new timer---%s", tname)

	tm.ttimers[tname].isActive = append(tm.ttimers[tname].isActive, true)

	counts := len(tm.ttimers[tname].isActive)
	logger.Errorf("Now exsits %d timer---%s", counts, tname)

	send := func() {
		if tm.ttimers[tname].isActive[counts - 1] {
			afterfunc()
		}
	}
	time.AfterFunc(tm.ttimers[tname].timeout, send)
}

//startTimerWithNewTT init and start a timer by name with new timeout
func (tm *timerManager) startTimerWithNewTT(tname string, d time.Duration, afterfunc func()) {
	logger.Errorf("Starting a new timer---%s with new duration %d", tname, d)

	tm.ttimers[tname].isActive = append(tm.ttimers[tname].isActive, true)
	tm.ttimers[tname].counts ++

	counts := len(tm.ttimers[tname].isActive)
	logger.Errorf("Now exsits %d---%d timer---%s", counts, tm.ttimers[tname].counts, tname)

	send := func() {
		if tm.ttimers[tname].isActive[counts - 1] {
			afterfunc()
		}
	}
	time.AfterFunc(d, send)
}

//stopTimer stop all timers by the same timerName.
func (tm *timerManager) stopTimer(tname string) {
	if !tm.containsTimer(tname) {
		logger.Errorf("Stop timer failed!, timer %s not created yet!", tname)
	}
	logger.Errorf("Stoping timer---%s", tname)

	counts := len(tm.ttimers[tname].isActive)
	logger.Errorf("Now exsits %d timer---%s", counts, tname)

	for i := range tm.ttimers[tname].isActive {
		tm.ttimers[tname].isActive[i] = false
	}

	tm.ttimers[tname].counts = 0

}

func (tm *timerManager) containsTimer(timerName string) bool {
	_, ok := tm.ttimers[timerName]
	return ok
}

// getTimeoutValue get event timer timeout
func (tm *timerManager) getTimeoutValue(timerName string) time.Duration {
	if !tm.containsTimer(timerName) {
		logger.Warningf("Get tiemout failed!, timer %s not created yet! no time out", timerName)
		return 0 * time.Second
	}
	return tm.ttimers[timerName].timeout
}

func (tm *timerManager) setTimeoutValue(timerName string, d time.Duration) {
	if !tm.containsTimer(timerName) {
		logger.Warningf("Set tiemout failed!, timer %s not created yet! no time out", timerName)
		return
	}
	tm.ttimers[timerName].timeout = d
}

//makeRequestTimeoutLegal if requestTimeout is not legal, make it legal
func (tm *timerManager) makeRequestTimeoutLegal() {
	nullRequestTimeout := tm.getTimeoutValue(NULL_REQUEST_TIMER)
	requestTimeout := tm.getTimeoutValue(PBFT_REQUEST_TIMEOUT)

	if requestTimeout >= nullRequestTimeout && nullRequestTimeout != 0{
		tm.ttimers[NULL_REQUEST_TIMER].timeout = 3 * tm.ttimers[PBFT_REQUEST_TIMEOUT].timeout / 2
		logger.Warningf("Configured null request timeout must be greater " +
			"than request timeout, setting to %v", tm.getTimeoutValue(NULL_REQUEST_TIMER))
	}
}