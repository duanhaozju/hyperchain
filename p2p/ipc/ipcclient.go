package ipc

import (
	"net"
	"context"
)

func newIPCConnection(ctx context.Context, endpoint string)(net.Conn,error){
	dialer := net.Dialer{
		KeepAlive:tcpKeepAliveInterval,
	}
	return dialer.DialContext(ctx,"unix",endpoint)
}

