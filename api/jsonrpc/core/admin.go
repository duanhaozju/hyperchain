//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"hyperchain/common"
	"hyperchain/namespace"
	"strings"
)

//Command command send from client.
type Command struct {
	MethodName string   `json:"methodname"`
	Args       []string `json:"args"`
}

//ToJson transform command into jsonrpc format string
func (cmd *Command) ToJson() string {
	var args = ""
	len := len(cmd.Args)
	if len > 0 {
		args = cmd.Args[0]
		if len > 1 {
			args = args + ","
		}
		for i := 1; i < len; i++ {
			if i == len-1 {
				args = args + cmd.Args[i]
			} else {
				args = args + cmd.Args[i] + ","
			}
		}
	}
	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"method\":\"%s\",\"params\":\"[%s]\",\"id\":1}",
		cmd.MethodName, args)

}

//CommandResult command execute result send back to the user.
type CommandResult struct {
	Ok     bool
	Result interface{}
	Error  common.RPCError
}

type Administrator struct {
	NsMgr         namespace.NamespaceManager
	CmdExecutor   map[string]func(command *Command) *CommandResult
	StopServer    chan bool
	RestartServer chan bool
}

//StopServer stop hyperchain server.
func (adm *Administrator) stopServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	adm.StopServer <- true
	return &CommandResult{Ok: true, Result: "stop server success!"}
}

//RestartServer start hyperchain server.
func (adm *Administrator) restartServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	adm.RestartServer <- true
	return &CommandResult{Ok: true, Result: "restart server success!"}
}

//StartMgr start namespace manager.
func (adm *Administrator) startNsMgr(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	err := adm.NsMgr.Start()
	if err != nil {
		log.Errorf("start namespace manager error %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "start namespace manager successful"}
}

//StopNsMgr stop namespace manager.
func (adm *Administrator) stopNsMgr(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	err := adm.NsMgr.Stop()
	if err != nil {
		log.Errorf("stop namespace manager error %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "stop namespace manager successful"}
}

//StartNamespace start namespace by name.
func (adm *Administrator) startNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if len(cmd.Args) != 1 {
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}
	//TODO: and synchronized start command
	go func() {
		err := adm.NsMgr.StartNamespace(cmd.Args[0])
		if err != nil {
			log.Error(err)
		}
	}()

	return &CommandResult{Ok: true, Result: "start namespace cmd executed!"}
}

//StartNamespace stop namespace by name.
func (adm *Administrator) stopNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if len(cmd.Args) != 1 {
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}
	go func() {
		err := adm.NsMgr.StopNamespace(cmd.Args[0])
		if err != nil {
			log.Error(err)
		}
	}()

	return &CommandResult{Ok: true, Result: "stop namespace cmd executed!"}
}

//RestartNamespace restart namespace by name.
func (adm *Administrator) restartNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if len(cmd.Args) != 1 {
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	err := adm.NsMgr.RestartNamespace(cmd.Args[0])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}

	return &CommandResult{Ok: true, Result: "restart namespace successful"}
}

//RegisterNamespace register a new namespace.
func (adm *Administrator) registerNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if len(cmd.Args) != 1 {
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	err := adm.NsMgr.Register(cmd.Args[0])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}

	return &CommandResult{Ok: true, Result: "register namespace successful"}
}

//DeregisterNamespace deregister a namespace
func (adm *Administrator) deregisterNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if len(cmd.Args) != 1 {
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}
	err := adm.NsMgr.DeRegister(cmd.Args[0])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "deregister namespace successful"}
}

func (adm *Administrator) listNamespaces(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	names := adm.NsMgr.List()
	return &CommandResult{Ok: true, Result: names}
}

//GetLevel get a log level.
func (adm *Administrator) getLevel(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 2 {
		log.Errorf("Invalid cmd nums %d", argLen)
	}
	level, err := common.GetLogLevel(cmd.Args[0], cmd.Args[1])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: level}
}

//SetLevel set a module log level.
func (adm *Administrator) setLevel(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 3 {
		log.Errorf("Invalid cmd nums %d", argLen)
	}

	err := common.SetLogLevel(cmd.Args[0], cmd.Args[1], cmd.Args[2])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	rs := strings.Join(cmd.Args, "_")
	return &CommandResult{Ok: true, Result: rs}
}

func (adm *Administrator) startJvmServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if err := adm.NsMgr.StartJvm(); err != nil {
		log.Notice(err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "start jvm successful."}
}

func (adm *Administrator) stopJvmServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if err := adm.NsMgr.StopJvm(); err != nil {
		log.Notice(err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "stop jvm successful."}
}

func (adm *Administrator) restartJvmServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if err := adm.NsMgr.RestartJvm(); err != nil {
		log.Notice(err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "restart jvm successful."}
}

func (adm *Administrator) Init() {
	adm.CmdExecutor = make(map[string]func(command *Command) *CommandResult)
	adm.CmdExecutor["stopServer"] = adm.stopServer
	adm.CmdExecutor["restartServer"] = adm.restartServer

	adm.CmdExecutor["startNsMgr"] = adm.startNsMgr
	adm.CmdExecutor["stopNsMgr"] = adm.stopNsMgr
	adm.CmdExecutor["startNamespace"] = adm.startNamespace
	adm.CmdExecutor["stopNamespace"] = adm.stopNamespace
	adm.CmdExecutor["restartNamespace"] = adm.restartNamespace
	adm.CmdExecutor["registerNamespace"] = adm.registerNamespace
	adm.CmdExecutor["deregisterNamespace"] = adm.deregisterNamespace
	adm.CmdExecutor["listNamespaces"] = adm.listNamespaces

	adm.CmdExecutor["getLevel"] = adm.getLevel
	adm.CmdExecutor["setLevel"] = adm.setLevel

	adm.CmdExecutor["startJvmServer"] = adm.startJvmServer
	adm.CmdExecutor["stopJvmServer"] = adm.stopJvmServer
	adm.CmdExecutor["restartJvmServer"] = adm.restartJvmServer
}
