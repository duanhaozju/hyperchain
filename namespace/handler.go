//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package namespace

import (
	"hyperchain/common"
)

// handleJsonRequest handles JsonRequest under current namespace.
func (ns *namespaceImpl) handleJsonRequest(request *common.RPCRequest) *common.RPCResponse {
	return ns.rpc.ProcessRequest(request)
}
