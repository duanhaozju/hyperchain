//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package namespace

import (
	"hyperchain/common"
)

//handleJsonRequest handle JsonRequest under current namespace.
func (ni *NamespaceImpl) handleJsonRequest(request *common.RPCRequest) *common.RPCResponse {

	return ni.rpcProcesser.ProcessRequest(request)

	//TODO: implement JsonRequest process logic as which handled in api/jsonrpc/core
	//1. check the invoke method params
	//2. invoke specific method
	//3. construct response
}
