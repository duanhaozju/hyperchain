package ipc

import (
	"github.com/abiosoft/ishell"
)

type IPCCmd struct {
	CMD string `json:"cmd"`
	Args []string `json:"args"`
}

func getCmds(endpoint string)[]*ishell.Cmd{
	handler := NewIPCHandler(endpoint)
	return []*ishell.Cmd{
		{
			Name: "network",
			Help: `network commands:
		network list	- show all connections and status
		network connect [hostname@ip:port]	- create a new connection by hostname and addr
		network close  [hostname]	- close the connection to hostname
		network reconnect  [hostname]	- reconnect to hostname
			`,
			Func: handler.handle,
		},
		{
			Name: "dn",
			Help: "hello user",
			Func: handler.handle,
		},

	}
}




