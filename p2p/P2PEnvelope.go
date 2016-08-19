package p2p
import (
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/core/node"
)
//信息传输时候的信封，包括各类信息
// 节点之间交换的信息的格式定义
// 节点之间主要交换3种信息：
// 1. transaction
// 2. block
// 3. node
type Envelope struct {
	Transactions []types.Transaction
	Blocks []types.Block
	Nodes []node.Node
}