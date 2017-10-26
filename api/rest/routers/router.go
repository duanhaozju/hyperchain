package routers

import (
	"github.com/astaxie/beego"
	"github.com/hyperchain/hyperchain/api/rest/controllers"
)

func NewRouter() {

	ns := beego.NewNamespace("/v1",

		beego.NSNamespace("/transactions",
			beego.NSRouter("/send", &controllers.TransactionsController{}, "post:SendTransaction"),
			beego.NSRouter("/list", &controllers.TransactionsController{}, "get:GetTransactions"),
			beego.NSRouter("/:transactionHash", &controllers.TransactionsController{}, "get:GetTransactionByHash"),
			beego.NSRouter("/query", &controllers.TransactionsController{}, "get:GetTransactionByBlockNumberOrBlockHashOrTime"),
			beego.NSRouter("/:transactionHash/receipt", &controllers.TransactionsController{}, "get:GetTransactionReceipt"),
			beego.NSRouter("/signature-hash", &controllers.TransactionsController{}, "post:GetSignHash"),
			beego.NSRouter("/count", &controllers.TransactionsController{}, "get:GetTransactionsCount"),
		),

		beego.NSNamespace("/blocks",
			beego.NSRouter("/list", &controllers.BlocksController{}, "get:GetBlocks"),
			beego.NSRouter("/list-no-transactions", &controllers.BlocksController{}, "get:GetPlainBlocks"),
			beego.NSRouter("/query", &controllers.BlocksController{}, "get:GetBlockByHashOrNumOrTime"),
			beego.NSRouter("/transactions/count", &controllers.TransactionsController{}, "get:GetBlockTransactionCountByHashOrNumber"),
			beego.NSRouter("/latest", &controllers.BlocksController{}, "get:GetLatestBlock"),
			beego.NSRouter("/transactions/average-time", &controllers.TransactionsController{}, "get:GetTxAvgTimeByBlockNumber"),
			beego.NSRouter("/average-time", &controllers.BlocksController{}, "get:GetAvgGenerateTimeByBlockNumber"),
		),

		beego.NSNamespace("/contracts",
			beego.NSRouter("/compile", &controllers.ContractsController{}, "post:CompileContract"),
			beego.NSRouter("/deploy", &controllers.ContractsController{}, "post:DeployContract"),
			beego.NSRouter("/invoke", &controllers.ContractsController{}, "post:InvokeContract"),
			beego.NSRouter("/query", &controllers.ContractsController{}, "get:GetCode"),
			beego.NSRouter("/homomorphic/check", &controllers.ContractsController{}, "post:CheckHmValue"),
			beego.NSRouter("/homomorphic/encry", &controllers.ContractsController{}, "post:EncryptoMessage"),
		),

		beego.NSNamespace("/accounts",
			beego.NSRouter("/:address/contracts/count", &controllers.ContractsController{}, "get:GetContractCountByAddr"),
		),

		beego.NSNamespace("/nodes",
			beego.NSRouter("/list", &controllers.NodesController{}, "get:GetNodes"),
			beego.NSRouter("/tcert", &controllers.CertController{}, "get:GetTCert"),
			beego.NSRouter("/hash", &controllers.NodesController{}, "get:GetNodeHash"),
			beego.NSRouter("/delete", &controllers.NodesController{}, "delete:DelNode"),
		),
	)
	beego.AddNamespace(ns)
}
