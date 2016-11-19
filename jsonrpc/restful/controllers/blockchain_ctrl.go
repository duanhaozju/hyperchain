package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/hpc"
)

type BlockchainController struct {
	beego.Controller
}

func (b *BlockchainController) GetTransactionsCount() {
	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

	count, err := PublicTxAPI.GetTransactionsCount()
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		b.Data["json"] = NewJSONObject(count, nil)
	}
	b.ServeJSON()
}

func (b *BlockchainController) GetLatestBlock() {
	PublicBlockAPIInterface := hpc.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*hpc.PublicBlockAPI)

	block, err := PublicBlockAPI.LatestBlock()
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		b.Data["json"] = NewJSONObject(block, nil)
	}
	b.ServeJSON()
}