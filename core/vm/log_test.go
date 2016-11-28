//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

import (
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	"testing"
)

type LogSuite struct {
}

func Test(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&LogSuite{})

func (s *LogSuite) TestEncodeDecode(c *checker.C) {
	oldLog := NewTestLog()
	buf, err := (&oldLog).EncodeLog()
	if err != nil {
		c.Fatal("marshal failed")
	}

	var newLog Log
	newLog, err = DecodeLog(buf)
	if err != nil {
		c.Fatal("unmarshal failed")
	}
	CheckLogEqual(oldLog, newLog, c)
}
func CheckLogEqual(log1, log2 Log, c *checker.C) {
	c.Assert(log1.Address, checker.DeepEquals, log2.Address)
	c.Assert(log1.Topics, checker.DeepEquals, log2.Topics)
	c.Assert(log1.Data, checker.DeepEquals, log2.Data)
	c.Assert(log1.BlockNumber, checker.DeepEquals, log2.BlockNumber)
	c.Assert(log1.TxHash, checker.DeepEquals, log2.TxHash)
	c.Assert(log1.TxIndex, checker.DeepEquals, log2.TxIndex)
	c.Assert(log1.BlockHash, checker.DeepEquals, log2.BlockHash)
	c.Assert(log1.Index, checker.DeepEquals, log2.Index)
}
func NewTestLog() Log {
	topic := []common.Hash{
		common.StringToHash("topic1"),
		common.StringToHash("topic2"),
		common.StringToHash("topic3"),
	}
	log := NewLog(common.StringToAddress("123456"), topic, []byte("data"), 123)
	(*log).BlockNumber = 123
	(*log).TxHash = common.StringToHash("txhash")
	(*log).TxIndex = 2
	(*log).BlockHash = common.StringToHash("blockhash")
	(*log).Index = 1
	return *log
}
func (s *LogSuite) TestEncodeDecodeLogs(c *checker.C) {
	var logs Logs
	var logs2 Logs
	var log Log
	log = NewTestLog()
	logs = append(logs, &log)
	logs = append(logs, &log)
	buf, _ := (&logs).EncodeLogs()

	logs2, _ = DecodeLogs(buf)
	for index, log := range logs2 {
		CheckLogEqual(*logs[index], *log, c)
	}
}
