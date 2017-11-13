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

package chain

import (
	"errors"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/hyperdb"

	"github.com/op/go-logging"
    "hyperchain/common/service/server"
)

var (
	ErrEmptyPointer      = errors.New("nil pointer")
	ErrNotFindTxMeta     = errors.New("not find tx meta")
	ErrNotFindBlock      = errors.New("not find block")
	ErrOutOfSliceRange   = errors.New("out of slice(transactions in block) range")
	ErrFromSmallerThanTo = errors.New("from smaller than to")
)

const (
	BlockVersion       = "1.3"
	ReceiptVersion     = "1.3"
	TransactionVersion = "1.3"
)

var (
	ChainKey = []byte("chain-key-")

	// Data item prefixes
	ReceiptsPrefix  = []byte("receipts-")
	InvalidTxPrefix = []byte("invalidTx-")
	BlockPrefix     = []byte("block-")
	BlockNumPrefix  = []byte("blockNum-")
	TxMetaSuffix    = []byte{0x01}

	JournalPrefix  = []byte("-journal")
	SnapshotPrefix = []byte("-snapshot")
)

// InitDBForNamespace inits the database with given namespace
func InitDBForNamespace(conf *common.Config, namespace string, is *server.InternalServer) error {
	err := hyperdb.InitDatabase(conf, namespace)
	if err != nil {
		return err
	}
    if conf.GetBool(common.EXECUTOR_EMBEDDED) {
        InitializeChain(namespace)
    } else {
        logger(namespace).Criticalf("InitializeRemoteChain")
        InitializeRemoteChain(is, namespace)
    }
	return err
}

func InitExecutorDBForNamespace(conf *common.Config, namespace string) error {
    err := hyperdb.InitBlockDatabase(conf, namespace)
    if err != nil {
        return err
    }
    err = hyperdb.InitArchiveDatabase(conf, namespace)
    if err != nil {
        return err
    }
    InitializeChain(namespace)
    return err
}


// logger returns the logger with given namespace
func logger(namespace string) *logging.Logger {
	return common.GetLogger(namespace, "core/ledger")
}
