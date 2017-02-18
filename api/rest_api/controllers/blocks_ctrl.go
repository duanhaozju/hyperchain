package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/api"
	"hyperchain/api/rest_api/utils"
	"strconv"
)

const (
	HashLength = 32
)

type BlocksController struct {
	beego.Controller
	PublicBlockAPI *hpc.PublicBlockAPI
}

func (b *BlocksController) Prepare() {
	PublicBlockAPIInterface := hpc.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*hpc.PublicBlockAPI)
	b.PublicBlockAPI = PublicBlockAPI
}

func (b *BlocksController) GetBlocks() {
	from := b.Input().Get("from")
	to := b.Input().Get("to")
	args, err := utils.CheckIntervalArgs(from, to)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &hpc.InvalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}

	blks, err := b.PublicBlockAPI.GetBlocks(args)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, err)
	} else {
		b.Data["json"] = NewJSONObject(blks, nil)
	}
	b.ServeJSON()
}

func (b *BlocksController) GetBlockByHashOrNum() {
	p_blkNum := b.Input().Get("blockNumber")
	p_blkHash := b.Input().Get("blockHash")

	var counts_params int = 0

	if p_blkNum != "" {
		counts_params++
	}
	if p_blkHash != "" {
		counts_params++
	}

	if counts_params != 1 {
		counts_params_str := strconv.Itoa(counts_params)
		b.Data["json"] = NewJSONObject(nil, &hpc.InvalidParamsError{"require 1 params, but get " + counts_params_str + " params"})
		b.ServeJSON()
		return
	}

	if p_blkNum != "" {
		if blkNum, err := utils.CheckBlockNumber(p_blkNum); err != nil {
			b.Data["json"] = NewJSONObject(nil, &hpc.InvalidParamsError{err.Error()})
		} else {
			if block, err := b.PublicBlockAPI.GetBlockByNumber(blkNum); err != nil {
				b.Data["json"] = NewJSONObject(nil, err)
			} else {
				b.Data["json"] = NewJSONObject(block, nil)
			}
		}
	} else if p_blkHash != "" {
		if blkHash, err := utils.CheckHash(p_blkHash); err != nil {
			b.Data["json"] = NewJSONObject(nil, &hpc.InvalidParamsError{err.Error()})
		} else {
			if block, err := b.PublicBlockAPI.GetBlockByHash(blkHash); err != nil {
				b.Data["json"] = NewJSONObject(nil, err)
			} else {
				b.Data["json"] = NewJSONObject(block, nil)
			}
		}
	} else {
		b.Data["json"] = NewJSONObject(nil, &hpc.InvalidParamsError{"invalid params"})
	}

	b.ServeJSON()
	return
}

//func (b *BlocksController) GetBlockTransactionCountByHash() {
//
//	hash, err := utils.CheckHash(b.Ctx.Input.Param(":blockHash"))
//	if err != nil {
//		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
//	} else {
//		PublicTxAPIInterface := api.GetApiObjectByNamespace("tx").Service
//		PublicTxAPI := PublicTxAPIInterface.(*api.PublicTransactionAPI)
//
//		num, err := PublicTxAPI.GetBlockTransactionCountByHash(hash)
//		if err != nil {
//			b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
//		} else {
//			b.Data["json"] = NewJSONObject(num, nil)
//		}
//	}
//
//	b.ServeJSON()
//}

func (b *BlocksController) GetLatestBlock() {

	block, err := b.PublicBlockAPI.LatestBlock()
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, err)
	} else {
		b.Data["json"] = NewJSONObject(block, nil)
	}
	b.ServeJSON()
}
