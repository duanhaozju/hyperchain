package network

import (
	"github.com/looplab/fsm"
	"fmt"
	"time"
	"hyperchain/p2p/message"
)

const (
	//EVENT
	c_EventConnect = "cEVconnect"
	c_EventError = "cEVerror"
	c_EventRecovery = "cEVrecovery"
	c_EventClose = "cEVclose"
	// STAT.
	c_StatCreated = "cSTcreated"
	c_StatWorking = "cSTworking"
	c_StatPending = "cSTpending"
	c_StatClosed = "cSTclosed"
)



func(c *Client)initState(){
	c.stateMachine = fsm.NewFSM(
		c_StatCreated,
		fsm.Events{
			{Name: c_EventConnect, Src: []string{c_StatCreated}, Dst: c_StatWorking},
			{Name: c_EventError, Src: []string{c_StatWorking}, Dst: c_StatPending},
			{Name: c_EventRecovery, Src: []string{c_StatPending,c_StatClosed}, Dst: c_StatWorking },
			{Name: c_EventClose, Src: []string{c_StatPending}, Dst: c_StatClosed},
		},
		fsm.Callbacks{
			"enter_"+c_StatCreated:  func(e *fsm.Event) { c.s_Create(e)  },
			"enter_"+c_StatWorking:  func(e *fsm.Event) { c.s_Working(e)  },
			"enter_"+c_StatPending:  func(e *fsm.Event) { c.s_Pending(e) },
			"enter_"+c_StatClosed:   func(e *fsm.Event) { c.s_Closed(e)  },
			//"before_event":        func(e *fsm.Event) { p.beforeEvent(e) },
		},
	)

}


func (c *Client)s_Create(e *fsm.Event){
	fmt.Println("client created")
}

func (c *Client)s_Working(e *fsm.Event){
	fmt.Println("client is working")
	c.working()
}

func (c *Client)s_Pending(e *fsm.Event){
	fmt.Println("client is pending")
	c.pending()
}

func (c *Client)s_Closed(e *fsm.Event){
	fmt.Println("client was closed")
}



// client stat change and keep methods
func (c *Client)working(){
	fmt.Println("start working loop")
	go func(c *Client){
		ticker := time.NewTicker(c.keepAliveDuration)
		counter := 0
		kap := message.NewPkg([]byte("keepalive"),message.ControlType_KeepAlive)
		for range ticker.C{
			if c.stateMachine.Current() != c_StatWorking{
				logger.Warning("unsuitable stat (working), ignore.")
				return
			}
			_,err := c.Discuss(kap)
			if err != nil{
				logger.Warningf("keepalive failed (to %s, times: %d)",c.addr,counter)
				counter++
				if counter > c.keepAliveFailTimes{
					logger.Warning("keepalive failed and connection change to pending (%s)",c.addr)
					c.stateMachine.Event(c_EventError)
					return
				}
			}else{
				logger.Noticef("keep alive successed. (%s)",c.hostname)
			}
		}
	}(c)
}

// when the connection is pending, keep recoverying
func (c *Client)pending(){
	go func(c *Client){
		ticker := time.NewTicker(c.pendingDuration)
		counter := 0
		kap := message.NewPkg([]byte("keepalive"),message.ControlType_Update)
		for range ticker.C{
			if c.stateMachine.Current() != c_StatPending{
				logger.Warning("unsuitable stat (pending), ignore.")
				return
			}
			_,err := c.Discuss(kap)
			if err != nil{
				logger.Warningf("recovery failed (to %s, times: %d)",c.addr,counter)
				counter++
				if counter > c.pendingFailTimes{
					logger.Warningf("recovery failed and connection change to closed (%s)",c.addr)
					c.stateMachine.Event(c_EventClose)
				}
			}else{
				c.reborn()
			}
		}
	}(c)
}

// reborn reset the connection
func (c *Client)reborn(){
	logger.Infof("recovery successful, and restart working (%s)",c.addr)
	c.stateMachine.Event(c_EventRecovery)
}
