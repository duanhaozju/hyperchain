package ipc

import (
	"net/rpc"
)

func newIPCConnection(endpoint string) (*rpc.Client, error) {
	return rpc.DialHTTP("unix", endpoint)
}
