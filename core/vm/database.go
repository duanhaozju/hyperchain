package vm

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb/db"
	"math/big"
)

type Database interface {
	GetAccount(common.Address) Account
	CreateAccount(common.Address) Account

	AddBalance(common.Address, *big.Int)
	GetBalance(common.Address) *big.Int

	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)

	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)

	GetCodeHash(common.Address) common.Hash

	GetStatus(common.Address) int
	SetStatus(common.Address, int)

	AddDeployedContract(common.Address, common.Address)
	GetDeployedContract(common.Address) []string

	SetCreator(common.Address, common.Address)
	GetCreator(common.Address) common.Address

	SetCreateTime(common.Address, uint64)
	GetCreateTime(common.Address) uint64

	GetState(common.Address, common.Hash) (bool, []byte)
	SetState(common.Address, common.Hash, []byte, int32)

	Delete(common.Address) bool
	Exist(common.Address) bool

	// Log
	StartRecord(common.Hash, common.Hash, int)
	AddLog(log *types.Log)
	GetLogs(hash common.Hash) types.Logs
	// Dump and Load
	Snapshot() interface{}
	RevertToSnapshot(interface{})
	RevertToJournal(uint64, uint64, []byte, db.Batch) error
	// Clear out all extemporaneous status
	Purge()

	Commit() (common.Hash, error)
	RecomputeCryptoHash() (common.Hash, error)
	ResetToTarget(uint64, common.Hash)
	Reset() error
	// Query
	GetAccounts() map[string]Account
	Dump() []byte
	GetCurrentTxHash() common.Hash
	NewIterator(common.Address, *IterRange) (Iterator, error)

	// Atomic Related
	MarkProcessStart(uint64)
	MarkProcessFinish(uint64)

	FetchBatch(seqNo uint64, typ int) db.Batch
	DeleteBatch(seqNo uint64)
	MakeArchive(uint64)
	Merge(db.Database, db.Batch, common.Hash) error
}

type Iterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Release()
}

type IterRange struct {
	Start *common.Hash
	Limit *common.Hash
}

func BytesPrefix(prefix []byte) *IterRange {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	startH := common.BytesToRightPaddingHash(prefix)
	limitH := common.BytesToRightPaddingHash(limit)
	return &IterRange{
		Start: &startH,
		Limit: &limitH,
	}
}
