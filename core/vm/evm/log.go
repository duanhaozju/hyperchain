//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"encoding/json"
	"fmt"
	"hyperchain/common"
	"hyperchain/core/vm"
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
	Type        string      `json:"Type"`
}

// assign block number as 0 temporarily
// because the blcok number in env is a seqNo actually
// primary's seqNo may not equal to other's
// correctly block number and block hash will be assigned in the commit phase
func NewLog(address common.Address, topics []common.Hash, data []byte, number uint64) *Log {
	return &Log{Address: address, Topics: topics, Data: data, BlockNumber: number, Type: "evm"}
}

func (l *Log) String() string {
	return fmt.Sprintf(`{address: %x, topics: %x, data: %x, txhash: %x, txIndex: %d, blockHash: %x, blockNumber: %d, index: %d}`, l.Address, l.Topics, l.Data, l.TxHash, l.TxIndex, l.BlockHash, l.BlockNumber, l.Index)
}

/*
	Attribute access
 */

func (l *Log) GetAttribute(t int) interface{} {
	switch t {
	case vm.LogAttr_Address:
		return l.Address
	case vm.LogAttr_Topics:
		return l.Topics
	case vm.LogAttr_Data:
		return l.Data
	case vm.LogAttr_BlockNumber:
		return l.BlockNumber
	case vm.LogAttr_BlockHash:
		return l.BlockHash
	case vm.LogAttr_TxHash:
		return l.TxHash
	case vm.LogAttr_TxIndex:
		return l.TxIndex
	case vm.LogAttr_Index:
		return l.Index
	case vm.LogAttr_Type:
		return l.Type
	default:
		return nil
	}
}
/*
	Attribute setter
 */
func (l *Log) SetAttribute(t int, v interface{}) {
	switch t {
	case vm.LogAttr_Address:
		tmp := v.(common.Address)
		l.Address = tmp
	case vm.LogAttr_Topics:
		tmp := v.([]common.Hash)
		l.Topics = tmp
	case vm.LogAttr_Data:
		tmp := v.([]byte)
		l.Data = tmp
	case vm.LogAttr_BlockNumber:
		tmp := v.(uint64)
		l.BlockNumber = tmp
	case vm.LogAttr_BlockHash:
		tmp := v.(common.Hash)
		l.BlockHash = tmp
	case vm.LogAttr_TxHash:
		tmp := v.(common.Hash)
		l.TxHash = tmp
	case vm.LogAttr_TxIndex:
		tmp := v.(uint)
		l.TxIndex = tmp
	case vm.LogAttr_Index:
		tmp := v.(uint)
		l.Index = tmp
	case vm.LogAttr_Type:
		tmp := v.(string)
		l.Type = tmp
	}
}
/*
	Mist
 */
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

func ReAssign(buf []byte, blockNumber uint64, blockHash common.Hash) ([]byte, error) {
	logs, err := DecodeLogs(buf)
	if err != nil {
		return nil, err
	}
	for _, log := range logs {
		log.BlockNumber = blockNumber
		log.BlockHash = blockHash
	}
	return logs.EncodeLogs()
}

type LogTrans struct {
	Address     string
	Topics      []string
	Data        string
	BlockNumber uint64
	BlockHash   string
	TxHash      string
	TxIndex     uint
	Index       uint
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
			Data:        common.BytesToHash(log.Data).Hex(),
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

// LogForStorage is a wrapper around a Log that flattens and parses the entire
// content of a log, as opposed to only the consensus fields originally (by hiding
// the rlp interface methods).
type LogForStorage Log
