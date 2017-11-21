package network

import (
	"time"
	"github.com/hyperchain/hyperchain/common"
)

type clientConf struct {
	// configurations
	keepAliveDuration  time.Duration
	keepAliveFailTimes int

	pendingDuration  time.Duration
	pendingFailTimes int

	//connection pool init capacity
	connInitCap int
	// connection pool upper limit
	connUpperlimit int
	// to avoid the EOF set the time
	connIdleTime time.Duration
}

func NewClientConf(common *common.Config) *clientConf {
	cconf := &clientConf{
		keepAliveDuration:  common.GetDuration("p2p.keepAliveDuration"),
		keepAliveFailTimes: common.GetInt("p2p.keepAliveFailTimes"),
		pendingDuration:    common.GetDuration("p2p.pendingDuration"),
		pendingFailTimes:   common.GetInt("p2p.pendingFailTimes"),
		connInitCap:        common.GetInt("p2p.connInitCap"),
		connUpperlimit:     common.GetInt("p2p.connUpperlimit"),
		connIdleTime:       common.GetDuration("p2p.connIdleTime"),
	}
	return cconf
}
