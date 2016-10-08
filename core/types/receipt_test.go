package types

import (
	proto "github.com/golang/protobuf/proto"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	"hyperchain/core/vm"
	"math/big"
	"testing"
)

type ReceiptSuite struct {
}

func Test(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&ReceiptSuite{})

func (s *ReceiptSuite) TestEncodeDecode(c *checker.C) {
	var logs vm.Logs
	var log vm.Log
	log = NewTestLog()
	logs = append(logs, &log)
	logs = append(logs, &log)
	receipt := NewReceipt([]byte("root"), big.NewInt(123))
	receipt.SetLogs(logs)
	buf, _ := proto.Marshal(receipt)
	var newReceipt Receipt
	proto.Unmarshal(buf, &newReceipt)

	newLogs, _ := newReceipt.GetLogs()
	c.Assert(newLogs[0].Address, checker.DeepEquals, common.StringToAddress("123456"))
}
func NewTestLog() vm.Log {
	topic := []common.Hash{
		common.StringToHash("topic1"),
		common.StringToHash("topic2"),
		common.StringToHash("topic3"),
	}
	log := vm.NewLog(common.StringToAddress("123456"), topic, []byte("data"), 123)
	(*log).BlockNumber = 123
	(*log).TxHash = common.StringToHash("txhash")
	(*log).TxIndex = 2
	(*log).BlockHash = common.StringToHash("blockhash")
	(*log).Index = 1
	return *log
}
