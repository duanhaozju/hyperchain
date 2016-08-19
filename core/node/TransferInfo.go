package node
import (
	"hyperchain-alpha/core/types"
)
// 节点之间交换的信息的格式定义
// 节点之间主要交换3种信息：
// 1. transaction
// 2. block
// 3. node
type TransferInfo struct {
	Transactions []types.Transaction
	Blocks []types.Block
	Nodes []Node
}