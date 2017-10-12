package network

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/looplab/fsm"
	"github.com/terasum/pool"
	"google.golang.org/grpc"
	"hyperchain/p2p/message"
	"regexp"
	"time"
)

const (
	//EVENT
	c_EventConnect  = "cEVconnect"
	c_EventError    = "cEVerror"
	c_EventRecovery = "cEVrecovery"
	c_EventClose    = "cEVclose"
	// STAT.
	c_StatCreated = "cSTcreated"
	c_StatWorking = "cSTworking"
	c_StatPending = "cSTpending"
	c_StatClosed  = "cSTclosed"
)

func (c *Client) initState() {
	c.stateMachine = fsm.NewFSM(
		c_StatCreated,
		fsm.Events{
			{Name: c_EventConnect, Src: []string{c_StatCreated, c_StatPending}, Dst: c_StatWorking},
			{Name: c_EventError, Src: []string{c_StatWorking}, Dst: c_StatPending},
			{Name: c_EventRecovery, Src: []string{c_StatPending, c_StatClosed}, Dst: c_StatWorking},
			{Name: c_EventClose, Src: []string{c_StatPending}, Dst: c_StatClosed},
		},
		fsm.Callbacks{
			"enter_" + c_StatCreated: func(e *fsm.Event) {
				c.s_Create(e)
			},
			"enter_" + c_StatWorking: func(e *fsm.Event) {
				c.s_Working(e)
			},
			"enter_" + c_StatPending: func(e *fsm.Event) {
				c.s_Pending(e)
			},
			"enter_" + c_StatClosed: func(e *fsm.Event) {
				c.s_Closed(e)
			},
			//"before_event":        func(e *fsm.Event) { p.beforeEvent(e) },
		},
	)

}

func (c *Client) s_Create(e *fsm.Event) {
	logger.Info("client created")
}

func (c *Client) s_Working(e *fsm.Event) {
	logger.Info("client is working")
	c.working()
}

func (c *Client) s_Pending(e *fsm.Event) {
	logger.Info("client is pending")
	c.pending()
}

func (c *Client) s_Closed(e *fsm.Event) {
	logger.Info("client was closed")
}

func (c *Client) working() {
	go func(c *Client) {
		ticker := time.NewTicker(c.cconf.keepAliveDuration)
		counter := 0
		kap := message.NewPkg([]byte("keepalive"), message.ControlType_KeepAlive)
		for range ticker.C {
			if c.stateMachine.Current() != c_StatWorking {
				logger.Warningf("unsuitable stat not working current is %s, ignore.", c.stateMachine.Current())
				return
			}
			_, err := c.Discuss(kap)
			if err != nil {
				logger.Warningf("keepalive failed (to %s, times: %d) rsp %s", c.addr, counter, err.Error())
				if ok, _ := regexp.MatchString(".+closing", err.Error()); ok {
					logger.Warningf("because of closing err, this client enforce into pending status")
					c.stateMachine.Event(c_EventError)
					return
				}
				counter++
				if counter > c.cconf.keepAliveFailTimes {
					logger.Warningf("keepalive failed and connection change to pending (%s)", c.addr)
					c.stateMachine.Event(c_EventError)
					return
				}
			} else {
				logger.Debugf("keep alive successed. (%s)", c.hostname)
			}
		}
	}(c)
}

// when the connection is pending, keep recoverying.
func (c *Client) pending() {
	go func(c *Client) {
		ticker := time.NewTicker(c.cconf.pendingDuration)
		counter := 0
		for range ticker.C {
			//this maybe have latency, when client status is working
			if c.stateMachine.Current() != c_StatPending {
				logger.Warningf("unsuitable stat not pending, current %s, ignore.", c.stateMachine.Current())
				return
			}
			if !c.ping(c.addr) {
				logger.Warningf("recovery failed (to %s, times: %d)", c.addr, counter)
				counter++
				if counter > c.cconf.pendingFailTimes {
					logger.Warningf("recovery failed and connection change to closed (%s)", c.addr)
					c.stateMachine.Event(c_EventClose)
				}
			} else {
				c.reborn()
			}
		}
	}(c)
}

// reborn resets the connection.
func (c *Client) reborn() {
	logger.Infof("recovery successful, and restart working (%s)", c.addr)
	c.connPool.Release()
	// recreate the connection pool
	poolConfig := &pool.PoolConfig{
		InitialCap: c.cconf.connInitCap,
		MaxCap:     c.cconf.connUpperlimit,
		Factory:    connCreator,
		Close:      connCloser,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: c.cconf.connIdleTime,
		EndPoint:    c.addr,
		Options:     c.sec.GetGrpcClientOpts(),
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		logger.Errorf("Fatal: cannot reborn this connection %s", err.Error())
		return
	}
	c.connPool = p
	c.stateMachine.Event(c_EventRecovery)
}

func (c *Client) ping(addr string) bool {
	logger.Infof("ping %s", addr)
	conn, err := grpc.Dial(addr, c.sec.GetGrpcClientOpts()...)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return false
	}
	client := NewChatClient(conn)
	in := message.NewPkg([]byte("ping"), message.ControlType_KeepAlive)
	// to ensure the retry times use the grpc_retry library
	_, err = client.Discuss(context.Background(), in, grpc.FailFast(true), grpc_retry.WithMax(1))
	if conn != nil {
		conn.Close()
	}
	if err != nil {
		return false
	}
	return true
}
