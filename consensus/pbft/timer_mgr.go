//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"github.com/op/go-logging"
	"time"
)

/**
this file provide a mechanism to manage different timers.
*/

//titletimer manage timer with the same timer name
type titletimer struct {
	timerName   string
	timeout     time.Duration
	alivecounts int
	isActive    []bool
}

//timerManager manage common used timer.
type timerManager struct {
	ttimers        map[string]*titletimer
	requestTimeout time.Duration
	logger         *logging.Logger
}

//newTimerMgr init a instance of timerManager.
func newTimerMgr(pbft *pbftImpl) *timerManager {
	tm := &timerManager{
		ttimers:        make(map[string]*titletimer),
		requestTimeout: pbft.config.GetDuration(PBFT_REQUEST_TIMEOUT),
	}

	return tm
}

//newTimer new a pbft timer by timer name and append into correspond map
func (tm *timerManager) newTimer(tname string, d time.Duration) {
	tm.ttimers[tname] = &titletimer{
		timerName:   tname,
		timeout:     d,
		alivecounts: 0,
	}
}

//Stop stop all timer managers.
func (tm *timerManager) Stop() {
	for timerName := range tm.ttimers {
		tm.stopTimer(timerName)
	}
}

//startTimer init and start a timer by name
func (tm *timerManager) startTimer(tname string, afterfunc func()) int {
	tm.stopTimer(tname)
	//logger.Errorf("Starting a new timer---%s", tname)

	tm.ttimers[tname].isActive = append(tm.ttimers[tname].isActive, true)
	tm.ttimers[tname].alivecounts++

	counts := len(tm.ttimers[tname].isActive)
	//logger.Errorf("Now exsits %d---%d timer---%s", counts, tm.ttimers[tname].alivecounts, tname)

	send := func() {
		if tm.ttimers[tname].isActive[counts-1] {
			afterfunc()
		}
	}
	time.AfterFunc(tm.ttimers[tname].timeout, send)
	return counts - 1
}

//softStartTimer init and start a timer by name
func (tm *timerManager) softStartTimer(tname string, afterfunc func()) int {
	if tm.ttimers[tname].alivecounts == 0 {
		//logger.Errorf("Soft starting a new timer---%s", tname)

		tm.ttimers[tname].isActive = append(tm.ttimers[tname].isActive, true)
		tm.ttimers[tname].alivecounts++

		counts := len(tm.ttimers[tname].isActive)
		//logger.Errorf("Now exsits %d---%d timer---%s", counts, tm.ttimers[tname].alivecounts, tname)

		send := func() {
			if tm.ttimers[tname].isActive[counts-1] {
				afterfunc()
			}
		}
		time.AfterFunc(tm.ttimers[tname].timeout, send)
		return counts - 1
	} else {
		return -1
	}
}

//startTimerWithNewTT init and start a timer by name with new timeout
func (tm *timerManager) startTimerWithNewTT(tname string, d time.Duration, afterfunc func()) int {
	tm.stopTimer(tname)
	//logger.Errorf("Starting a new timer---%s with new duration %d", tname, d)

	tm.ttimers[tname].isActive = append(tm.ttimers[tname].isActive, true)
	tm.ttimers[tname].alivecounts++

	counts := len(tm.ttimers[tname].isActive)
	//logger.Errorf("Now exsits %d---%d timer---%s", counts, tm.ttimers[tname].alivecounts, tname)

	send := func() {
		if tm.ttimers[tname].isActive[counts-1] {
			afterfunc()
		}
	}
	time.AfterFunc(d, send)
	return counts - 1
}

//softStartTimerWithNewTT init and start a timer by name with new timeout
func (tm *timerManager) softStartTimerWithNewTT(tname string, d time.Duration, afterfunc func()) int {
	if tm.ttimers[tname].alivecounts == 0 {
		//logger.Errorf("Soft starting a new timer---%s with new duration %d", tname, d)

		tm.ttimers[tname].isActive = append(tm.ttimers[tname].isActive, true)
		tm.ttimers[tname].alivecounts++

		counts := len(tm.ttimers[tname].isActive)
		//logger.Errorf("Now exsits %d---%d timer---%s", counts, tm.ttimers[tname].alivecounts, tname)

		send := func() {
			if tm.ttimers[tname].isActive[counts-1] {
				afterfunc()
			}
		}
		time.AfterFunc(d, send)
		return counts - 1
	} else {
		return -1
	}

}

//stopTimer stop all timers by the same timerName.
func (tm *timerManager) stopTimer(tname string) {
	if !tm.containsTimer(tname) {
		tm.logger.Errorf("Stop timer failed!, timer %s not created yet!", tname)
		return
	}

	for i := range tm.ttimers[tname].isActive {
		tm.ttimers[tname].isActive[i] = false
	}

	tm.ttimers[tname].alivecounts = 0

}

//stopOneTimer stop one timer by the timerName and index.
func (tm *timerManager) stopOneTimer(tname string, num int) {
	if !tm.containsTimer(tname) {
		tm.logger.Errorf("Stop timer failed!, timer %s not created yet!", tname)
		return
	}

	counts := len(tm.ttimers[tname].isActive)
	if num >= counts {
		//tm.logger.Errorf("Stop timer failed!, timer %s index out of range!", tname)
		return
	}

	tm.ttimers[tname].isActive[num] = false

	tm.ttimers[tname].alivecounts--

}

//containsTimer returns true if there exists a timer named timerName
func (tm *timerManager) containsTimer(timerName string) bool {
	_, ok := tm.ttimers[timerName]
	return ok
}

//getTimeoutValue get event timer timeout
func (tm *timerManager) getTimeoutValue(timerName string) time.Duration {
	if !tm.containsTimer(timerName) {
		tm.logger.Warningf("Get tiemout failed!, timer %s not created yet! no time out", timerName)
		return 0 * time.Second
	}
	return tm.ttimers[timerName].timeout
}

func (tm *timerManager) setTimeoutValue(timerName string, d time.Duration) {
	if !tm.containsTimer(timerName) {
		tm.logger.Warningf("Set tiemout failed!, timer %s not created yet! no time out", timerName)
		return
	}
	tm.ttimers[timerName].timeout = d
}

//makeRequestTimeoutLegal if requestTimeout is not legal, make it legal
func (tm *timerManager) makeRequestTimeoutLegal() {
	nullRequestTimeout := tm.getTimeoutValue(NULL_REQUEST_TIMER)
	requestTimeout := tm.requestTimeout

	if requestTimeout >= nullRequestTimeout && nullRequestTimeout != 0 {
		tm.setTimeoutValue(NULL_REQUEST_TIMER, 3*requestTimeout/2)
		tm.logger.Warningf("Configured null request timeout must be greater "+
			"than request timeout, setting to %v", tm.getTimeoutValue(NULL_REQUEST_TIMER))
	}
}
