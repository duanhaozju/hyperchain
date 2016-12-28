package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/api"
)

type NodesController struct {
	beego.Controller
	PublicNodeAPI *hpc.PublicNodeAPI

}

func (n *NodesController) Prepare() {
	PublicNodeAPIInterface := hpc.GetApiObjectByNamespace("node").Service
	PublicNodeAPI := PublicNodeAPIInterface.(*hpc.PublicNodeAPI)
	n.PublicNodeAPI = PublicNodeAPI
}

func (n *NodesController) GetNodes() {

	if nodes, err := n.PublicNodeAPI.GetNodes();err != nil {
		n.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		n.Data["json"] = NewJSONObject(nodes, nil)
	}
	n.ServeJSON()
}
