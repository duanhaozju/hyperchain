package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/hyperchain/hyperchain/api"
	"github.com/hyperchain/hyperchain/api/rest/utils"
	"github.com/hyperchain/hyperchain/common"
)

type TransactionsController struct {
	beego.Controller
	PublicTxAPI *api.Transaction
}

type requestInterval struct {
	From *api.BlockNumber `form:"from"`
	To   *api.BlockNumber `form:"to"`
}

func (t *TransactionsController) Prepare() {
	PublicTxAPIInterface := api.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*api.Transaction)
	t.PublicTxAPI = PublicTxAPI
}

func (t *TransactionsController) SendTransaction() {
	var args api.SendTxArgs

	if err := json.Unmarshal(t.Ctx.Input.RequestBody, &args); err != nil {
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	hash, err := t.PublicTxAPI.SendTransaction(args)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, err)
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
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	txs, err := t.PublicTxAPI.GetTransactions(args)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, err)
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
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	tx, err := t.PublicTxAPI.GetTransactionByHash(hash)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, err)
	} else {
		t.Data["json"] = NewJSONObject(tx, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetTransactionByBlockNumberOrBlockHashOrTime() {
	p_blkNum := t.Input().Get("blockNumber")
	p_blkHash := t.Input().Get("blockHash")
	p_index := t.Input().Get("index")

	p_startTime := t.Input().Get("startTime")
	p_endTime := t.Input().Get("endTime")

	flag := 0 // "1" means query by blockNumber and index, "2" means query by blockHash and index, "3" means query by startTime and endTime.

	if p_blkNum != "" && p_index != "" && p_blkHash == "" {
		flag = 1
	} else if p_blkHash != "" && p_index != "" && p_blkNum == "" {
		flag = 2
	} else if p_startTime != "" && p_endTime != "" && p_blkNum == "" && p_blkHash == "" && p_index == "" {
		flag = 3
	} else {
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"The number of params or the name of params is invalid"})
		t.ServeJSON()
		return
	}

	if flag == 1 {
		if blkNum, index, err := utils.CheckBlkNumAndIndexParams(p_blkNum, p_index); err != nil {
			t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		} else {
			if tx, err := t.PublicTxAPI.GetTransactionByBlockNumberAndIndex(blkNum, index); err != nil {
				t.Data["json"] = NewJSONObject(nil, err)
			} else {
				t.Data["json"] = NewJSONObject(tx, nil)
			}
		}
	} else if flag == 2 {
		if blkHash, index, err := utils.CheckBlkHashAndIndexParams(p_blkHash, p_index); err != nil {
			t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		} else {
			if tx, err := t.PublicTxAPI.GetTransactionByBlockHashAndIndex(blkHash, index); err != nil {
				t.Data["json"] = NewJSONObject(nil, err)
			} else {
				t.Data["json"] = NewJSONObject(tx, nil)
			}
		}
	} else if flag == 3 {
		if args, err := utils.CheckIntervalTimeArgs(p_startTime, p_endTime); err != nil {
			t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"Invalid params, value may be out of range"})
		} else {
			if txs, err := t.PublicTxAPI.GetTransactionsByTime(args); err != nil {
				t.Data["json"] = NewJSONObject(nil, err)
			} else {
				t.Data["json"] = NewJSONObject(txs, nil)
			}
		}

	} else {
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"invalid params"})
	}

	t.ServeJSON()
	return
}

func (t *TransactionsController) GetTransactionReceipt() {
	hash, err := utils.CheckHash(t.Ctx.Input.Param(":transactionHash"))
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	rep, err := t.PublicTxAPI.GetTransactionReceipt(hash)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, err)
	} else {
		if rep != nil && rep.Ret == "0x0" {
			if tx, err := t.PublicTxAPI.GetTransactionByHash(hash); err != nil {
				t.Data["json"] = NewJSONObject(nil, err)
			} else if tx.Invalid == true {
				// 交易非法
				t.Data["json"] = NewJSONObject(nil, &common.CallbackError{tx.InvalidMsg})
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
	var args api.SendTxArgs

	if err := json.Unmarshal(t.Ctx.Input.RequestBody, &args); err != nil {
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	hash, err := t.PublicTxAPI.GetSignHash(args)
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, err)
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
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		t.ServeJSON()
		return
	}

	num, err := t.PublicTxAPI.GetTxAvgTimeByBlockNumber(args)

	if err != nil {
		t.Data["json"] = NewJSONObject(nil, err)
	} else {
		t.Data["json"] = NewJSONObject(num, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetTransactionsCount() {

	count, err := t.PublicTxAPI.GetTransactionsCount()
	if err != nil {
		t.Data["json"] = NewJSONObject(nil, err)
	} else {
		t.Data["json"] = NewJSONObject(count, nil)
	}
	t.ServeJSON()
}

func (t *TransactionsController) GetBlockTransactionCountByHashOrNumber() {

	p_blkNum := t.Input().Get("blockNumber")
	p_blkHash := t.Input().Get("blockHash")

	flag := 0 // "1" means query by blockNumber, "2" means query by blockHash

	if p_blkNum != "" && p_blkHash == "" {
		flag = 1
	} else if p_blkHash != "" && p_blkNum == "" {
		flag = 2
	} else {
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"The number of params or the name of params is invalid"})
		t.ServeJSON()
		return
	}

	if flag == 1 {
		if blkNum, err := utils.CheckBlockNumber(p_blkNum); err != nil {
			t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		} else {
			if count, err := t.PublicTxAPI.GetBlockTransactionCountByNumber(blkNum); err != nil {
				t.Data["json"] = NewJSONObject(nil, err)
			} else {
				t.Data["json"] = NewJSONObject(count, nil)
			}
		}
	} else if flag == 2 {
		if blkHash, err := utils.CheckHash(p_blkHash); err != nil {
			t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		} else {
			if count, err := t.PublicTxAPI.GetBlockTransactionCountByHash(blkHash); err != nil {
				t.Data["json"] = NewJSONObject(nil, err)
			} else {
				t.Data["json"] = NewJSONObject(count, nil)
			}
		}
	} else {
		t.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"invalid params"})
	}

	t.ServeJSON()
	return
}
