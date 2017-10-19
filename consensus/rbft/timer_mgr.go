//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"github.com/op/go-logging"
	"hyperchain/manager/event"
	"time"
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
	ttimers        map[string]*titletimer
	requestTimeout time.Duration
	logger         *logging.Logger
}

// newTimerMgr news an instance of timerManager.
func newTimerMgr(rbft *rbftImpl) *timerManager {
	tm := &timerManager{
		ttimers:        make(map[string]*titletimer),
		requestTimeout: rbft.config.GetDuration(RBFT_REQUEST_TIMEOUT),
		logger:         rbft.logger,
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

// makeRequestTimeoutLegal checks if requestTimeout is legal or not, if not, make it legal, which, nullRequest timeout
// must be larger than requestTimeout
func (tm *timerManager) makeRequestTimeoutLegal() {
	nullRequestTimeout := tm.getTimeoutValue(NULL_REQUEST_TIMER)
	requestTimeout := tm.requestTimeout

	if requestTimeout >= nullRequestTimeout && nullRequestTimeout != 0 {
		tm.setTimeoutValue(NULL_REQUEST_TIMER, 3*requestTimeout/2)
		tm.logger.Warningf("Configured null request timeout must be greater "+
			"than request timeout, setting to %v", tm.getTimeoutValue(NULL_REQUEST_TIMER))
	}
}
