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
		network close  [hostname]	- close the connection to host
		network reconnect  [hostname]	- reconnect to host
		network ping  [hostname]	- ping to host and get the latency
			`,
			Func: handler.handle,
		},
		{
			Name: "dn",
			Help: "hello user",
			Func: handler.handle,
		},
		{
			Name: "service",
			Help: `service commands:
		service http start [port] 	- start http service
		service http stop	- stop http service
		service http restart	- restart http service
		service websocket start [port] 	- start http service
		service websocket stop	- stop http service
		service websocket restart	- restart http service
			`,
			Func: handler.handle,
		},

	}
}




