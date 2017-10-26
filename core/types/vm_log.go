// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package types

import (
	"encoding/json"
	"fmt"
	"hyperchain/common"
)

type Log struct {
	// consensus fields
	Address common.Address `json:"Address"`
	Topics  []common.Hash  `json:"Topics"`
	Data    []byte         `json:"Data"`

	// Derived fields (don't reorder!)
	BlockNumber uint64      `json:"BlockNumber"`
	TxHash      common.Hash `json:"TxHash"`
	TxIndex     uint        `json:"TxIndex"`
	BlockHash   common.Hash `json:"BlockHash"`
	Index       uint        `json:"Index"`
}

// assign block number as 0 temporarily
// because the blcok number in env is a seqNo actually
// primary's seqNo may not equal to other's
// correctly block number and block hash will be assigned in the commit phase
func NewLog(address common.Address, topics []common.Hash, data []byte, number uint64) *Log {
	return &Log{Address: address, Topics: topics, Data: data, BlockNumber: 0}
}

func (l *Log) String() string {
	return fmt.Sprintf(`log: %x %x %x %x %d %x %d`, l.Address, l.Topics, l.Data, l.TxHash, l.TxIndex, l.BlockHash, l.Index)
}

func (l *Log) EncodeLog() ([]byte, error) {
	return json.Marshal(*l)
}

func DecodeLog(buf []byte) (Log, error) {
	var tmp Log
	err := json.Unmarshal(buf, &tmp)
	return tmp, err
}

type Logs []*Log

func (ls *Logs) EncodeLogs() ([]byte, error) {
	return json.Marshal(*ls)
}

func DecodeLogs(buf []byte) (Logs, error) {
	var tmp Logs
	err := json.Unmarshal(buf, &tmp)
	return tmp, err
}

type LogTrans struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber uint64   `json:"blockNumber"`
	BlockHash   string   `json:"blockHash"`
	TxHash      string   `json:"txHash"`
	TxIndex     uint     `json:"txIndex"`
	Index       uint     `json:"index"`
}

func (ls Logs) ToLogsTrans() []LogTrans {
	var ret = make([]LogTrans, len(ls))
	for idx, log := range ls {
		var topics = make([]string, len(log.Topics))
		for ti, t := range log.Topics {
			topics[ti] = t.Hex()
		}
		ret[idx] = LogTrans{
			Address:     log.Address.Hex(),
			Data:        common.Bytes2Hex(log.Data),
			BlockNumber: log.BlockNumber,
			BlockHash:   log.BlockHash.Hex(),
			Topics:      topics,
			TxHash:      log.TxHash.Hex(),
			Index:       log.Index,
			TxIndex:     log.TxIndex,
		}
	}
	return ret
}
