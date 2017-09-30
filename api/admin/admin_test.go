package jsonrpc

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"hyperchain/namespace/mocks"

	"hyperchain/common"
	"time"
	"strings"
	"errors"
)

var keyPath = "../../hypercli/keyconfigs/key/key"
var mockError = errors.New("mock error")

func init() {
	common.InitRawHyperLogger(common.DEFAULT_LOG)
	common.SetLogLevel(common.DEFAULT_LOG, "jsonrpc/admin", "CRITICAL")
}

func TestCommand_ToJson(t *testing.T) {
	ast := assert.New(t)

	cmd := &Command{
		MethodName: "admin_setLevel",
		Args:       []string{"global", "consensus", "error"},
	}
	jsonCmd := cmd.ToJson()
	expect := `{"jsonrpc":"2.0","method":"admin_setLevel","params":["global,consensus,error"],"id":1}`

	ast.JSONEq(expect, jsonCmd, "This two json string should be equal.")
}

func TestAdministrator_PreHandle(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	var err error

	err = admin.PreHandle("", "admin_getLevel")
	ast.EqualError(err, ErrTokenInvalid.Error(), "An InValidToken error was expected!")

	err = admin.PreHandle("balabala", "")
	ast.EqualError(err, ErrNotSupport.Error(), "A Not Support error was expected!")

	err = admin.PreHandle("balabala", "admin_getLevel")
	ast.EqualError(err, ErrTokenInvalid.Error(), "An InValidToken error was expected!")

	token, err := signToken("root", keyPath, "RS256")
	if err != nil {
		t.Error("Faild in sign token", err)
	}
	pub_key = "../../hypercli/keyconfigs/key/key.pub"
	// here we don't set the last operation time of root.
	err = admin.PreHandle(token, "admin_getLevel")
	ast.EqualError(err, ErrTimeoutPermission.Error(), "An Timeout error was expected!")

	// mock: last operation time of root is just now.
	admin.user_opTime["root"] = time.Now().Unix()
	err = admin.PreHandle(token, "admin_getLevel")
	ast.Nil(err, "No error was expected!")
	// sleep a duration with expiration seconds
	time.Sleep(admin.expiration)
	time.Sleep(admin.expiration)
	err = admin.PreHandle(token, "admin_getLevel")
	ast.EqualError(err, ErrTimeoutPermission.Error(), "An Timeout error was expected!")
}

