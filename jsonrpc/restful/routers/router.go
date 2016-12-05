package routers

import (
	"github.com/astaxie/beego"
	"hyperchain/jsonrpc/restful/controllers"
)

func init() {
	ns := beego.NewNamespace("/v1",

		beego.NSNamespace("/transactions",
			//beego.NSInclude(
			//	&controllers.TransactionsController{},
			//),
			beego.NSRouter("/send", &controllers.TransactionsController{}, "post:SendTransaction"),
			beego.NSRouter("/list", &controllers.TransactionsController{}, "get:GetTransactions"),
			beego.NSRouter("/:transactionHash", &controllers.TransactionsController{}, "get:GetTransactionByHash"),
			beego.NSRouter("/query", &controllers.TransactionsController{}, "get:GetTransactionByBlockNumberOrBlockHash"),
			beego.NSRouter("/:transactionHash/receipt", &controllers.TransactionsController{}, "get:GetTransactionReceipt"),
			beego.NSRouter("/get-hash-for-sign", &controllers.TransactionsController{}, "post:GetSignHash"),
			beego.NSRouter("/average-time", &controllers.TransactionsController{}, "get:GetTxAvgTimeByBlockNumber"),
			beego.NSRouter("/count", &controllers.TransactionsController{}, "get:GetTransactionsCount"),
		),

		beego.NSNamespace("/blocks",
			beego.NSRouter("/list", &controllers.BlocksController{}, "get:GetBlocks"),
			beego.NSRouter("/query", &controllers.BlocksController{}, "get:GetBlockByHashOrNum"),
			beego.NSRouter("/:blockHash/transactions/count", &controllers.BlocksController{}, "get:GetBlockTransactionCountByHash"),
			beego.NSRouter("/latest", &controllers.BlocksController{}, "get:GetLatestBlock"),
		),

		beego.NSNamespace("/contracts",
			beego.NSRouter("/compile", &controllers.ContractsController{}, "post:CompileContract"),
			beego.NSRouter("/deploy", &controllers.ContractsController{}, "post:DeployContract"),
			beego.NSRouter("/invoke", &controllers.ContractsController{}, "post:InvokeContract"),
			beego.NSRouter("/query", &controllers.ContractsController{}, "get:GetCode"),
		),

		beego.NSNamespace("/accounts",
			beego.NSRouter("/:address/contracts/count", &controllers.AccountsController{}, "get:GetContractCountByAddr"),
		),

		beego.NSNamespace("/nodes",
			beego.NSRouter("/list", &controllers.NodesController{}, "get:GetNodes"),
		),

	)
	beego.AddNamespace(ns)
}