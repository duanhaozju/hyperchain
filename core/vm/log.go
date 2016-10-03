package vm

import (
	"encoding/json"
	"fmt"
	"hyperchain/common"
)

type Log struct {
	// Consensus fields
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

func NewLog(address common.Address, topics []common.Hash, data []byte, number uint64) *Log {
	return &Log{Address: address, Topics: topics, Data: data, BlockNumber: number}
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
	Address     string
	Topics      []string
	Data        string
	BlockNumber uint64
	TxHash      string
	TxIndex     uint
	BlockHash   string
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
			Topics:      topics,
			BlockHash:   log.BlockHash.Hex(),
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
