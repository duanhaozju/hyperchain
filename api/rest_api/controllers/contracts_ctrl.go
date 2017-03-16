package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"hyperchain/api"
	"hyperchain/api/rest_api/utils"
	"hyperchain/common"
)

type ContractsController struct {
	beego.Controller
	PublicContractAPI *hpc.Contract
}

func (c *ContractsController) Prepare() {
	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.Contract)
	c.PublicContractAPI = PublicContractAPI
}

func (c *ContractsController) CompileContract() {

	jsonObj := struct {
		Source string
	}{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj); err != nil {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	compiled_res, err := c.PublicContractAPI.CompileContract(jsonObj.Source)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, err)
	} else {
		c.Data["json"] = NewJSONObject(compiled_res, nil)
	}

	c.ServeJSON()
}

func (c *ContractsController) DeployContract() {
	var args hpc.SendTxArgs

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &args); err != nil {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	hash, err := c.PublicContractAPI.DeployContract(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, err)
	} else {
		c.Data["json"] = NewJSONObject(hash, nil)
	}

	c.ServeJSON()
}

func (c *ContractsController) InvokeContract() {
	var args hpc.SendTxArgs

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &args); err != nil {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	hash, err := c.PublicContractAPI.InvokeContract(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, err)
	} else {
		c.Data["json"] = NewJSONObject(hash, nil)
	}

	c.ServeJSON()
}

func (c *ContractsController) GetCode() {

	p_address := c.Input().Get("address") // contract address

	// check the number of params
	if p_address == "" {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"the param 'address' can't be empty"})
		c.ServeJSON()
		return
	}

	// check params type
	if address, err := utils.CheckAddress(p_address); err != nil {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
	} else {
		if block, err := c.PublicContractAPI.GetCode(address); err != nil {
			c.Data["json"] = NewJSONObject(nil, err)
		} else {
			c.Data["json"] = NewJSONObject(block, nil)
		}
	}
	c.ServeJSON()
}

func (c *ContractsController) GetContractCountByAddr() {
	p_address := c.Ctx.Input.Param(":address") // account address

	// check the number of params
	if p_address == "" {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{"the param 'address' can't be empty"})
		c.ServeJSON()
		return
	}

	// check params type
	if address, err := utils.CheckAddress(p_address); err != nil {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
	} else {
		if count, err := c.PublicContractAPI.GetContractCountByAddr(address); err != nil {
			c.Data["json"] = NewJSONObject(nil, err)
		} else {
			c.Data["json"] = NewJSONObject(count, nil)
		}
	}
	c.ServeJSON()
}

func (c *ContractsController) CheckHmValue() {
	var args hpc.ValueArgs

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &args); err != nil {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	hash, err := c.PublicContractAPI.CheckHmValue(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, err)
	} else {
		c.Data["json"] = NewJSONObject(hash, nil)
	}

	c.ServeJSON()
}

func (c *ContractsController) EncryptoMessage() {
	var args hpc.EncryptoArgs

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &args); err != nil {
		c.Data["json"] = NewJSONObject(nil, &common.InvalidParamsError{err.Error()})
		c.ServeJSON()
		return
	}

	hash, err := c.PublicContractAPI.EncryptoMessage(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, err)
	} else {
		c.Data["json"] = NewJSONObject(hash, nil)
	}

	c.ServeJSON()
}