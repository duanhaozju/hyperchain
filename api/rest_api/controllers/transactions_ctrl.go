package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/api"
	"hyperchain/api/rest_api/utils"
	"encoding/json"
	"strconv"
)

type TransactionsController struct {
	beego.Controller
	PublicTxAPI *hpc.PublicTransactionAPI
}

type requestInterval struct {
	From *hpc.BlockNumber `form:"from"`
	To *hpc.BlockNumber `form:"to"`
}

func (t *TransactionsController) Prepare() {
	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)
	t.PublicTxAPI = PublicTxAPI
}

func (t *TransactionsController) SendTransaction() {
	var args hpc.SendTxArgs

	if err := json.Unmarshal(t.Ctx.Input.RequestBody, &args); err != nil {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	hash, err := t.PublicTxAPI.SendTransaction(args)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		t.Data["json"] = NewJSONObject(hash, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetTransactions() {

	from := t.Input().Get("from")
	to := t.Input().Get("to")

	args, err := utils.CheckIntervalArgs(from, to)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	txs, err := t.PublicTxAPI.GetTransactions(args)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		t.Data["json"] = NewJSONObject(txs, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetTransactionByHash() {

	//log.Error(t.Ctx.Input.Param(":transactionHash"))
	hash, err := utils.CheckHash(t.Ctx.Input.Param(":transactionHash"))
	//log.Error(hash.Hex())
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	tx, err := t.PublicTxAPI.GetTransactionByHash(hash)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		t.Data["json"] = NewJSONObject(tx, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetTransactionByBlockNumberOrBlockHash() {
	p_blkNum := t.Input().Get("blockNumber")
	p_blkHash := t.Input().Get("blockHash")
	p_index := t.Input().Get("index")

	var counts_params int = 0

	if p_blkNum != ""{
		counts_params++
	}
	if p_blkHash != "" {
		counts_params++
	}
	if p_index != "" {
		counts_params++
	}

	if counts_params != 2 {
		counts_params_str := strconv.Itoa(counts_params)
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{"require 2 params, but get "+counts_params_str+" params"})
		t.ServeJSON()
		return
	}

	if p_blkNum != "" {
		if blkNum, index, err := utils.CheckBlkNumAndIndexParams(p_blkNum, p_index); err != nil {
			t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		} else {
			if tx, err := t.PublicTxAPI.GetTransactionByBlockNumberAndIndex(blkNum, index); err != nil {
				t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
			} else {
				t.Data["json"] = NewJSONObject(tx, nil)
			}
		}
	} else if p_blkHash != "" {
		if blkHash, index, err := utils.CheckBlkHashAndIndexParams(p_blkHash, p_index); err != nil {
			t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		} else {
			if tx, err := t.PublicTxAPI.GetTransactionByBlockHashAndIndex(blkHash, index); err != nil {
				t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
			} else {
				t.Data["json"] = NewJSONObject(tx, nil)
			}
		}
	} else {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{"invalid params"})
	}

	t.ServeJSON()
	return
}

func (t *TransactionsController) GetTransactionReceipt() {
	hash, err := utils.CheckHash(t.Ctx.Input.Param(":transactionHash"))
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	rep, err := t.PublicTxAPI.GetTransactionReceipt(hash)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		if rep != nil && rep.Ret == "0x0" {
			if tx, err := t.PublicTxAPI.GetTransactionByHash(hash);err != nil {
				t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
			} else if tx.Invalid == true {
				// 交易非法
				t.Data["json"] = NewJSONObject(nil, &callbackError{tx.InvalidMsg})
			} else {
				// 交易合法
				t.Data["json"] = NewJSONObject(rep, nil)
			}

		} else {
			t.Data["json"] = NewJSONObject(rep, nil)
		}
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetSignHash() {
	var args hpc.SendTxArgs

	if err := json.Unmarshal(t.Ctx.Input.RequestBody, &args); err != nil {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	hash, err := t.PublicTxAPI.GetSignHash(args)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		t.Data["json"] = NewJSONObject(hash, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetTxAvgTimeByBlockNumber() {
	from := t.Input().Get("from")
	to := t.Input().Get("to")
	args, err := utils.CheckIntervalArgs(from, to)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	num, err := t.PublicTxAPI.GetTxAvgTimeByBlockNumber(args)

	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		t.Data["json"] = NewJSONObject(num, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetTransactionsCount() {

	count, err := t.PublicTxAPI.GetTransactionsCount()
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		t.Data["json"] = NewJSONObject(count, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetBlockTransactionCountByHash() {

	hash, err := utils.CheckHash(t.Ctx.Input.Param(":blockHash"))
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else {

		num, err := t.PublicTxAPI.GetBlockTransactionCountByHash(hash)
		if err != nil {
			t.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
		} else {
			t.Data["json"] = NewJSONObject(num, nil)
		}
	}

	t.ServeJSON()
}