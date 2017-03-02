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

	if nodes, err := n.PublicNodeAPI.GetNodes(); err != nil {
		n.Data["json"] = NewJSONObject(nil, err)
	} else {
		n.Data["json"] = NewJSONObject(nodes, nil)
	}
	n.ServeJSON()
}

func (n *NodesController) GetNodeHash() {
	if hash, err := n.PublicNodeAPI.GetNodeHash(); err != nil {
		n.Data["json"] = NewJSONObject(nil, err)
	} else {
		n.Data["json"] = NewJSONObject(hash, nil)
	}
	n.ServeJSON()
}

func (n *NodesController) DelNode() {
	var args hpc.NodeArgs
	args.NodeHash = n.Input().Get("hash")

	err := n.PublicNodeAPI.DelNode(args)
	if err != nil {
		n.Data["json"] = NewJSONObject(nil, err)
	} else {
		n.Data["json"] = NewJSONObject(true, nil)
	}

	n.ServeJSON()
}
