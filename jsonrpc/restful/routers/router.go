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
			beego.NSRouter("/:txHash", &controllers.TransactionsController{}, "get:GetTransactionByHash"),
			beego.NSRouter("/query", &controllers.TransactionsController{}, "get:GetTransactionByBlockNumberOrBlockHash"),
			//beego.NSRouter("/{blkHash}/{index}", &controllers.TransactionsController{}, "get:GetTransactionByBlockHashAndIndex"),
			beego.NSRouter("/:txHash/receipt", &controllers.TransactionsController{}, "get:GetTransactionReceipt"),
			beego.NSRouter("/count", &controllers.TransactionsController{}, "get:GetTransactionsCount"),
			beego.NSRouter("/sign-hash", &controllers.TransactionsController{}, "post:GetSignHash"),
		),

		beego.NSNamespace("/blocks",
			//beego.NSRouter("/:blkNum/:index", &controllers.BlocksController{}, "get:GetTransactionByBlockNumberAndIndex"),
			//beego.NSRouter("/:blkHash/:index", &controllers.BlocksController{}, "get:GetTransactionByBlockHashAndIndex"),
			beego.NSRouter("/list", &controllers.BlocksController{}, "get:GetBlocks"),
			beego.NSRouter("/query", &controllers.BlocksController{}, "get:GetBlockByHashOrNum"),
			beego.NSRouter("/latest", &controllers.BlocksController{}, "get:GetLatestBlock"),
		),

		beego.NSNamespace("/contracts",
			beego.NSRouter("/compile", &controllers.ContractsController{}, "post:CompileContract"),
			beego.NSRouter("/deploy", &controllers.ContractsController{}, "post:DeployContract"),
			beego.NSRouter("/invoke", &controllers.ContractsController{}, "post:InvokeContract"),
			beego.NSRouter("/query", &controllers.ContractsController{}, "get:GetCode"),
		),
	)
	beego.AddNamespace(ns)
}