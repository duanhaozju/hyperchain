package jsonrpc

import "strings"

const (
	expiration int64 = 300
	beforetime int64 = 300
	pri_key string   = "../../api/jsonrpc/core/key/sample_key"
	pub_key string   = "../../api/jsonrpc/core/key/sample_key.pub"
)

const (
	LoggedIn   = "Log in Succeefully"
)

const (
	// server cmd
	admin_stopServer          int = iota
	admin_restartServer

	// namespace cmd
	admin_startNsMgr
	admin_stopNsMgr
	admin_registerNamespace
	admin_deregisterNamespace
	admin_startNamespace
	admin_stopNamespace
	admin_restartNamespace
	admin_listNamespaces

	// http cmd
	admin_startHttpServer
	admin_stopHttpServer
	admin_restartHttpServer

	// log cmd
	admin_setLevel
	admin_getLevel

	// user cmd
	admin_createUser
	admin_alterUser
	admin_dropUser

	// permission cmd
	admin_grantPermission
	admin_revokePermission
	admin_listPermission

	// contract cmd
	contract_deployContract
	contract_invokeContract

	// node cmd
	node_delNode

	MAXNUM
)

var defaultScope = []int{admin_getLevel, admin_listNamespaces, contract_deployContract,
	contract_invokeContract, admin_listPermission}

// convertToScope converts method name to corresponding scope
func convertToScope(method string) int {
	method = toUpper(method)
	switch method {
	case toUpper("admin_stopServer"):
		return admin_stopServer
	case toUpper("admin_restartServer"):
		return admin_restartServer
	case toUpper("admin_startNsMgr"):
		return admin_startNsMgr
	case toUpper("admin_stopNsMgr"):
		return admin_stopNsMgr
	case toUpper("admin_registerNamespace"):
		return admin_registerNamespace
	case toUpper("admin_deregisterNamespace"):
		return admin_deregisterNamespace
	case toUpper("admin_startNamespace"):
		return admin_startNamespace
	case toUpper("admin_stopNamespace"):
		return admin_stopNamespace
	case toUpper("admin_restartNamespace"):
		return admin_restartNamespace
	case toUpper("admin_listNamespaces"):
		return admin_listNamespaces
	case toUpper("admin_startHttpServer"):
		return admin_startHttpServer
	case toUpper("admin_stopHttpServer"):
		return admin_stopHttpServer
	case toUpper("admin_restartHttpServer"):
		return admin_restartHttpServer
	case toUpper("admin_setLevel"):
		return admin_setLevel
	case toUpper("admin_getLevel"):
		return admin_getLevel
	case toUpper("admin_createUser"):
		return admin_createUser
	case toUpper("admin_alterUser"):
		return admin_alterUser
	case toUpper("admin_delUser"):
		return admin_dropUser
	case toUpper("admin_grantPermission"):
		return admin_grantPermission
	case toUpper("admin_revokePermission"):
		return admin_revokePermission
	case toUpper("admin_listPermission"):
		return admin_listPermission
	case toUpper("contract_deployContract"):
		return contract_deployContract
	case toUpper("contract_invokeContract"):
		return contract_invokeContract
	case toUpper("node_delNode"):
		return node_delNode

	default:
		return -1
	}
}

// convertToMethod converts scope to method
func convertToMethod(scope int) string {
	switch scope {
	case admin_stopServer:
		return "admin_stopServer"
	case admin_restartServer:
		return "admin_restartServer"
	case admin_startNsMgr:
		return "admin_startNsMgr"
	case admin_stopNsMgr:
		return "admin_stopNsMgr"
	case admin_registerNamespace:
		return "admin_registerNamespace"
	case admin_deregisterNamespace:
		return "admin_deregisterNamespace"
	case admin_startNamespace:
		return "admin_startNamespace"
	case admin_stopNamespace:
		return "admin_stopNamespace"
	case admin_restartNamespace:
		return "admin_restartNamespace"
	case admin_listNamespaces:
		return "admin_listNamespaces"
	case admin_startHttpServer:
		return "admin_startHttpServer"
	case admin_stopHttpServer:
		return "admin_stopHttpServer"
	case admin_restartHttpServer:
		return "admin_restartHttpServer"
	case admin_setLevel:
		return "admin_setLevel"
	case admin_getLevel:
		return "admin_getLevel"
	case admin_createUser:
		return "admin_createUser"
	case admin_alterUser:
		return "admin_alterUser"
	case admin_dropUser:
		return "admin_dropUser"
	case admin_grantPermission:
		return "admin_grantPermission"
	case admin_revokePermission:
		return "admin_revokePermission"
	case admin_listPermission:
		return "admin_listPermission"
	case contract_deployContract:
		return "contract_deployContract"
	case contract_invokeContract:
		return "contract_invokeContract"
	case node_delNode:
		return "node_delNode"

	default:
		return ""
	}
}

func ReadablePermission(scope float64) string {
	permission := int(scope)
	switch permission {
	case admin_stopServer:
		return "server::stop"
	case admin_restartServer:
		return "server:restart"
	case admin_startNsMgr:
		return "namespace::startNsMgr"
	case admin_stopNsMgr:
		return "namespace::stopNsMgr"
	case admin_registerNamespace:
		return "namespace::register [ns-name]"
	case admin_deregisterNamespace:
		return "namespace::deregister [ns-name]"
	case admin_startNamespace:
		return "namespace::start [ns-name]"
	case admin_stopNamespace:
		return "namespace::stop [ns-name]"
	case admin_restartNamespace:
		return "namespace::restart [ns-name]"
	case admin_listNamespaces:
		return "namespace::list"
	case admin_startHttpServer:
		return "http::start"
	case admin_stopHttpServer:
		return "http::stop"
	case admin_restartHttpServer:
		return "http::restart"
	case admin_setLevel:
		return "log::setLevel [ns-name] [module] [logLevel]"
	case admin_getLevel:
		return "log::getLevel [ns-name] [module]"
	case admin_createUser:
		return "auth::create [username] [password]"
	case admin_alterUser:
		return "auth::alter [username] [password]"
	case admin_dropUser:
		return "auth::drop [username]"
	case admin_grantPermission:
		return "auth::grant [username] [permissions...]"
	case admin_revokePermission:
		return "auth::revoke [username] [permissions...]"
	case admin_listPermission:
		return "auth::list [username]"
	case contract_deployContract:
		return "contract::deploy [params...]"
	case contract_invokeContract:
		return "contract::invoke [params...]"
	case node_delNode:
		return "node::delete [params...]"

	default:
		return "Undified permission!"
	}
}

func rootScopes() permissionSet {
	pset := make(permissionSet)
	for i := 0; i< MAXNUM; i++ {
		pset[i] = true
	}
	return pset
}

func defaultScopes() permissionSet {
	pset := make(permissionSet)
	for _, scope := range defaultScope {
		pset[scope] = true
	}
	return pset
}

func toUpper(origin string) string {
	return strings.ToUpper(origin)
}