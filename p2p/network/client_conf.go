package network

import (
	"time"
	"github.com/terasum/viper"
)

type clientConf struct {
	// configurations
	keepAliveDuration time.Duration
	keepAliveFailTimes int

	pendingDuration time.Duration
	pendingFailTimes int

	//connection pool init capacity
	connInitCap int
	// connection pool upper limit
	connUpperlimit int
	// to avoid the EOF set the time
	connIdleTime time.Duration
}

func NewClientConf(vip *viper.Viper) *clientConf{
	cconf := &clientConf{
		keepAliveDuration:vip.GetDuration("p2p.keepAliveDuration"),
		keepAliveFailTimes:vip.GetInt("p2p.keepAliveFailTimes"),
		pendingDuration:vip.GetDuration("p2p.pendingDuration"),
		pendingFailTimes:vip.GetInt("p2p.pendingFailTimes"),
		connInitCap:vip.GetInt("p2p.connInitCap"),
		connUpperlimit:vip.GetInt("p2p.connUpperlimit"),
		connIdleTime:vip.GetDuration("p2p.connIdleTime"),
	}
	return cconf
}
