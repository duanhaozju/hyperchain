package jvm

import (
	"github.com/looplab/fsm"
	"time"
	"github.com/op/go-logging"
)


const (
	conn_init   = "init"
	conn_health = "health"
	conn_sick   = "sick"
)

const (
	initialize       = "initialize"
	ping_success     = "ping_success"
	ping_failed      = "ping_failed"
	rebirth          = "rebirth"
)

const retryThreshold = 5

type ConnMaintainer struct {
	fsm      *fsm.FSM
	cli      *contractExecutorImpl

	rc       int32
	pc       int32
	exit     chan bool
	logger   *logging.Logger
}

func NewConnMaintainer(cli *contractExecutorImpl, logger *logging.Logger) *ConnMaintainer {
	exit := make(chan bool)
	mr :=  &ConnMaintainer{
		cli:    cli,
		exit:   exit,
		logger: logger,
	}
	mr.fsm = fsm.NewFSM(
		conn_init,
		fsm.Events{
			{Name: initialize, Src: []string{conn_init}, Dst: conn_health},
			{Name: ping_success, Src: []string{conn_health, conn_sick}, Dst: conn_health},
			{Name: ping_failed, Src: []string{conn_health, conn_sick}, Dst: conn_sick},
			{Name: rebirth, Src: []string{conn_sick}, Dst: conn_health},
		},
		fsm.Callbacks{
			"before_" + ping_success: func(e *fsm.Event) {mr.pingS(e)},
			"before_" + ping_failed: func(e *fsm.Event) {mr.pingF(e)},
			"after_" + ping_failed: func(e *fsm.Event) {mr.reconn(e)},
		},
	)
	mr.fsm.Event(initialize)
	return mr
}

func (maintainer *ConnMaintainer) Serve() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-maintainer.exit:
			maintainer.logger.Notice("conn maintainer exit")
			return
		case <-ticker.C:
			switch maintainer.fsm.Current() {
			case conn_health:
				if err := maintainer.ping(); err != nil {
					maintainer.fsm.Event(ping_failed)
				} else {
					maintainer.fsm.Event(ping_success)
				}
			case conn_sick:
				if err := maintainer.ping(); err != nil {
					maintainer.fsm.Event(ping_failed)
				} else {
					maintainer.fsm.Event(ping_success)

					if maintainer.pc == 0 {
						maintainer.fsm.Event(rebirth)
					}
				}

			default:
				maintainer.logger.Noticef("current fsm state %s, no need to send heartbeat", maintainer.fsm.Current())
			}
		}
	}
}

func (maintainer *ConnMaintainer) Exit() {
	maintainer.exit <- true
}

func (maintainer *ConnMaintainer) ping() error {
	_, err := maintainer.cli.heartbeat()
	if err != nil {
		return err
	} else {
		return nil
	}
}


func (maintainer *ConnMaintainer) conn() error {
	return maintainer.cli.client.Connect()
}

/*
	Callbacks
 */

func (maintainer *ConnMaintainer) pingS(e *fsm.Event) {
	maintainer.logger.Debug("ping success")
	if maintainer.pc > 0 {
		maintainer.pc -= 1
	}
}

func (maintainer *ConnMaintainer) pingF(e *fsm.Event) {
	maintainer.logger.Debug("ping failed")
	maintainer.pc += 1
}


func (maintainer *ConnMaintainer) reconn(e *fsm.Event) {
	if maintainer.pc < retryThreshold || !maintainer.fsm.Is(conn_sick) {
		maintainer.logger.Debugf("try to reconnect to %s, doesn't satisify.", maintainer.cli.Address())
		return
	}
	if err := maintainer.conn(); err != nil {
		maintainer.logger.Warningf("connect to %s failed", maintainer.cli.Address())
		return
	} else {
		maintainer.logger.Debugf("connect to %s success", maintainer.cli.Address())
		maintainer.pc = 0
		return
	}
}

func (maintainer *ConnMaintainer) trace(e *fsm.Event) {
	maintainer.logger.Debugf("[FSM TRACE] event %s", e.Event)
}

