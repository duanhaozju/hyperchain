//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/namespace"
	"strings"
	"time"
)

// This file defines the hypercli admin interface. Users invoke
// admin services with the format Command as shown below and gets
// a result in the format CommandResult using JSON-RPC.

var log *logging.Logger

// Command defines the command format sent from client.
type Command struct {
	MethodName string   `json:"methodname"`
	Args       []string `json:"args"`
}

// ToJson transfers the command into JSON-RPC format string.
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

// CommandResult defines the result format sent back to the user.
type CommandResult struct {
	Ok     bool
	Result interface{}
	Error  common.RPCError
}

// Administrator manages the admin interface used by system
// administrator to manage hyperchain.
type Administrator struct {
	// checks token or not, used for test
	Check         bool

	// expiration records the expire timeout from last operation.
	Expiration    time.Duration

	// NsMgr is the global namespace Manager, used to get the system
	// level interface.
	NsMgr         namespace.NamespaceManager

	// CmdExecutor maps the interface name to the processing functions.
	CmdExecutor   map[string]func(command *Command) *CommandResult
}

// PreHandle is used to verify token, update and check user permission
// before handling admin services if admin.Check is true.
func (adm *Administrator) PreHandle(token, method string) error {
	if token == "" {
		return ErrTokenInvalid
	}

	if method == "" {
		return ErrNotSupport
	}
	// verify signed token.
	if claims, err := verifyToken(token, pub_key, "RS256"); err != nil {
		return err
	} else {
		username := getUserFromClaim(claims)
		if username == "" {
			return ErrPermission
		}
		// check if operation has expired or not, if expired, return error,
		// else update last operation time.
		if adm.checkOpTimeExpire(username) {
			return ErrTimeoutPermission
		}
		updateLastOperationTime(username)

		// check permission.
		if ok, err := adm.checkPermission(username, method); !ok {
			return err
		}
	}
	return nil
}

// StopServer stops hyperchain server, waiting for the command to be delivered
// successfully.
func (adm *Administrator) stopServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	adm.NsMgr.GetStopFlag() <- true
	return &CommandResult{Ok: true, Result: "stop server success!"}
}

// RestartServer restarts hyperchain server, waiting for the command to be delivered
// successfully.
func (adm *Administrator) restartServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	adm.NsMgr.GetRestartFlag() <- true
	return &CommandResult{Ok: true, Result: "restart server success!"}
}

// StartMgr starts namespace manager, waiting for the command to be executed
// successfully.
func (adm *Administrator) startNsMgr(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	err := adm.NsMgr.Start()
	if err != nil {
		log.Errorf("start namespace manager error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "start namespace manager successful"}
}

// StopNsMgr stops namespace manager, waiting for the command to be executed
// successfully.
func (adm *Administrator) stopNsMgr(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	err := adm.NsMgr.Stop()
	if err != nil {
		log.Errorf("stop namespace manager error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "stop namespace manager successful"}
}

// StartNamespace starts namespace by name, waiting for the command to be executed
// successfully.
func (adm *Administrator) startNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	err := adm.NsMgr.StartNamespace(cmd.Args[0])
	if err != nil {
		log.Errorf("start namespace error: %v", err)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: err.Error()}}
	}

	return &CommandResult{Ok: true, Result: "start namespace cmd executed!"}
}

// StartNamespace stops namespace by name, waiting for the command to be executed
// successfully.
func (adm *Administrator) stopNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	err := adm.NsMgr.StopNamespace(cmd.Args[0])
	if err != nil {
		log.Errorf("stop namespace error: %v", err)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: err.Error()}}
	}

	return &CommandResult{Ok: true, Result: "stop namespace cmd executed!"}
}

