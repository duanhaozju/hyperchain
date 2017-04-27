package vm

import (
	"math/big"
	"hyperchain/hyperdb/db"
	"hyperchain/common"
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

	GetState(common.Address, common.Hash) (bool, []byte)
	SetState(common.Address, common.Hash, []byte, int32)

	Delete(common.Address) bool
	Exist(common.Address) bool
	IsDeleted(common.Address) bool

	// Log
	StartRecord(common.Hash, common.Hash, int)
	AddLog(log Log)
	GetLogs(hash common.Hash) Logs
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
	GetCurrentTxHash() common.Hash
	// Atomic Related
	MarkProcessStart(uint64)
	MarkProcessFinish(uint64)

	FetchBatch(seqNo uint64) db.Batch
	DeleteBatch(seqNo uint64)
	MakeArchive(uint64)
	ShowArchive(common.Address, string) map[string]map[string]string
}
