//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"time"

	"hyperchain/manager/event"

	"github.com/op/go-logging"
)

/**
this file provide a mechanism to manage different timers.
*/

// titletimer manages timer with the same timer name, which, we allow different timer with the same timer name, such as:
// we allow several request timers at the same time, each timer started after received a new request batch
type titletimer struct {
	timerName string        // the unique timer name
	timeout   time.Duration // default timeout of this timer
	isActive  []bool        // track all the timers with this timerName if it is active now
}

// timerManager manages common used timers.
type timerManager struct {
	ttimers map[string]*titletimer
	logger  *logging.Logger
}

// newTimerMgr news an instance of timerManager.
func newTimerMgr(logger *logging.Logger) *timerManager {
	tm := &timerManager{
		ttimers: make(map[string]*titletimer),
		logger:  logger,
	}

	return tm
}

// newTimer news a titletimer with the given name and default timeout, then add this timer to timerManager
func (tm *timerManager) newTimer(name string, d time.Duration) {
	tm.ttimers[name] = &titletimer{
		timerName: name,
		timeout:   d,
	}
}

// Stop stops all timers managed by timerManager
func (tm *timerManager) Stop() {
	for timerName := range tm.ttimers {
		tm.stopTimer(timerName)
	}
}

// startTimer starts the timer with the given name and default timeout, then sets the event which will be triggered
// after this timeout
func (tm *timerManager) startTimer(name string, event *LocalEvent, queue *event.TypeMux) int {
	tm.stopTimer(name)

	tm.ttimers[name].isActive = append(tm.ttimers[name].isActive, true)
	counts := len(tm.ttimers[name].isActive)

	send := func() {
		if tm.ttimers[name].isActive[counts-1] {
			queue.Post(event)
		}
	}
	time.AfterFunc(tm.ttimers[name].timeout, send)
	return counts - 1
}

// startTimerWithNewTT starts the timer with the given name and timeout, then sets the event which will be triggered
// after this timeout
func (tm *timerManager) startTimerWithNewTT(name string, d time.Duration, event *LocalEvent, queue *event.TypeMux) int {
	if queue == nil {
		tm.logger.Error("find a nil TypeMux when start %s with duration: %v", name, d)
		return 0
	}
	tm.stopTimer(name)

	tm.ttimers[name].isActive = append(tm.ttimers[name].isActive, true)
	counts := len(tm.ttimers[name].isActive)

	send := func() {
		if tm.ttimers[name].isActive[counts-1] {
			queue.Post(event)
		}
	}
	time.AfterFunc(d, send)
	return counts - 1
}

// stopTimer stops all timers with the same timerName.
func (tm *timerManager) stopTimer(name string) {
	if !tm.containsTimer(name) {
		tm.logger.Errorf("Stop timer failed!, timer %s not created yet!", name)
		return
	}

	for i := range tm.ttimers[name].isActive {
		tm.ttimers[name].isActive[i] = false
	}
}

// stopOneTimer stops one timer by the timerName and index.
func (tm *timerManager) stopOneTimer(tname string, num int) {
	if !tm.containsTimer(tname) {
		tm.logger.Errorf("Stop timer failed!, timer %s not created yet!", tname)
		return
	}

	counts := len(tm.ttimers[tname].isActive)
	if num >= counts {
		tm.logger.Errorf("Stop timer failed!, timer %s index out of range!", tname)
		return
	}

	tm.ttimers[tname].isActive[num] = false
}

// containsTimer returns true if there exists a timer named timerName
func (tm *timerManager) containsTimer(timerName string) bool {
	_, ok := tm.ttimers[timerName]
	return ok
}

// getTimeoutValue gets the default timeout of the given timer
func (tm *timerManager) getTimeoutValue(timerName string) time.Duration {
	if !tm.containsTimer(timerName) {
		tm.logger.Warningf("Get tiemout failed!, timer %s not created yet! no time out", timerName)
		return 0 * time.Second
	}
	return tm.ttimers[timerName].timeout
}

// setTimeoutValue sets the default timeout of the given timer with a new timeout
func (tm *timerManager) setTimeoutValue(timerName string, d time.Duration) {
	if !tm.containsTimer(timerName) {
		tm.logger.Warningf("Set tiemout failed!, timer %s not created yet! no time out", timerName)
		return
	}
	tm.ttimers[timerName].timeout = d
}

// makeNullRequestTimeoutLegal checks if nullrequestTimeout is legal or not, if not, make it
// legal, which, nullRequest timeout must be larger than requestTimeout
func (tm *timerManager) makeNullRequestTimeoutLegal() {
	nullRequestTimeout := tm.getTimeoutValue(NULL_REQUEST_TIMER)
	requestTimeout := tm.getTimeoutValue(REQUEST_TIMER)

	if requestTimeout >= nullRequestTimeout && nullRequestTimeout != 0 {
		tm.setTimeoutValue(NULL_REQUEST_TIMER, 3*requestTimeout/2)
		tm.logger.Warningf("Configured null request timeout must be greater "+
			"than request timeout, setting to %v", tm.getTimeoutValue(NULL_REQUEST_TIMER))
	}

	if tm.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		tm.logger.Infof("RBFT null request timeout = %v", tm.getTimeoutValue(NULL_REQUEST_TIMER))
	} else {
		tm.logger.Infof("RBFT null request disabled")
	}
}

// makeCleanVcTimeoutLegal checks if requestTimeout is legal or not, if not, make it
// legal, which, request timeout must be lager than batch timeout
func (tm *timerManager) makeRequestTimeoutLegal() {
	requestTimeout := tm.getTimeoutValue(REQUEST_TIMER)
	batchTimeout := tm.getTimeoutValue(BATCH_TIMER)
	tm.logger.Infof("RBFT Batch timeout = %v", batchTimeout)

	if batchTimeout >= requestTimeout {
		tm.setTimeoutValue(REQUEST_TIMER, 3*batchTimeout/2)
		tm.logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", tm.getTimeoutValue(REQUEST_TIMER))
	}
	tm.logger.Infof("RBFT request timeout = %v", tm.getTimeoutValue(REQUEST_TIMER))
}

// makeCleanVcTimeoutLegal checks if clean vc timeout is legal or not, if not, make it
// legal, which, cleanVcTimeout should more then 6* viewChange time
func (tm *timerManager) makeCleanVcTimeoutLegal() {
	cleanVcTimeout := tm.getTimeoutValue(CLEAN_VIEW_CHANGE_TIMER)
	nvTimeout := tm.getTimeoutValue(NEW_VIEW_TIMER)

	if cleanVcTimeout < 6*nvTimeout {
		cleanVcTimeout = 6 * nvTimeout
		tm.setTimeoutValue(CLEAN_VIEW_CHANGE_TIMER, cleanVcTimeout)
		tm.logger.Criticalf("set timeout of cleaning out-of-time viewChange message to %v since it's too short", 6*nvTimeout)
	}
}