// RestartNamespace restarts namespace by name, waiting for the command to be executed
// successfully.
func (adm *Administrator) restartNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	err := adm.NsMgr.RestartNamespace(cmd.Args[0])
	if err != nil {
		log.Errorf("restart namespace error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}

	return &CommandResult{Ok: true, Result: "restart namespace successful"}
}

// RegisterNamespace registers a new namespace, waiting for the command to be executed
// successfully.
func (adm *Administrator) registerNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	err := adm.NsMgr.Register(cmd.Args[0])
	if err != nil {
		log.Errorf("register namespace error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}

	return &CommandResult{Ok: true, Result: "register namespace successful"}
}

// DeregisterNamespace deregisters a namespace, waiting for the command to be executed
// successfully.
func (adm *Administrator) deregisterNamespace(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}
	err := adm.NsMgr.DeRegister(cmd.Args[0])
	if err != nil {
		log.Errorf("deregister namespace error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "deregister namespace successful"}
}

// listNamespaces lists the namespaces participating in currently.
func (adm *Administrator) listNamespaces(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	names := adm.NsMgr.List()
	return &CommandResult{Ok: true, Result: names}
}

// GetLevel returns the log level with the given namespace and module.
func (adm *Administrator) getLevel(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 2 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message:
			fmt.Sprintf("Invalid parameter numbers, expects 2 parameters, got %d", argLen)}}
	}
	level, err := common.GetLogLevel(cmd.Args[0], cmd.Args[1])
	if err != nil {
		log.Errorf("get log level failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: level}
}

// SetLevel sets the log level of the given namespace and module.
func (adm *Administrator) setLevel(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 3 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message:
			fmt.Sprintf("Invalid parameter numbers, expects 3 parameters, got %d", argLen)}}
	}

	err := common.SetLogLevel(cmd.Args[0], cmd.Args[1], cmd.Args[2])
	if err != nil {
		log.Errorf("set log level failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	rs := strings.Join(cmd.Args, "_")
	return &CommandResult{Ok: true, Result: rs}
}

// startJvmServer starts the JVM service, waiting for the command to be executed
// successfully.
func (adm *Administrator) startJvmServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if err := adm.NsMgr.StartJvm(); err != nil {
		log.Noticef("start jvm server failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "start jvm successful."}
}

// stopJvmServer stops the JVM service, waiting for the command to be executed
// successfully.
func (adm *Administrator) stopJvmServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if err := adm.NsMgr.StopJvm(); err != nil {
		log.Noticef("stop jvm server failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "stop jvm successful."}
}

// restartJvmServer restarts the JVM service, waiting for the command to be executed
// successfully.
func (adm *Administrator) restartJvmServer(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	if err := adm.NsMgr.RestartJvm(); err != nil {
		log.Noticef("restart jvm server failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "restart jvm successful."}
}

// grantPermission grants the modular permissions to the given user.
func (adm *Administrator) grantPermission(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen < 3 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects >=3 parameters, got %d", argLen)}}
	}
	invalidPms, err := adm.grantpermission(cmd.Args[0], cmd.Args[1], cmd.Args[2:])
	if err != nil {
		log.Errorf("grant permission failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	if len(invalidPms) == 0 {
		return &CommandResult{Ok: true, Result: "grant permission successfully."}
	}
	return &CommandResult{Ok: true, Result: fmt.Sprintf("grant permission successfully, but there are some invalid permissions: %v", invalidPms)}
}

// revokePermission revokes the modular permissions from the given user.
func (adm *Administrator) revokePermission(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen < 3 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects >=3 parameters, got %d", argLen)}}
	}
	invalidPms, err := adm.revokepermission(cmd.Args[0], cmd.Args[1], cmd.Args[2:])
	if err != nil {
		log.Errorf("revoke permission failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	if len(invalidPms) == 0 {
		return &CommandResult{Ok: true, Result: "revoke permission successfully."}
	}
	return &CommandResult{Ok: true, Result: fmt.Sprintf("revoke permission successfully, but there are some invalid permissions: %v", invalidPms)}
}

// listPermission lists the permission of the given user.
func (adm *Administrator) listPermission(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen)}}
	}
	result, err := adm.listpermission(cmd.Args[0])
	if err != nil {
		log.Errorf("list permission failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: result}
}

// createUser is used by root user to create other users with the
// given password and group.
func (adm *Administrator) createUser(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 3 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 2/3 parameters, got %d", argLen)}}
	}
	username := cmd.Args[0]
	password := cmd.Args[1]
	group := cmd.Args[2]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := isUserExist(username, password); err != ErrUserNotExist {
		log.Debugf("User %s: %s", username, ErrDuplicateUsername.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrDuplicateUsername.Error()}}
	}

	err := adm.createuser(username, password, group)
	if err != nil {
		log.Errorf("create user failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "Create user successfully"}
}

// alterUser is used by all users to modify their password.
func (adm *Administrator) alterUser(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 2 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen-1)}}
	}
	username := cmd.Args[0]
	password := cmd.Args[1]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := isUserExist(username, password); err == ErrUserNotExist {
		log.Debugf("User %s: %s", username, ErrUserNotExist.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrUserNotExist.Error()}}
	}

	adm.alteruser(username, password)
	return &CommandResult{Ok: true, Result: "Alter user password successfully"}
}

// delUser is used by root user to delete user with the given name.
func (adm *Administrator) delUser(cmd *Command) *CommandResult {
	log.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		log.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen)}}
	}
	username := cmd.Args[0]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := isUserExist(username, ""); err == ErrUserNotExist {
		log.Errorf("User %s: %s", username, ErrUserNotExist.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrUserNotExist.Error()}}
	}

	adm.deluser(username)
	return &CommandResult{Ok: true, Result: "Delete user successfully"}
}

// Init initializes the CmdExecutor map.
func (adm *Administrator) Init() {
	log = common.GetLogger(common.DEFAULT_LOG, "jsonrpc/admin")
	log.Debugf("Start administrator with check permission: %v; expire " +
		"timeout: %v", adm.Check, adm.Expiration)

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
