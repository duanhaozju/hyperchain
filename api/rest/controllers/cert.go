package controllers

import (
	"github.com/astaxie/beego"
	"github.com/hyperchain/hyperchain/api"
)

type CertController struct {
	beego.Controller
	PublicCertAPI *api.Cert
}

func (c *CertController) Prepare() {
	PublicCertAPIInterface := api.GetApiObjectByNamespace("cert").Service
	PublicCertAPI := PublicCertAPIInterface.(*api.Cert)
	c.PublicCertAPI = PublicCertAPI
}

func (c *CertController) GetTCert() {

	var args api.CertArgs
	args.Pubkey = c.Input().Get("pubkey")

	tcertReturn, err := c.PublicCertAPI.GetTCert(args)
	if err != nil {
		c.Data["json"] = NewJSONObject(nil, err)
	} else {
		c.Data["json"] = NewJSONObject(tcertReturn, nil)
	}

	c.ServeJSON()
}
