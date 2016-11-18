package routers

import (
	"net/http"
	"hyperchain/jsonrpc/restful/controllers"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"GetTransactions",
		"GET",
		"/v1/transactions",
		controllers.GetTransactions,
	},
	Route{
		"GetTransactionByHash",
		"GET",
		"/v1/transactions/{txHash}",
		controllers.GetTransactionByHash,
	},
	Route{
		"GetReceiptByHash",
		"GET",
		"/v1/transactions/{txHash}/receipt",
		controllers.GetReceiptByHash,
	},
	Route{
		"SendTransaction",
		"POST",
		"/v1/transactions/actions/send",
		controllers.SendTransaction,
	},
	Route{
		"GetSignHash",
		"POST",
		"/v1/transactions/sign-hash",
		controllers.GetSignHash,
	},
	Route{
		"CompileContract",
		"POST",
		"/v1/contracts/actions/compile",
		controllers.CompileContract,
	},
	Route{
		"DeployContract",
		"POST",
		"/v1/contracts/actions/deploy",
		controllers.DeployContract,
	},
	Route{
		"InvokeContract",
		"POST",
		"/v1/contracts/actions/invoke",
		controllers.InvokeContract,
	},
	Route{
		"GetLatestBlock",
		"GET",
		"/v1/blocks/latest",
		controllers.GetLatestBlock,
	},
	Route{
		"GetBlockByNumberOrHash",
		"GET",
		"/v1/blocks/{block_num_or_hash}",
		controllers.GetBlockByNumberOrHash,
	},
	//Route{
	//	"GetBlockByHash",
	//	"GET",
	//	"/v1/blocks/{block_hash}",
	//	controller.GetBlockByHash,
	//},
	Route{
		"GetBlocks",
		"POST",
		"/v1/blocks",
		controllers.GetBlocks,
	},
}