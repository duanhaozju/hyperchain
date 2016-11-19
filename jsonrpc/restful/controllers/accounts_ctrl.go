package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/hpc"
	"hyperchain/jsonrpc/restful/utils"
)

type AccountsController struct {
	beego.Controller
}

func (a *AccountsController) GetContractCountByAddr() {
	p_blkNum := a.Input().Get("blkNum")
	p_address := a.Ctx.Input.Param(":address") // account address

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	// check the number of params
	if p_blkNum == ""{
		a.Data["json"] = NewJSONObject(nil, &invalidParamsError{"the param 'blkNum' can't be empty"})
		a.ServeJSON()
		return
	}
	if p_address == "" {
		a.Data["json"] = NewJSONObject(nil, &invalidParamsError{"the param 'address' can't be empty"})
		a.ServeJSON()
		return
	}

	// check params type
	if blkNum, err := utils.CheckBlockNumber(p_blkNum); err != nil {
		a.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else if address, err := utils.CheckAddress(p_address); err != nil{
		a.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else {
		if count, err := PublicContractAPI.GetContractCountByAddr(address, blkNum); err != nil {
			a.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
		} else {
			a.Data["json"] = NewJSONObject(count, nil)
		}
	}
	a.ServeJSON()
}
