package controllers

import (
	"github.com/astaxie/beego"
	"encoding/json"
	"hyperchain/hpc"
	"hyperchain/jsonrpc/restful/utils"
)

type ContractsController struct {
	beego.Controller
}

func (c *ContractsController) CompileContract() {

	jsonObj := struct {
		Source string
	}{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj); err != nil {
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	compiled_res, err := PublicContractAPI.CompileContract(jsonObj.Source)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		c.Data["json"] = NewJSONObject(compiled_res, nil)
	}

	c.ServeJSON()
}

func (c *ContractsController) DeployContract() {
	var args hpc.SendTxArgs

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &args); err != nil {
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	hash, err := PublicContractAPI.DeployContract(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		c.Data["json"] = NewJSONObject(hash, nil)
	}

	c.ServeJSON()
}

func (c *ContractsController) InvokeContract() {
	var args hpc.SendTxArgs

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &args); err != nil {
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	hash, err := PublicContractAPI.InvokeContract(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		c.Data["json"] = NewJSONObject(hash, nil)
	}

	c.ServeJSON()
}

func (c *ContractsController) GetCode() {

	p_blkNum := c.Input().Get("blockNumber")
	p_address := c.Input().Get("address") // contract address

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	// check the number of params
	if p_blkNum == ""{
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{"the param 'blockNumber' can't be empty"})
		c.ServeJSON()
		return
	}
	if p_address == "" {
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{"the param 'address' can't be empty"})
		c.ServeJSON()
		return
	}

	// check params type
	if blkNum, err := utils.CheckBlockNumber(p_blkNum); err != nil {
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else if address, err := utils.CheckAddress(p_address); err != nil{
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else {
		if block, err := PublicContractAPI.GetCode(address, blkNum); err != nil {
			c.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
		} else {
			c.Data["json"] = NewJSONObject(block, nil)
		}
	}
	c.ServeJSON()
}
