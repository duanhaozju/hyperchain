package p2p
import "hyperchain.cn/app/jsonrpc/model"

//全局节点存储
var GLOBALNODES model.Nodes
//全局本地I节点存储
var LOCALNODE model.Node

//交易信息传输结构
type TxTransfer struct {
	Tx model.Transaction
	Node model.Node
}