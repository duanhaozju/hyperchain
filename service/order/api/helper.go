package api

import (
	"time"
	"github.com/hyperchain/hyperchain/common"
)

var (
	DEFAULT_GAS       int64 = 100000000
	DEFAULT_GAS_PRICE int64 = 10000
)

// prepareExcute checks if arguments are valid.
// 0 value for txType means sending normal transaction, 1 means deploying contract,
// 2 means invoking contract, 3 means signing hash, 4 means maintaining contract.
func prepareExcute(args SendTxArgs, txType int) (SendTxArgs, error) {
	if args.From.Hex() == (common.Address{}).Hex() {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "address `from` is invalid"}
	}
	if (txType == 0 || txType == 2 || txType == 4) && args.To == nil {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "address `to` is invalid"}
	}
	if args.Timestamp <= 0 || (5*int64(time.Minute)+time.Now().UnixNano()) < args.Timestamp {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "`timestamp` is invalid"}
	}
	if txType != 3 && args.Signature == "" {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "missing params `signature`"}
	}
	if args.Nonce <= 0 {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "`nonce` is invalid"}
	}
	if txType == 4 && args.Opcode == 1 && (args.Payload == "" || args.Payload == "0x") {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "contract code is empty"}
	}
	if txType == 1 && (args.Payload == "" || args.Payload == "0x") {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "contract code is empty"}
	}
	if args.SnapshotId != "" && args.Simulate != true {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "can not query history ledger without `simulate` mode"}
	}
	if args.Timestamp+time.Duration(24*time.Hour).Nanoseconds() < time.Now().UnixNano() {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "transaction out of date"}
	}

	return args, nil
}

