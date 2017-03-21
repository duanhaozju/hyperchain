//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"fmt"
)

//Command command send from client.
type Command struct {
	MethodName string   `json:"methodname"`
	Args       []string `json:"args"`
}

func (cmd *Command) ToJson() string {
	//{"jsonrpc":"2.0","method":"block_stopServer","params":"","id":1}'
	var args = ""
	len := len(cmd.Args)
	for i, arg := range cmd.Args {
		if i == 0 {
			args = "[" + arg
		}
		if i == len-1 {
			args = args + "]"
		}
		if i > 0 && i < len-1 {
			args = args + "," + arg
		}
	}
	fmt.Println(args)
	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"method\":\"%s\",\"params\":\"%s\",\"id\":1}'",
		cmd.MethodName, args)

}

//CommandResult command execute result send back to the user.
type CommandResult struct {
	Ok     bool   `json:"ok"`
	Result string `json:"result"`
}

type Administrator struct {
}

func NewAdminAPI() *Administrator {
	return &Administrator{}
}

//StopServer stop hyperchain server.
func (adm *Administrator) StopServer() *CommandResult {
	//TODO: stop hyperchain server
	log.Critical("stop hyperchain server")
	return &CommandResult{Ok: true, Result: "stop success"}
}

//RestartServer start hyperchain server.
func (adm *Administrator) RestartServer() *CommandResult {
	//TODO: restart hyperchain server
	log.Critical("restart hyperchain server")
	return nil
}

//StartMgr start namespace manager.
func (adm *Administrator) StartNsMgr() *CommandResult {
	return nil
}

//StopNsMgr stop namespace manager.
func (adm *Administrator) StopNsMgr() *CommandResult {
	return nil
}

//StartNamespace start namespace by name.
func (adm *Administrator) StartNamespace(name string) *CommandResult {
	return nil
}

//StartNamespace stop namespace by name.
func (adm *Administrator) StopNamespace(name string) *CommandResult {
	return nil
}

//RestartNamespace restart namespace by name.
func (adm *Administrator) RestartNamespace(name string) *CommandResult {
	return nil
}

//RegisterNamespace register a new namespace.
func (adm *Administrator) RegisterNamespace(name string) *CommandResult {
	return nil
}

//DeregisterNamespace deregister a namespace
func (adm *Administrator) DeregisterNamespace(name string) *CommandResult {
	return nil
}

//GetLevel get a log level.
func (adm *Administrator) GetLevel(namespace, module string) *CommandResult {
	return nil
}

//SetLevel set a module log level.
func (adm *Administrator) SetLevel(name, module, loglevel string) *CommandResult {
	return nil
}

//TODO: contract method?
