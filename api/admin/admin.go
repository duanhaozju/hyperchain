//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"fmt"
	"github.com/hyperchain/github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/common/interface"
	pb "github.com/hyperchain/hyperchain/common/protos"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/namespace"
	"github.com/op/go-logging"
	"strings"
	"time"
)

// This file defines the hypercli admin interface. Users invoke
// admin services with the format Command as shown below and gets
// a result in the format CommandResult using JSON-RPC.

// Command defines the command format sent from client.
type Command struct {
	MethodName string   `json:"methodname"`
	Args       []string `json:"args"`
}

// ToJson transfers the command into JSON-RPC format string.
func (cmd *Command) ToJson() string {
	var args = ""
	length := len(cmd.Args)
	if length > 0 {
		args = cmd.Args[0]
		if length > 1 {
			args = args + ","
		}
		for i := 1; i < length; i++ {
			if i == length-1 {
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

// permissionSet is used to maintain user permissions
type permissionSet map[int]bool

// Administrator manages the admin interface used by system
// administrator to manage hyperchain.
type Administrator struct {
	// CmdExecutor maps the interface name to the processing functions.
	CmdExecutor map[string]func(command *Command) *CommandResult

	// checks token or not, used for test
	check bool

	// expiration records the expire timeout from last operation.
	expiration time.Duration

	// nsMgr is the global namespace Manager, used to get the system
	// level interface.
	nsMgr namespace.NamespaceManager

	nsMgrProcessor intfc.NsMgrProcessor

	// valid_user records the current valid users with its password.
	valid_user map[string]string

	// user_scope records the current user with its permissions.
	user_scope map[string]permissionSet

	// user_opTimer records the current user with its last operation time.
	user_opTime map[string]int64

	config *common.Config
	logger *logging.Logger
}

// NewAdministrator news a raw administrator with default settings.
func NewAdministrator(nsMgrProcessor intfc.NsMgrProcessor, config *common.Config, is_executor bool) *Administrator {
	if is_executor {
		return nil
	}
	adm := &Administrator{
		CmdExecutor:    make(map[string]func(command *Command) *CommandResult),
		valid_user:     make(map[string]string),
		user_scope:     make(map[string]permissionSet),
		user_opTime:    make(map[string]int64),
		check:          config.GetBool(common.ADMIN_CHECK),
		expiration:     config.GetDuration(common.ADMIN_EXPIRATION),
		nsMgrProcessor: nsMgrProcessor,
		config:         config,
		nsMgr:          nsMgrProcessor.(namespace.NamespaceManager),
	}
	adm.init()
	return adm
}

// init initializes administrator with a root account who has all permissions
// and initializes the CmdExecutor map.
func (admin *Administrator) init() {
	admin.valid_user = map[string]string{
		"root": "hyperchain",
	}

	admin.user_scope = map[string]permissionSet{
		"root": rootScopes(),
	}

	admin.user_opTime = map[string]int64{
		"root": 0,
	}

	admin.logger = common.GetLogger(common.DEFAULT_LOG, "jsonrpc/admin")
	admin.logger.Debugf("Start administrator with check permission: %v; expire "+
		"timeout: %v", admin.check, admin.expiration)

	admin.CmdExecutor = make(map[string]func(command *Command) *CommandResult)
	admin.CmdExecutor["stopServer"] = admin.stopServer
	admin.CmdExecutor["restartServer"] = admin.restartServer

	admin.CmdExecutor["startNsMgr"] = admin.startNsMgr
	admin.CmdExecutor["stopNsMgr"] = admin.stopNsMgr
	admin.CmdExecutor["startNamespace"] = admin.startNamespace
	admin.CmdExecutor["stopNamespace"] = admin.stopNamespace
	admin.CmdExecutor["restartNamespace"] = admin.restartNamespace
	admin.CmdExecutor["registerNamespace"] = admin.registerNamespace
	admin.CmdExecutor["deregisterNamespace"] = admin.deregisterNamespace
	admin.CmdExecutor["listNamespaces"] = admin.listNamespaces

	admin.CmdExecutor["getLevel"] = admin.getLevel
	admin.CmdExecutor["setLevel"] = admin.setLevel

	admin.CmdExecutor["startJvmServer"] = admin.startJvmServer
	admin.CmdExecutor["stopJvmServer"] = admin.stopJvmServer
	admin.CmdExecutor["restartJvmServer"] = admin.restartJvmServer

	admin.CmdExecutor["grantPermission"] = admin.grantPermission
	admin.CmdExecutor["revokePermission"] = admin.revokePermission
	admin.CmdExecutor["listPermission"] = admin.listPermission

	admin.CmdExecutor["createUser"] = admin.createUser
	admin.CmdExecutor["alterUser"] = admin.alterUser
	admin.CmdExecutor["delUser"] = admin.delUser
}

// PreHandle is used to verify token, update and check user permission
// before handling admin services if admin.check is true.
func (admin *Administrator) PreHandle(token, method string) error {
	// if we don't need to check permission, return directly.
	if !admin.check {
		return nil
	}

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
		if admin.checkOpTimeExpire(username) {
			return ErrTimeoutPermission
		}
		admin.updateLastOperationTime(username)

		// check permission.
		if ok, err := admin.checkPermission(username, method); !ok {
			return err
		}
	}
	return nil
}

// StopServer stops hyperchain server, waiting for the command to be delivered
// successfully.
func (admin *Administrator) stopServer(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	admin.nsMgr.GetStopFlag() <- true
	return &CommandResult{Ok: true, Result: "stop server success!"}
}

// RestartServer restarts hyperchain server, waiting for the command to be delivered
// successfully.
func (admin *Administrator) restartServer(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	admin.nsMgr.GetRestartFlag() <- true
	return &CommandResult{Ok: true, Result: "restart server success!"}
}

// StartMgr starts namespace manager, waiting for the command to be executed
// successfully.
func (admin *Administrator) startNsMgr(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	err := admin.nsMgr.Start()
	if err != nil {
		admin.logger.Errorf("start namespace manager error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "start namespace manager successful"}
}

// StopNsMgr stops namespace manager, waiting for the command to be executed
// successfully.
func (admin *Administrator) stopNsMgr(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	err := admin.nsMgr.Stop()
	if err != nil {
		admin.logger.Errorf("stop namespace manager error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "stop namespace manager successful"}
}

// StartNamespace starts namespace by name, waiting for the command to be executed
// successfully.
func (admin *Administrator) startNamespace(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	go func() {
		err := admin.nsMgr.StartNamespace(cmd.Args[0])
		if err != nil {
			admin.logger.Errorf("start namespace error: %v", err)
			return
		}
	}()

	return &CommandResult{Ok: true, Result: "start namespace cmd executed!"}
}

// StartNamespace stops namespace by name, waiting for the command to be executed
// successfully.
func (admin *Administrator) stopNamespace(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	go func() {
		err := admin.nsMgr.StopNamespace(cmd.Args[0])
		if err != nil {
			admin.logger.Errorf("stop namespace error: %v", err)
			return
		}
	}()

	return &CommandResult{Ok: true, Result: "stop namespace cmd executed!"}
}

// RestartNamespace restarts namespace by name, waiting for the command to be executed
// successfully.
func (admin *Administrator) restartNamespace(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	go func() {
		err := admin.nsMgr.RestartNamespace(cmd.Args[0])
		if err != nil {
			admin.logger.Errorf("restart namespace error: %v", err)
			return
		}
	}()

	return &CommandResult{Ok: true, Result: "restart namespace cmd executed!"}
}

// RegisterNamespace registers a new namespace, waiting for the command to be executed
// successfully.
func (admin *Administrator) registerNamespace(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}

	err := admin.nsMgr.Register(cmd.Args[0])
	if err != nil {
		admin.logger.Errorf("register namespace error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}

	return &CommandResult{Ok: true, Result: "register namespace successful"}
}

// DeregisterNamespace deregisters a namespace, waiting for the command to be executed
// successfully.
func (admin *Administrator) deregisterNamespace(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: "Need only 1 param."}}
	}
	err := admin.nsMgr.DeRegister(cmd.Args[0])
	if err != nil {
		admin.logger.Errorf("deregister namespace error: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "deregister namespace successful"}
}

// listNamespaces lists the namespaces participating in currently.
func (admin *Administrator) listNamespaces(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	names := admin.nsMgr.List()
	return &CommandResult{Ok: true, Result: names}
}

// GetLevel returns the log level with the given namespace and module.
func (admin *Administrator) getLevel(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 2 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 2 parameters, got %d", argLen)}}
	}
	level, err := common.GetLogLevel(cmd.Args[0], cmd.Args[1])
	if err != nil {
		admin.logger.Errorf("get log level failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: level}
}

// SetLevel sets the log level of the given namespace and module.
func (admin *Administrator) setLevel(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 3 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 3 parameters, got %d", argLen)}}
	}

	err := common.SetLogLevel(cmd.Args[0], cmd.Args[1], cmd.Args[2])
	if err != nil {
		admin.logger.Errorf("set admin.logger level failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	rs := strings.Join(cmd.Args, "_")
	return &CommandResult{Ok: true, Result: rs}
}

// startJvmServer starts the JVM service, waiting for the command to be executed
// successfully.
func (admin *Administrator) startJvmServer(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	if admin.config.GetBool(common.EXECUTOR_EMBEDDED) {
		if err := admin.nsMgr.StartJvm(); err != nil {
			admin.logger.Noticef("start jvm server failed: %v", err)
			return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
		}
		return &CommandResult{Ok: true, Result: "start jvm successful."}
	}
	adminService := admin.nsMgr.InternalServer().ServerRegistry().AdminService(cmd.Args[0])
	msg := pb.IMessage{
		From:  pb.FROM_ADMINISTRATOR,
		Type:  pb.Type_ADMIN,
		Event: pb.Event_StartJVMEvent,
	}
	rsp, _ := adminService.SyncSend(msg)
	rspEvent := &event.AdminResponseEvent{}
	err := proto.Unmarshal(rsp.Payload, rspEvent)
	if err != nil {
		return &CommandResult{Ok: false, Result: "start jvm failed."}
	} else if rspEvent.Ok {
		return &CommandResult{Ok: true, Result: "start jvm successful."}
	} else {
		return &CommandResult{Ok: false, Result: "start jvm failed."}
	}

}

// stopJvmServer stops the JVM service, waiting for the command to be executed
// successfully.
func (admin *Administrator) stopJvmServer(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	if admin.config.GetBool(common.EXECUTOR_EMBEDDED) {
		if err := admin.nsMgr.StopJvm(); err != nil {
			admin.logger.Noticef("stop jvm server failed: %v", err)
			return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
		}
		return &CommandResult{Ok: true, Result: "stop jvm successful."}
	}
	adminService := admin.nsMgr.InternalServer().ServerRegistry().AdminService(cmd.Args[0])

	msg := pb.IMessage{
		From:  pb.FROM_ADMINISTRATOR,
		Type:  pb.Type_ADMIN,
		Event: pb.Event_StopJVMEvent,
	}
	rsp, _ := adminService.SyncSend(msg)
	rspEvent := &event.AdminResponseEvent{}
	err := proto.Unmarshal(rsp.Payload, rspEvent)
	if err != nil {
		return &CommandResult{Ok: false, Result: "start jvm failed."}
	} else if rspEvent.Ok {
		return &CommandResult{Ok: true, Result: "start jvm successful."}
	} else {
		return &CommandResult{Ok: false, Result: "start jvm failed."}
	}
}

// restartJvmServer restarts the JVM service, waiting for the command to be executed
// successfully.
func (admin *Administrator) restartJvmServer(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	if admin.config.GetBool(common.EXECUTOR_EMBEDDED) {
		if err := admin.nsMgr.RestartJvm(); err != nil {
			admin.logger.Noticef("restart jvm server failed: %v", err)
			return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
		}
		return &CommandResult{Ok: true, Result: "restart jvm successful."}
	}
	adminService := admin.nsMgr.InternalServer().ServerRegistry().AdminService(cmd.Args[0])

	msg := pb.IMessage{
		From:  pb.FROM_ADMINISTRATOR,
		Type:  pb.Type_ADMIN,
		Event: pb.Event_RestartJVMEvent,
	}
	rsp, _ := adminService.SyncSend(msg)
	rspEvent := &event.AdminResponseEvent{}
	err := proto.Unmarshal(rsp.Payload, rspEvent)
	if err != nil {
		return &CommandResult{Ok: false, Result: "start jvm failed."}
	} else if rspEvent.Ok {
		return &CommandResult{Ok: true, Result: "start jvm successful."}
	} else {
		return &CommandResult{Ok: false, Result: "start jvm failed."}
	}
}

// grantPermission grants the modular permissions to the given user.
func (admin *Administrator) grantPermission(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen < 3 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects >=3 parameters, got %d", argLen)}}
	}
	invalidPms, err := admin.grantpermission(cmd.Args[0], cmd.Args[1], cmd.Args[2:])
	if err != nil {
		admin.logger.Errorf("grant permission failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	if len(invalidPms) == 0 {
		return &CommandResult{Ok: true, Result: "grant permission successfully."}
	}
	return &CommandResult{Ok: true, Result: fmt.Sprintf("grant permission successfully, but there are some invalid permissions: %v", invalidPms)}
}

// revokePermission revokes the modular permissions from the given user.
func (admin *Administrator) revokePermission(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen < 3 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects >=3 parameters, got %d", argLen)}}
	}
	invalidPms, err := admin.revokepermission(cmd.Args[0], cmd.Args[1], cmd.Args[2:])
	if err != nil {
		admin.logger.Errorf("revoke permission failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	if len(invalidPms) == 0 {
		return &CommandResult{Ok: true, Result: "revoke permission successfully."}
	}
	return &CommandResult{Ok: true, Result: fmt.Sprintf("revoke permission successfully, but there are some invalid permissions: %v", invalidPms)}
}

// listPermission lists the permission of the given user.
func (admin *Administrator) listPermission(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen)}}
	}
	result, err := admin.listpermission(cmd.Args[0])
	if err != nil {
		admin.logger.Errorf("list permission failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: result}
}

// createUser is used by root user to create other users with the
// given password and group.
func (admin *Administrator) createUser(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 3 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 2/3 parameters, got %d", argLen)}}
	}
	username := cmd.Args[0]
	password := cmd.Args[1]
	group := cmd.Args[2]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := admin.isUserExist(username, password); err != ErrUserNotExist {
		admin.logger.Debugf("User %s: %s", username, ErrDuplicateUsername.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrDuplicateUsername.Error()}}
	}

	err := admin.createuser(username, password, group)
	if err != nil {
		admin.logger.Errorf("create user failed: %v", err)
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: err.Error()}}
	}
	return &CommandResult{Ok: true, Result: "Create user successfully"}
}

// alterUser is used by all users to modify their password.
func (admin *Administrator) alterUser(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 2 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen-1)}}
	}
	username := cmd.Args[0]
	password := cmd.Args[1]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := admin.isUserExist(username, password); err == ErrUserNotExist {
		admin.logger.Debugf("User %s: %s", username, ErrUserNotExist.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrUserNotExist.Error()}}
	}

	admin.alteruser(username, password)
	return &CommandResult{Ok: true, Result: "Alter user password successfully"}
}

// delUser is used by root user to delete user with the given name.
func (admin *Administrator) delUser(cmd *Command) *CommandResult {
	admin.logger.Noticef("process cmd %v", cmd.MethodName)
	argLen := len(cmd.Args)
	if argLen != 1 {
		admin.logger.Errorf("Invalid arg numbers: %d", argLen)
		return &CommandResult{Ok: false, Error: &common.InvalidParamsError{Message: fmt.Sprintf("Invalid parameter numbers, expects 1 parameters, got %d", argLen)}}
	}
	username := cmd.Args[0]
	// judge if the user exist or not, if username exists, return a duplicate name error
	if _, err := admin.isUserExist(username, ""); err == ErrUserNotExist {
		admin.logger.Errorf("User %s: %s", username, ErrUserNotExist.Error())
		return &CommandResult{Ok: false, Error: &common.CallbackError{Message: ErrUserNotExist.Error()}}
	}

	admin.deluser(username)
	return &CommandResult{Ok: true, Result: "Delete user successfully"}
}
