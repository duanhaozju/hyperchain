package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/api"
)

type CertController struct {
	beego.Controller
	PublicCertAPI *hpc.PublicCertAPI
}

func (c *CertController) Prepare() {
	PublicCertAPIInterface := hpc.GetApiObjectByNamespace("cert").Service
	PublicCertAPI := PublicCertAPIInterface.(*hpc.PublicCertAPI)
	c.PublicCertAPI = PublicCertAPI
}

func (c *CertController) GetTCert() {

	var args hpc.CertArgs
	args.Pubkey = c.Input().Get("pubkey")

	tcertReturn, err := c.PublicCertAPI.GetTCert(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, err)
	} else {
		c.Data["json"] = NewJSONObject(tcertReturn, nil)
	}

	c.ServeJSON()
}

