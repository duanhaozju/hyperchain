package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/hpc"
	"hyperchain/jsonrpc/restful/utils"
	"strconv"
)

const (
	HashLength    = 32
)

type BlocksController struct {
	beego.Controller
}

func (b *BlocksController) GetTransactionByBlockNumberAndIndex() {
	blkNum, err := utils.CheckBlockNumber(b.Ctx.Input.Param(":blkNum"))
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}
	index, err := utils.CheckNumber(b.Ctx.Input.Param(":index"))
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}

	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

	tx, err := PublicTxAPI.GetTransactionByBlockNumberAndIndex(blkNum, index)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		b.Data["json"] = NewJSONObject(tx, nil)
	}
	b.ServeJSON()
}

func (b *BlocksController) GetTransactionByBlockHashAndIndex() {
	blkHash, err := utils.CheckHash(b.Ctx.Input.Param(":blkHash"))
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}
	index, err := utils.CheckNumber(b.Ctx.Input.Param(":index"))
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}

	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

	tx, err := PublicTxAPI.GetTransactionByBlockHashAndIndex(blkHash, index)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		b.Data["json"] = NewJSONObject(tx, nil)
	}
	b.ServeJSON()
}

func (b *BlocksController) GetBlocks() {
	from := b.Input().Get("from")
	to := b.Input().Get("to")
	args, err := utils.CheckIntervalArgs(from, to)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}

	PublicBlockAPIInterface := hpc.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*hpc.PublicBlockAPI)

	blks, err := PublicBlockAPI.GetBlocks(args)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		b.Data["json"] = NewJSONObject(blks, nil)
	}
	b.ServeJSON()
}

func (b *BlocksController) GetBlockByHashOrNum() {
	p_blkNum := b.Input().Get("blkNum")
	p_blkHash := b.Input().Get("blkHash")

	PublicBlockAPIInterface := hpc.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*hpc.PublicBlockAPI)

	var counts_params int = 0

	if p_blkNum != ""{
		counts_params++
	}
	if p_blkHash != "" {
		counts_params++
	}

	if counts_params != 1 {
		counts_params_str := strconv.Itoa(counts_params)
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{"require 1 params, but get "+counts_params_str+" params"})
		b.ServeJSON()
		return
	}

	if p_blkNum != "" {
		if blkNum, err := utils.CheckBlockNumber(p_blkNum); err != nil {
			b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		} else {
			if block, err := PublicBlockAPI.GetBlockByNumber(blkNum); err != nil {
				b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
			} else {
				b.Data["json"] = NewJSONObject(block, nil)
			}
		}
	} else if p_blkHash != "" {
		if blkHash, err := utils.CheckHash(p_blkHash); err != nil {
			b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		} else {
			if block, err := PublicBlockAPI.GetBlockByHash(blkHash); err != nil {
				b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
			} else {
				b.Data["json"] = NewJSONObject(block, nil)
			}
		}
	} else {
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{"invalid params"})
	}

	b.ServeJSON()
	return
}

func (b *BlocksController) GetBlockTransactionCountByHash() {

	hash, err := utils.CheckHash(b.Ctx.Input.Param(":blkHash"))
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else {
		PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
		PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

		num, err := PublicTxAPI.GetBlockTransactionCountByHash(hash)
		if err != nil {
			b.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
		} else {
			b.Data["json"] = NewJSONObject(num, nil)
		}
	}

	b.ServeJSON()
}
