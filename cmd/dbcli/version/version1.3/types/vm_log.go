package version1_3

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
)

const (
	LogVmType_EVM = iota
	LogVmType_JVM
)

type LogVmType int

func (vmType LogVmType) String() string {
	switch vmType {
	case LogVmType_EVM:
		return "EVM"
	case LogVmType_JVM:
		return "JVM"
	default:
		return ""
	}
}

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

// LogForStorage is a wrapper around a Log that flattens and parses the entire
// content of a log, as opposed to only the consensus fields originally (by hiding
// the rlp interface methods).
type LogForStorage Log
