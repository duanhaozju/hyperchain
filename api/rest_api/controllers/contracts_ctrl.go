package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"hyperchain/api"
	"hyperchain/api/rest_api/utils"
)

type ContractsController struct {
	beego.Controller
	PublicContractAPI *hpc.PublicContractAPI
}

func (c *ContractsController) Prepare() {
	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)
	c.PublicContractAPI = PublicContractAPI
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

	compiled_res, err := c.PublicContractAPI.CompileContract(jsonObj.Source)
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

	hash, err := c.PublicContractAPI.DeployContract(args)
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

	hash, err := c.PublicContractAPI.InvokeContract(args)
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

	// check the number of params
	if p_blkNum == "" {
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
	} else if address, err := utils.CheckAddress(p_address); err != nil {
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else {
		if block, err := c.PublicContractAPI.GetCode(address, blkNum); err != nil {
			c.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
		} else {
			c.Data["json"] = NewJSONObject(block, nil)
		}
	}
	c.ServeJSON()
}

func (c *ContractsController) GetContractCountByAddr() {
	p_blkNum := c.Input().Get("blockNumber")
	p_address := c.Ctx.Input.Param(":address") // account address

	// check the number of params
	if p_blkNum == "" {
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
	} else if address, err := utils.CheckAddress(p_address); err != nil {
		c.Data["json"] = NewJSONObject(nil, &invalidParamsError{err.Error()})
	} else {
		if count, err := c.PublicContractAPI.GetContractCountByAddr(address, blkNum); err != nil {
			c.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
		} else {
			c.Data["json"] = NewJSONObject(count, nil)
		}
	}
	c.ServeJSON()
}
