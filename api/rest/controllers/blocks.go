package controllers

import (
	"github.com/astaxie/beego"
	"github.com/hyperchain/hyperchain/api"
	"github.com/hyperchain/hyperchain/api/rest/utils"
	"github.com/hyperchain/hyperchain/common"
)

type BlocksController struct {
	beego.Controller
	PublicBlockAPI *api.Block
}

func (b *BlocksController) Prepare() {
	PublicBlockAPIInterface := api.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*api.Block)
	b.PublicBlockAPI = PublicBlockAPI
}

func (b *BlocksController) GetBlocks() {
	from := b.Input().Get("from")
	to := b.Input().Get("to")
	args, err := utils.CheckIntervalArgs(from, to)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
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

func (b *BlocksController) GetPlainBlocks() {
	from := b.Input().Get("from")
	to := b.Input().Get("to")
	args, err := utils.CheckIntervalArgs(from, to)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}

	blks, err := b.PublicBlockAPI.GetPlainBlocks(args)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, err)
	} else {
		b.Data["json"] = NewJSONObject(blks, nil)
	}
	b.ServeJSON()
}

func (b *BlocksController) GetBlockByHashOrNumOrTime() {
	p_blkNum := b.Input().Get("blockNumber")
	p_blkHash := b.Input().Get("blockHash")

	p_startTime := b.Input().Get("startTime")
	p_endTime := b.Input().Get("endTime")

	flag := 0 // "1" means query by blockNumber, "2" means query by blockHash, "3" means query by startTime and endTime.

	if p_blkNum != "" && p_blkHash == "" && p_startTime == "" && p_endTime == "" {
		flag = 1
	} else if p_blkHash != "" && p_blkNum == "" && p_startTime == "" && p_endTime == "" {
		flag = 2
	} else if p_startTime != "" && p_endTime != "" && p_blkNum == "" && p_blkHash == "" {
		flag = 3
	} else {
		b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"The number of params or the name of params is invalid"})
		b.ServeJSON()
		return
	}

	if flag == 1 {
		if blkNum, err := utils.CheckBlockNumber(p_blkNum); err != nil {
			b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		} else {
			if block, err := b.PublicBlockAPI.GetBlockByNumber(blkNum); err != nil {
				b.Data["json"] = NewJSONObject(nil, err)
			} else {
				b.Data["json"] = NewJSONObject(block, nil)
			}
		}
	} else if flag == 2 {
		if blkHash, err := utils.CheckHash(p_blkHash); err != nil {
			b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		} else {
			if block, err := b.PublicBlockAPI.GetBlockByHash(blkHash); err != nil {
				b.Data["json"] = NewJSONObject(nil, err)
			} else {
				b.Data["json"] = NewJSONObject(block, nil)
			}
		}
	} else if flag == 3 {
		if args, err := utils.CheckIntervalTimeArgs(p_startTime, p_endTime); err != nil {
			b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"Invalid params, value may be out of range"})
		} else {
			if blks, err := b.PublicBlockAPI.GetBlocksByTime(args); err != nil {
				b.Data["json"] = NewJSONObject(nil, err)
			} else {
				b.Data["json"] = NewJSONObject(blks, nil)
			}
		}

	} else {
		b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"invalid params"})
	}

	b.ServeJSON()
}

func (b *BlocksController) GetLatestBlock() {

	block, err := b.PublicBlockAPI.LatestBlock()
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, err)
	} else {
		b.Data["json"] = NewJSONObject(block, nil)
	}
	b.ServeJSON()
}

func (b *BlocksController) GetAvgGenerateTimeByBlockNumber() {

	from := b.Input().Get("from")
	to := b.Input().Get("to")
	args, err := utils.CheckIntervalArgs(from, to)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		b.ServeJSON()
		return
	}

	time, err := b.PublicBlockAPI.GetAvgGenerateTimeByBlockNumber(args)
	if err != nil {
		b.Data["json"] = NewJSONObject(nil, err)
	} else {
		b.Data["json"] = NewJSONObject(time, nil)
	}
	b.ServeJSON()
}
