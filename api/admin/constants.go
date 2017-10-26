//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package jsonrpc

import (
	"strings"
)

// public/private key locations.
var (
	pri_key string = "./hypercli/keyconfigs/key/key"
	pub_key string = "./hypercli/keyconfigs/key/key.pub"
)

const (
	LoggedIn = "Log in Succeefully"
)

const (
	// server cmd
	admin_stopServer int = iota
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

	// jvm cmd
	admin_startJvmServer
	admin_stopJvmServer
	admin_restartJvmServer

	// contract cmd
	contract_deployContract
	contract_invokeContract
	contract_maintainContract

	// node cmd
	node_getNodes
	node_getNodeHash
	node_delNode

	// tx cmd
	tx_getTransactionReceipt

	MAXNUM
)

var defaultGroup = []int{admin_getLevel, admin_listNamespaces, admin_listPermission, admin_alterUser}

var namespaceGroup = []int{admin_startNsMgr, admin_stopNsMgr, admin_registerNamespace, admin_deregisterNamespace,
	admin_startNamespace, admin_stopNamespace, admin_restartNamespace, admin_listNamespaces}

var httpGroup = []int{admin_startHttpServer, admin_stopHttpServer, admin_restartHttpServer}

var logGroup = []int{admin_setLevel, admin_getLevel}

var authGroup = []int{admin_createUser, admin_alterUser, admin_dropUser, admin_grantPermission,
	admin_revokePermission, admin_listPermission}

var jvmGroup = []int{admin_startJvmServer, admin_stopJvmServer, admin_restartJvmServer}

var contractGroup = []int{contract_deployContract, contract_invokeContract, contract_maintainContract}

var nodeGroup = []int{node_getNodes, node_getNodeHash, node_delNode}

var txGroup = []int{tx_getTransactionReceipt}

// convertToScope converts method name to corresponding scope.
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
	case toUpper("admin_startJvmServer"):
		return admin_startJvmServer
	case toUpper("admin_stopJvmServer"):
		return admin_stopJvmServer
	case toUpper("admin_restartJvmServer"):
		return admin_restartJvmServer
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
	case toUpper("contract_maintainContract"):
		return contract_maintainContract
	case toUpper("node_getNodes"):
		return node_getNodes
	case toUpper("node_getNodeHash"):
		return node_getNodeHash
	case toUpper("node_delNode"):
		return node_delNode
	case toUpper("tx_getTransactionReceipt"):
		return tx_getTransactionReceipt

	default:
		return -1
	}
}

// convertToIntegers converts scope to corresponding integer.
func convertToIntegers(scope string) []int {
	switch scope {
	case "server::stop":
		return []int{admin_stopServer}
	case "server:restart":
		return []int{admin_restartServer}
	case "server::all":
		return []int{admin_stopServer, admin_restartServer}
	case "namespace::startNsMgr":
		return []int{admin_startNsMgr}
	case "namespace::stopNsMgr":
		return []int{admin_stopNsMgr}
	case "namespace::register":
		return []int{admin_registerNamespace}
	case "namespace::deregister":
		return []int{admin_deregisterNamespace}
	case "namespace::start":
		return []int{admin_startNamespace}
	case "namespace::stop":
		return []int{admin_stopNamespace}
	case "namespace::restart":
		return []int{admin_restartNamespace}
	case "namespace::list":
		return []int{admin_listNamespaces}
	case "namespace::all":
		return namespaceGroup
	case "http::start":
		return []int{admin_startHttpServer}
	case "http::stop":
		return []int{admin_stopHttpServer}
	case "http::restart":
		return []int{admin_restartHttpServer}
	case "http::all":
		return httpGroup
	case "jvm::start":
		return []int{admin_startJvmServer}
	case "jvm::stop":
		return []int{admin_stopJvmServer}
	case "jvm::restart":
		return []int{admin_restartJvmServer}
	case "jvm::all":
		return jvmGroup
	case "log::setLevel":
		return []int{admin_setLevel}
	case "log::getLevel":
		return []int{admin_getLevel}
	case "log::all":
		return logGroup
	case "auth::create":
		return []int{admin_createUser}
	case "auth::alter":
		return []int{admin_alterUser}
	case "auth::drop":
		return []int{admin_dropUser}
	case "auth::grant":
		return []int{admin_grantPermission}
	case "auth::revoke":
		return []int{admin_revokePermission}
	case "auth::list":
		return []int{admin_listPermission}
	case "auth::all":
		return authGroup
	case "contract::deploy":
		return []int{contract_deployContract}
	case "contract::invoke":
		return []int{contract_invokeContract}
	case "contract::maintain":
		return []int{contract_maintainContract}
	case "contract::all":
		return contractGroup
	case "node::getNodes":
		return []int{node_getNodes}
	case "node::getNodeHash":
		return []int{node_getNodeHash}
	case "node::delete":
		return []int{node_delNode}
	case "node::all":
		return nodeGroup
	case "tx::getTransactionReceipt":
		return []int{tx_getTransactionReceipt}
	case "tx::all":
		return txGroup

	default:
		return nil
	}
}

