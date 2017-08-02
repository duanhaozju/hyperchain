//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"hyperchain/common"
	"hyperchain/namespace"
	"strings"
	"encoding/json"
)

func (s *Server) handleCMD(req *common.RPCRequest) *common.RPCResponse {
	cmd := &Command{MethodName: req.Method}
	if args, ok := req.Params.(json.RawMessage); !ok {
		log.Notice("nil parms in json")
		cmd.Args = nil
	} else {
		args, err := splitRawMessage(args)
		if err != nil {
			return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Error: &common.InvalidParamsError{Message: err.Error()}}
		}
		cmd.Args = args
	}
	if _, ok := s.admin.CmdExecutor[req.Method] ; !ok {
		return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Error: &common.MethodNotFoundError{Service: req.Service, Method: req.Method}}
	}
	rs := s.admin.CmdExecutor[req.Method](cmd)
	if rs.Ok == false {
		return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Error: rs.Error}
	}
	return &common.RPCResponse{Id: req.Id, Namespace: req.Namespace, Reply: rs.Result}

}

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
		"{\"jsonrpc\":\"2.0\",\"method\":\"%s\",\"params\":[\"%s\"],\"id\":1}",
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
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 2 parameters, got %d", argLen)}}
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
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 3 parameters, got %d", argLen)}}
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

func (adm *Administrator) grantPermission(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen < 3 {
		log.Warningf("Invalid cmd nums %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects >=3 parameters, got %d", argLen)}}
	}
	invalidPms, err := grantpermission(cmd.Args[0], cmd.Args[1], cmd.Args[2:])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	if len(invalidPms) == 0 {
		return &CommandResult{Ok: true, Result: "grant permission successfully."}
	}
	return &CommandResult{Ok: true, Result: fmt.Sprintf("grant permission successfully, but there are some invalid permissions: %v", invalidPms)}
}

func (adm *Administrator) revokePermission(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen < 3 {
		log.Warningf("Invalid cmd nums %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects >=3 parameters, got %d", argLen)}}
	}
	invalidPms, err := revokepermission(cmd.Args[0], cmd.Args[1], cmd.Args[2:])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	if len(invalidPms) == 0 {
		return &CommandResult{Ok: true, Result: "revoke permission successfully."}
	}
	return &CommandResult{Ok: true, Result: fmt.Sprintf("revoke permission successfully, but there are some invalid permissions: %v", invalidPms)}
}

func (adm *Administrator) listPermission(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Warningf("Invalid cmd nums %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen)}}
	}
	result, err := listpermission(cmd.Args[0])
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: result}
}

func (adm *Administrator) createUser(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 3 {
		log.Warningf("Invalid cmd nums %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 2/3 parameters, got %d", argLen)}}
	}
	username := cmd.Args[0]
	password := cmd.Args[1]
	group := cmd.Args[2]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := IsUserExist(username, password); err != ErrUserNotExist {
		log.Debugf("User %s: %s", username, ErrDuplicateUsername.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrDuplicateUsername.Error()}}
	}

	err := createUser(username, password, group)
	if err != nil {
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "Create user successfully"}
}

func (adm *Administrator) alterUser(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 2 {
		log.Warningf("Invalid cmd nums %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen - 1)}}
	}
	username := cmd.Args[0]
	password := cmd.Args[1]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := IsUserExist(username, password); err == ErrUserNotExist {
		log.Debugf("User %s: %s", username, ErrUserNotExist.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrUserNotExist.Error()}}
	}

	alterUser(username, password)
	return &CommandResult{Ok: true, Result: "Alter user password successfully"}
}

func (adm *Administrator) delUser(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Warningf("Invalid cmd nums %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen)}}
	}
	username := cmd.Args[0]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := IsUserExist(username, ""); err == ErrUserNotExist {
		log.Debugf("User %s: %s", username, ErrUserNotExist.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrUserNotExist.Error()}}
	}

	delUser(username)
	return &CommandResult{Ok: true, Result: "Delete user successfully"}
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

	adm.CmdExecutor["grantPermission"] = adm.grantPermission
	adm.CmdExecutor["revokePermission"] = adm.revokePermission
	adm.CmdExecutor["listPermission"] = adm.listPermission

	adm.CmdExecutor["createUser"] = adm.createUser
	adm.CmdExecutor["alterUser"] = adm.alterUser
	adm.CmdExecutor["delUser"] = adm.delUser
}
