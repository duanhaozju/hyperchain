package p2p
import (
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/core/node"
	"strconv"
	"time"
	"encoding/hex"
	"hyperchain-alpha/core"
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
	Chain types.Chain
}

func (envelope Envelope)String() string{
	printString := "\n================<ENVELOPE>======================\n"
	printString +=  "DATA TO EXCHANGE:\t交易数据一共："+ strconv.Itoa(len(envelope.Transactions)) +"\n"
	printString +=  "DATA TO EXCHANGE:\t节点一共有：" + strconv.Itoa(len(envelope.Nodes)) +"\n"
	printString +=  "DATA TO EXCHANGE:\t区块一共有：" + strconv.Itoa(len(envelope.Blocks)) +"\n"
	printString +=  "DATA TO EXCHANGE:\t节点最新Chain Header Hash：" + hex.EncodeToString([]byte(core.GetChain().LastestBlockHash)) +"\n"
	printString +=  "=================</ENVELOPE>=====================\n"
	return printString
}