// ReadablePermission returns the user readable permission
// of the given scope.
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
	case admin_startJvmServer:
		return "jvm::start"
	case admin_stopJvmServer:
		return "jvm::stop"
	case admin_restartJvmServer:
		return "jvm::restart"
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
	case contract_maintainContract:
		return "contract:: [update/frozen/unfrozen/destroy] [params...]"
	case node_getNodes:
		return "node::getNodes"
	case node_getNodeHash:
		return "node::getNodeHash"
	case node_delNode:
		return "node::delete [params...]"
	case tx_getTransactionReceipt:
		return "tx::getTransactionReceipt"

	default:
		return "Undefined permission!"
	}
}

// getGroupPermission returns the group's permissions
// of the given group name.
func getGroupPermission(group string) permissionSet {
	group = toUpper(group)
	switch group {
	case toUpper("root"):
		return rootScopes()
	case toUpper("default"):
		return defaultScopes()
	case toUpper("namespace"):
		return namespaceScopes()
	case toUpper("http"):
		return httpScopes()
	case toUpper("log"):
		return logScopes()
	case toUpper("auth"):
		return authScopes()
	case toUpper("contract"):
		return contractScopes()
	case toUpper("node"):
		return nodeScopes()
	case toUpper("tx"):
		return txScopes()

	default:
		return nil
	}
}

// rootScopes returns root user's permissions, which is
// all the permissions.
func rootScopes() permissionSet {
	pset := make(permissionSet)
	for i := 0; i < MAXNUM; i++ {
		pset[i] = true
	}
	return pset
}

// defaultScopes returns the default user permissions.
func defaultScopes() permissionSet {
	return getScope(defaultGroup)
}

// namespaceScopes returns the namespace group permissions.
func namespaceScopes() permissionSet {
	return getScope(namespaceGroup)
}

// httpScopes returns the http group permissions.
func httpScopes() permissionSet {
	return getScope(httpGroup)
}

// logScopes returns the log group permissions.
func logScopes() permissionSet {
	return getScope(logGroup)
}

// authScopes returns the auth group permissions.
func authScopes() permissionSet {
	return getScope(authGroup)
}

// contractScopes returns the contract group permissions.
func contractScopes() permissionSet {
	return getScope(contractGroup)
}

// nodeScopes returns the node group permissions.
func nodeScopes() permissionSet {
	return getScope(nodeGroup)
}

// txScopes returns the transaction group permissions.
func txScopes() permissionSet {
	return getScope(txGroup)
}

// toUpper converts the given string to all upper letters.
func toUpper(origin string) string {
	return strings.ToUpper(origin)
}

// getScope returns the scopes, all users have the default scopes.
func getScope(scope []int) permissionSet {
	pset := make(permissionSet)
	for _, pms := range scope {
		pset[pms] = true
	}
	// every one has a default scope
	for _, pms := range defaultGroup {
		pset[pms] = true
	}
	return pset
}
