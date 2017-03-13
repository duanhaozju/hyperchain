//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package namespace

import "hyperchain/api/jsonrpc/core"

//handleJsonRequest handle JsonRequest under current namespace.
func (*namespaceImpl) handleJsonRequest(request *jsonrpc.JSONRequest) *jsonrpc.JSONResponse {
	//TODO: implement JsonRequest process logic as which handled in api/jsonrpc/core
	//1. check the invoke method params
	//2. invoke specific method
	//3. construct response
	return nil
}
