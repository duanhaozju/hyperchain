package controllers

import (
	"github.com/astaxie/beego"
	"hyperchain/hpc"
)

type NodesController struct {
	beego.Controller
}

func (n *NodesController) GetNodes() {

	PublicNodeAPIInterface := hpc.GetApiObjectByNamespace("node").Service
	PublicNodeAPI := PublicNodeAPIInterface.(*hpc.PublicNodeAPI)

	if nodes, err := PublicNodeAPI.GetNodes();err != nil {
		n.Data["json"] = NewJSONObject(nil, &callbackError{err.Error()})
	} else {
		n.Data["json"] = NewJSONObject(nodes, nil)
	}
	n.ServeJSON()
}
