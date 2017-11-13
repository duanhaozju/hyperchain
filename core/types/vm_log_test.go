// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package types

import (
	"github.com/hyperchain/hyperchain/common"
	checker "gopkg.in/check.v1"
	"testing"
)

type LogSuite struct {
}

func TestVmLog(t *testing.T) {
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
func NewTestLogTrans() LogTrans {
	topics := []string{
		common.StringToHash("topic1").Hex(),
		common.StringToHash("topic2").Hex(),
		common.StringToHash("topic3").Hex(),
	}
	lt := LogTrans{
		Address:     common.StringToAddress("123456").Hex(),
		Topics:      topics,
		Data:        common.Bytes2Hex([]byte("data")),
		BlockNumber: 123,
		TxHash:      common.StringToHash("txhash").Hex(),
		TxIndex:     2,
		BlockHash:   common.StringToHash("blockhash").Hex(),
		Index:       1,
	}
	return lt
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
func (s *LogSuite) TestToLogsTrans(c *checker.C) {
	var logs Logs
	var log Log
	log = NewTestLog()
	logs = append(logs, &log)
	logs = append(logs, &log)

	logTrans := logs.ToLogsTrans()
	lt1 := NewTestLogTrans()
	for _, lt := range logTrans {
		CheckTranEqual(lt1, lt, c)
	}

}

func CheckTranEqual(lt1, lt2 LogTrans, c *checker.C) {
	c.Assert(lt1.Address, checker.DeepEquals, lt2.Address)
	c.Assert(lt1.Topics, checker.DeepEquals, lt2.Topics)
	c.Assert(lt1.Data, checker.DeepEquals, lt2.Data)
	c.Assert(lt1.BlockNumber, checker.DeepEquals, lt2.BlockNumber)
	c.Assert(lt1.TxHash, checker.DeepEquals, lt2.TxHash)
	c.Assert(lt1.TxIndex, checker.DeepEquals, lt2.TxIndex)
	c.Assert(lt1.BlockHash, checker.DeepEquals, lt2.BlockHash)
	c.Assert(lt1.Index, checker.DeepEquals, lt2.Index)
}
