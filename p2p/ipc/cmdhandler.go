package ipc

import (
	"encoding/json"
	"fmt"
	"context"
	"github.com/abiosoft/ishell"
)

type CMDHandler	struct {
	ipcEndpoint string
	callbacks map[string]func(args []string)error
}



func NewIPCHandler(endpoint string)*CMDHandler{
	return &CMDHandler{
		ipcEndpoint:endpoint,
	}
}

func(handler *CMDHandler)handle(c *ishell.Context){
	cmd := &IPCCmd{
		CMD:c.Cmd.Name,
		Args:c.Args,
	}
	cmdByte,err := json.Marshal(cmd)
	if err != nil{
		fmt.Printf("cannot handle this cmd, error: %s\n",err.Error())
		return
	}
	ipcClient,err := newIPCConnection(context.Background(),handler.ipcEndpoint)
	if err != nil {
		fmt.Printf("can't create IPC pipe, error: %s\n", err)
		return
	}
	ipcClient.Write(cmdByte)

}
