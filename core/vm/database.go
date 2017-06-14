package vm

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb/db"
)
// Database is a EVM database for full state querying.
type Database interface {
	GetAccount(common.Address) Account
	CreateAccount(common.Address) Account

	AddBalance(common.Address, *big.Int)
	GetBalance(common.Address) *big.Int

	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)

	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)

	GetStatus(common.Address) int
	SetStatus(common.Address, int)

	AddDeployedContract(common.Address, common.Address)
	GetDeployedContract(common.Address) []string

	SetCreator(common.Address, common.Address)
	GetCreator(common.Address) common.Address

	SetCreateTime(common.Address, uint64)
	GetCreateTime(common.Address) uint64

	AddRefund(*big.Int)
	GetRefund() *big.Int

	GetState(common.Address, common.Hash) (bool, common.Hash)
	SetState(common.Address, common.Hash, common.Hash, int32)

	Delete(common.Address) bool
	Exist(common.Address) bool
	IsDeleted(common.Address) bool

	// Log
	StartRecord(common.Hash, common.Hash, int)
	AddLog(log *types.Log)
	GetLogs(hash common.Hash) types.Logs
	// Dump and Load
	Snapshot() interface{}
	RevertToSnapshot(interface{})
	RevertToJournal(uint64, uint64, []byte, db.Batch) error
	// Reset statuso
	Purge()

	Commit() (common.Hash, error)
	Reset() error
	// Query
	GetAccounts() map[string]Account
	Dump() []byte
	GetTree() interface{}
	// Atomic Related
	MarkProcessStart(uint64)
	MarkProcessFinish(uint64)

	FetchBatch(seqNo uint64) db.Batch
	DeleteBatch(seqNo uint64)
	MakeArchive(uint64)
	ShowArchive(common.Address, string) map[string]map[string]string
}

