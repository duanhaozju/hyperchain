//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package namespace

import (
	"hyperchain/common"
)

//handleJsonRequest handle JsonRequest under current namespace.
func (ni *namespaceImpl) handleJsonRequest(request *common.RPCRequest) *common.RPCResponse {
	return ni.rpcProcesser.ProcessRequest(request)
}
