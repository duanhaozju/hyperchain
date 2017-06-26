package ipc

import (
	"github.com/abiosoft/ishell"
)

type IPCCmd struct {
	CMD string `json:"cmd"`
	Args []string `json:"args"`
}

func getCmds()[]*ishell.Cmd{
	handler := NewIPCHandler("./hpc.ipc")
	return []*ishell.Cmd{
		{
			Name: "greet",
			Help: "greet user",
			Func: handler.handle,
		},
		{
			Name: "hello",
			Help: "hello user",
			Func: handler.handle,
		},

	}
}