func TestStopServer(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr
	stop := make(chan bool)
	mockNSMgr.On("GetStopFlag").Return(stop)

	// start a go-routine to listen the stop flag.
	go func(){
		flag := <- admin.nsMgr.GetStopFlag()
		t.Logf("Got stop flag: %v", flag)
	}()

	cmd := &Command{
		MethodName: "admin_stopServer",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully stop the hyperchain server.")
}

func TestRestartServer(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr
	restart := make(chan bool)
	mockNSMgr.On("GetRestartFlag").Return(restart)

	// start a go-routine to listen the restart flag.
	go func(){
		flag := <- admin.nsMgr.GetRestartFlag()
		t.Logf("Got restart flag: %v", flag)
	}()

	cmd := &Command{
		MethodName: "admin_restartServer",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully restart the hyperchain server.")
}

func TestStartNsMgr(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	// mock: start nsMgr with no error return, limit only 1 return time
	// as we need to call Start() interface below two times, if we don't
	// limit time here, next mockNSMgr.On(...).Return(...) will be
	// replaced by this one.
	mockNSMgr.On("Start").Return(nil).Times(1)
	cmd := &Command{
		MethodName: "admin_startNsMgr",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully start the namespace manager.")

	// mock: start nsMgr with an error, do not need to limit return times.
	mockNSMgr.On("Start").Return(mockError)
	result = admin.CmdExecutor[service](cmd)
	ast.Equal(result.Error.Error(), mockError.Error(), "We should fail to start " +
		"the namespace manager because of some errors.")
}

func TestStopNsMgr(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	// mock: stop nsMgr with no error return, limit only 1 return time
	// as we need to call Stop() interface below two times, if we don't
	// limit time here, next mockNSMgr.On(...).Return(...) will be
	// replaced by this one.
	mockNSMgr.On("Stop").Return(nil).Times(1)
	cmd := &Command{
		MethodName: "admin_stopNsMgr",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully stop the namespace manager.")

	// mock: stop nsMgr with an error, do not need to limit return times.
	mockNSMgr.On("Stop").Return(mockError)
	result = admin.CmdExecutor[service](cmd)
	ast.Equal(result.Error.Error(), mockError.Error(), "We should fail to stop " +
		"the namespace manager because of some errors.")
}

func TestStartNamespace(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	mockNSMgr.On("StartNamespace", "global").Return(nil)
	mockNSMgr.On("StartNamespace", "mock_ns").Return(mockError)

	// mock: start namespace with invalid arg numbers.
	cmd := &Command{
		MethodName: "admin_startNamespace",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	invalidParamsErr := &common.InvalidParamsError{}
	ast.Equal(invalidParamsErr.Code(), result.Error.Code(),
		"We should fail to start the namespace as we provide 0 args.")

	// mock: start namespace with no error return.
	cmd = &Command{
		MethodName: "admin_startNamespace",
		Args:       []string{"global"},
	}
	service = getService(cmd.MethodName, t)
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully start the namespace <global>.")

	// mock: start namespace with an error.
	cmd = &Command{
		MethodName: "admin_startNamespace",
		Args:       []string{"mock_ns"},
	}
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully start the namespace <global> " +
		"as we start namespace in another go-routine, we always send 'start namespace cmd' successfully.")
}

func TestStopNamespace(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	mockNSMgr.On("StopNamespace", "global").Return(nil)
	mockNSMgr.On("StopNamespace", "mock_ns").Return(mockError)

	// mock: stop namespace with invalid arg numbers.
	cmd := &Command{
		MethodName: "admin_stopNamespace",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	invalidParamsErr := &common.InvalidParamsError{}
	ast.Equal(invalidParamsErr.Code(), result.Error.Code(),
		"We should fail to stop the namespace as we provide 0 args.")

	// mock: stop namespace with no error return.
	cmd = &Command{
		MethodName: "admin_stopNamespace",
		Args:       []string{"global"},
	}
	service = getService(cmd.MethodName, t)
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully stop the namespace <global>.")

	// mock: stop namespace with an error.
	cmd = &Command{
		MethodName: "admin_stopNamespace",
		Args:       []string{"mock_ns"},
	}
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully start the namespace <global> " +
		"as we start namespace in another go-routine, we always send 'stop namespace cmd' successfully.")
}

func TestRestartNamespace(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	mockNSMgr.On("RestartNamespace", "global").Return(nil)
	mockNSMgr.On("RestartNamespace", "mock_ns").Return(mockError)

	// mock: restart namespace with invalid arg numbers.
	cmd := &Command{
		MethodName: "admin_restartNamespace",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	invalidParamsErr := &common.InvalidParamsError{}
	ast.Equal(invalidParamsErr.Code(), result.Error.Code(),
		"We should fail to restart the namespace as we provide 0 args.")

	// mock: restart namespace with no error return.
	cmd = &Command{
		MethodName: "admin_restartNamespace",
		Args:       []string{"global"},
	}
	service = getService(cmd.MethodName, t)
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully restart the namespace <global>.")

	// mock: restart namespace with an error.
	cmd = &Command{
		MethodName: "admin_restartNamespace",
		Args:       []string{"mock_ns"},
	}
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully start the namespace <global> " +
		"as we start namespace in another go-routine, we always send 'restart namespace cmd' successfully.")
}

func TestRegisterNamespace(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	mockNSMgr.On("Register", "global").Return(nil)
	mockNSMgr.On("Register", "mock_ns").Return(mockError)

	// mock: register namespace with invalid arg numbers.
	cmd := &Command{
		MethodName: "admin_registerNamespace",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	invalidParamsErr := &common.InvalidParamsError{}
	ast.Equal(invalidParamsErr.Code(), result.Error.Code(),
		"We should fail to register the namespace as we provide 0 args.")

	// mock: register namespace with no error return.
	cmd = &Command{
		MethodName: "admin_registerNamespace",
		Args:       []string{"global"},
	}
	service = getService(cmd.MethodName, t)
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully register the namespace <global>.")

	// mock: register namespace with an error.
	cmd = &Command{
		MethodName: "admin_registerNamespace",
		Args:       []string{"mock_ns"},
	}
	result = admin.CmdExecutor[service](cmd)
	ast.Equal(result.Error.Error(), mockError.Error(), "We should fail to register the " +
		"namespace <mock_ns> because of some errors.")
}

func TestDeRegisterNamespace(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	mockNSMgr.On("DeRegister", "global").Return(nil)
	mockNSMgr.On("DeRegister", "mock_ns").Return(mockError)

	// mock: deregister namespace with invalid arg numbers.
	cmd := &Command{
		MethodName: "admin_deregisterNamespace",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	invalidParamsErr := &common.InvalidParamsError{}
	ast.Equal(invalidParamsErr.Code(), result.Error.Code(),
		"We should fail to deregister the namespace as we provide 0 args.")

	// mock: deregister namespace with no error return.
	cmd = &Command{
		MethodName: "admin_deregisterNamespace",
		Args:       []string{"global"},
	}
	service = getService(cmd.MethodName, t)
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully deregister the namespace <global>.")

	// mock: deregister namespace with an error.
	cmd = &Command{
		MethodName: "admin_deregisterNamespace",
		Args:       []string{"mock_ns"},
	}
	result = admin.CmdExecutor[service](cmd)
	ast.Equal(result.Error.Error(), mockError.Error(), "We should fail to deregister the " +
		"namespace <mock_ns> because of some errors.")
}

func TestListNamespace(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	mockList := []string{"global", "mock_ns"}
	mockNSMgr.On("List").Return(mockList)

	// mock: list namespace.
	cmd := &Command{
		MethodName: "admin_listNamespaces",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.Equal(mockList, result.Result,
		"We should successfully list the namespace <global,mock_ns>.")
}

func TestGetLevel(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	common.InitRawHyperLogger("global")
	common.SetLogLevel("global", "consensus", "debug")

	// mock: get log level with invalid arg numbers.
	cmd := &Command{
		MethodName: "admin_getLevel",
		Args:       []string{"global"},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	invalidParamsErr := &common.InvalidParamsError{}
	ast.Equal(invalidParamsErr.Code(), result.Error.Code(),
		"We should fail to get the log level as we only provide 1 args.")

	// mock: get log level with no error return.
	cmd = &Command{
		MethodName: "admin_getLevel",
		Args:       []string{"global", "consensus"},
	}
	service = getService(cmd.MethodName, t)
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully get the log level of " +
		"<consensus> module in namespace <global>.")

	// mock: get log level with an error.
	cmd = &Command{
		MethodName: "admin_getLevel",
		Args:       []string{"mock_ns", "consensus"},
	}
	result = admin.CmdExecutor[service](cmd)
	ast.Contains(result.Error.Error(), "not init yet!")
}

func TestSetLevel(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	common.InitRawHyperLogger("global")
	common.SetLogLevel("global", "consensus", "debug")

	// mock: set log level with invalid arg numbers.
	cmd := &Command{
		MethodName: "admin_setLevel",
		Args:       []string{"global"},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	invalidParamsErr := &common.InvalidParamsError{}
	ast.Equal(invalidParamsErr.Code(), result.Error.Code(),
		"We should fail to set the log level as we only provide 1 args.")

	// mock: set log level with no error return.
	cmd = &Command{
		MethodName: "admin_setLevel",
		Args:       []string{"global", "consensus", "ERROR"},
	}
	service = getService(cmd.MethodName, t)
	result = admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully set the log level of " +
		"<consensus> module in namespace <global>.")
	level, _ := common.GetLogLevel("global", "consensus")
	ast.Equal("ERROR", level, "They should be equal.")

	// mock: get log level with an error.
	cmd = &Command{
		MethodName: "admin_setLevel",
		Args:       []string{"mock_ns", "consensus", "error"},
	}
	result = admin.CmdExecutor[service](cmd)
	ast.Contains(result.Error.Error(), "not init yet!")
}

func TestStartJvmServer(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	// mock: start jvm server with no error return, limit only 1 return time
	// as we need to call StartJvm() interface below two times, if we don't
	// limit time here, next mockNSMgr.On(...).Return(...) will be replaced
	// by this one.
	mockNSMgr.On("StartJvm").Return(nil).Times(1)
	cmd := &Command{
		MethodName: "admin_startJvmServer",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully start the jvm manager.")

	// mock: start jvm server with an error, do not need to limit return times.
	mockNSMgr.On("StartJvm").Return(mockError)
	result = admin.CmdExecutor[service](cmd)
	ast.Equal(result.Error.Error(), mockError.Error(), "We should fail to " +
		"start the jvm manager because of some errors.")

}

func TestRestartJvmServer(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	// mock: restart jvm server with no error return, limit only 1 return time
	// as we need to call RestartJvm() interface below two times, if we don't
	// limit time here, next mockNSMgr.On(...).Return(...) will be replaced
	// by this one.
	mockNSMgr.On("RestartJvm").Return(nil).Times(1)
	cmd := &Command{
		MethodName: "admin_restartJvmServer",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully restart the jvm manager.")

	// mock: restart jvm server with an error, do not need to limit return times.
	mockNSMgr.On("RestartJvm").Return(mockError)
	result = admin.CmdExecutor[service](cmd)
	ast.Equal(result.Error.Error(), mockError.Error(), "We should fail to " +
		"restart the jvm manager because of some errors.")

}

func TestStopJvmServer(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()
	mockNSMgr := &mocks.MockNSMgr{}
	admin.nsMgr = mockNSMgr

	// mock: stop jvm server with no error return, limit only 1 return time
	// as we need to call StopJvm() interface below two times, if we don't
	// limit time here, next mockNSMgr.On(...).Return(...) will be replaced
	// by this one.
	mockNSMgr.On("StopJvm").Return(nil).Times(1)
	cmd := &Command{
		MethodName: "admin_stopJvmServer",
		Args:       []string{},
	}
	service := getService(cmd.MethodName, t)
	result := admin.CmdExecutor[service](cmd)
	ast.True(result.Ok, "We should successfully stop the jvm manager.")

	// mock: stop jvm server with an error, do not need to limit return times.
	mockNSMgr.On("StopJvm").Return(mockError)
	result = admin.CmdExecutor[service](cmd)
	ast.Equal(result.Error.Error(), mockError.Error(), "We should fail to " +
		"stop the jvm manager because of some errors.")

}

func TestGrantPermission(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()

	// mock: grant permission with invalid arg numbers.
	cmd := &Command{
		MethodName: "grantPermission",
		Args:       []string{"root", "contract"},
	}
	result := admin.grantPermission(cmd)
	invalidParamsError := &common.InvalidParamsError{}
	ast.Equal(invalidParamsError.Code(),result.Error.Code(),
		"Grant failed as we only provide 2 args.")

	// mock: grant permission to a non-exist user.
	cmd = &Command{
		MethodName: "grantPermission",
		Args:       []string{"hpc", "contract", "all"},
	}
	result = admin.grantPermission(cmd)
	callbackError := &common.CallbackError{}
	ast.Equal(callbackError.Code(), result.Error.Code(),
		"Grant failed as user hpc is not existed.")

	// mock: grant permission with an invalid permission.
	cmd = &Command{
		MethodName: "grantPermission",
		Args:       []string{"root", "contract", "mock_permission"},
	}
	result = admin.grantPermission(cmd)
	ast.Contains(result.Result.(string), "invalid permissions",
		"Grant failed as we provide some invalid permissions.")

	// mock: grant permission with correct permission to an existed user.
	cmd = &Command{
		MethodName: "grantPermission",
		Args:       []string{"root", "contract", "all"},
	}
	result = admin.grantPermission(cmd)
	ast.True(result.Ok, "We should grant successfully.")

}

func TestRevokePermission(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()

	// mock: revoke permission with invalid arg numbers.
	cmd := &Command{
		MethodName: "revokePermission",
		Args:       []string{"root", "contract"},
	}
	result := admin.revokePermission(cmd)
	invalidParamsError := &common.InvalidParamsError{}
	ast.Equal(invalidParamsError.Code(),result.Error.Code(),
		"Revoke failed as we only provide 2 args.")

	// mock: revoke permission to a non-exist user.
	cmd = &Command{
		MethodName: "revokePermission",
		Args:       []string{"hpc", "contract", "all"},
	}
	result = admin.revokePermission(cmd)
	callbackError := &common.CallbackError{}
	ast.Equal(callbackError.Code(), result.Error.Code(),
		"Revoke failed as user hpc is not existed.")

	// mock: revoke permission with an invalid permission.
	cmd = &Command{
		MethodName: "revokePermission",
		Args:       []string{"root", "contract", "mock_permission"},
	}
	result = admin.revokePermission(cmd)
	ast.Contains(result.Result.(string), "invalid permissions",
		"Revoke failed as we provide some invalid permissions.")

	// mock: revoke permission with correct permission to an existed user.
	cmd = &Command{
		MethodName: "revokePermission",
		Args:       []string{"root", "contract", "all"},
	}
	result = admin.revokePermission(cmd)
	ast.True(result.Ok, "We should revoke successfully.")

}

func TestListPermission(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()

	// mock: list permission with invalid arg numbers.
	cmd := &Command{
		MethodName: "listPermission",
		Args:       []string{"root", "contract"},
	}
	result := admin.listPermission(cmd)
	invalidParamsError := &common.InvalidParamsError{}
	ast.Equal(invalidParamsError.Code(),result.Error.Code(),
		"Revoke failed as we provide 2 args.")

	// mock: list permission to a non-exist user.
	cmd = &Command{
		MethodName: "listPermission",
		Args:       []string{"hpc"},
	}
	result = admin.listPermission(cmd)
	callbackError := &common.CallbackError{}
	ast.Equal(callbackError.Code(), result.Error.Code(),
		"List permission failed as user hpc is not existed.")

	// mock: list permission with an existed user.
	cmd = &Command{
		MethodName: "revokePermission",
		Args:       []string{"root"},
	}
	result = admin.listPermission(cmd)
	ast.True(result.Ok, "We should list successfully.")

}

func TestCreateUser(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()

	// mock: create user with invalid arg numbers.
	cmd := &Command{
		MethodName: "createUser",
		Args:       []string{"hpc", "pwd"},
	}
	result := admin.createUser(cmd)
	invalidParamsError := &common.InvalidParamsError{}
	ast.Equal(invalidParamsError.Code(),result.Error.Code(),
		"Create failed as we provide 2 args.")

	// mock: create user with an exist user.
	cmd = &Command{
		MethodName: "createUser",
		Args:       []string{"root", "hyperchain", "ALL"},
	}
	result = admin.createUser(cmd)
	callbackError := &common.CallbackError{}
	ast.Equal(callbackError.Code(), result.Error.Code(),
		"Create failed as user root is existed.")

	// mock: create user with an unrecognized group.
	cmd = &Command{
		MethodName: "createUser",
		Args:       []string{"hpc", "pwd", "mock_group"},
	}
	result = admin.createUser(cmd)
	ast.Equal(callbackError.Code(), result.Error.Code(),
		"Create failed as mock_group is a unrecognized group.")

	// mock: create a non-existed user with a correct group.
	cmd = &Command{
		MethodName: "createUser",
		Args:       []string{"hpc", "pwd", "contract"},
	}
	result = admin.createUser(cmd)
	ast.True(result.Ok, "We should create successfully.")

}

func TestAlterUser(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()

	// mock: alter user with invalid arg numbers.
	cmd := &Command{
		MethodName: "alterUser",
		Args:       []string{"root"},
	}
	result := admin.alterUser(cmd)
	invalidParamsError := &common.InvalidParamsError{}
	ast.Equal(invalidParamsError.Code(),result.Error.Code(),
		"Alter failed as we provide 1 args.")

	// mock: alter user with a non-exist user.
	cmd = &Command{
		MethodName: "alterUser",
		Args:       []string{"hpc", "pwd"},
	}
	result = admin.alterUser(cmd)
	callbackError := &common.CallbackError{}
	ast.Equal(callbackError.Code(), result.Error.Code(),
		"Alter failed as user hpc is not existed.")

	// mock: alter an existed user with a new password.
	cmd = &Command{
		MethodName: "alterUser",
		Args:       []string{"root", "pwd"},
	}
	result = admin.alterUser(cmd)
	ast.True(result.Ok, "We should alter successfully.")

}

func TestDelUser(t *testing.T) {
	ast := assert.New(t)
	admin := getAdmin()

	// mock: delete user with invalid arg numbers.
	cmd := &Command{
		MethodName: "delUser",
		Args:       []string{"root", "hyperchain"},
	}
	result := admin.delUser(cmd)
	invalidParamsError := &common.InvalidParamsError{}
	ast.Equal(invalidParamsError.Code(),result.Error.Code(),
		"Delete failed as we provide 2 args.")

	// mock: delete user with a non-exist user.
	cmd = &Command{
		MethodName: "delUser",
		Args:       []string{"hpc"},
	}
	result = admin.delUser(cmd)
	callbackError := &common.CallbackError{}
	ast.Equal(callbackError.Code(), result.Error.Code(),
		"Delete failed as user root is existed.")

	// mock: delete an existed user.
	cmd = &Command{
		MethodName: "delUser",
		Args:       []string{"root"},
	}
	result = admin.delUser(cmd)
	ast.True(result.Ok, "We should delete successfully.")

}

func getAdmin() *Administrator{
	admin := &Administrator{
		check:       true,
		expiration:  1 * time.Second,
		CmdExecutor: make(map[string]func(command *Command) *CommandResult),
		valid_user:  make(map[string]string),
		user_scope:  make(map[string]permissionSet),
		user_opTime: make(map[string]int64),
	}
	admin.init()
	return admin
}

func getService(name string, t *testing.T) string {
	ss := strings.Split(name, common.ServiceMethodSeparator)
	if len(ss) != 2 {
		t.Errorf("Invalid Method name(%s) which should be splited to" +
			" 2 components by separator '%s'", name, common.ServiceMethodSeparator)
	}
	return ss[1]
}
