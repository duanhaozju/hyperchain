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
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"

	"github.com/golang/protobuf/proto"
)

// GetReceipt gets the receipt(web format) with given txHash.
func GetReceipt(namespace string, txHash common.Hash) *types.ReceiptTrans {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil
	}
	receipt := getReceiptFunc(db, txHash)
	if receipt == nil {
		return nil
	}
	return receipt.ToReceiptTrans()
}

// GetReceipt gets the receipt with given txHash.
func GetRawReceipt(namespace string, txHash common.Hash) *types.Receipt {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil
	}
	return getReceiptFunc(db, txHash)
}

// PersistReceipt persists receipt into database using batch.
// KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false.
func PersistReceipt(batch db.Batch, receipt *types.Receipt, flush bool, sync bool, extra ...interface{}) ([]byte, error) {
	// Check pointer value
	if receipt == nil || batch == nil {
		return nil, ErrEmptyPointer
	}

	data, err := encapsulateReceipt(receipt, extra...)
	if err != nil {
		return nil, err
	}
	if err := batch.Put(append(ReceiptsPrefix, receipt.TxHash...), data); err != nil {
		return nil, err
	}
	// Flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write() //TODO: async
		}
	}
	return data, nil
}

// DeleteReceipt deletes the receipt with given receipt key.
func DeleteReceipt(batch db.Batch, key []byte, flush, sync bool) error {
	keyFact := append(ReceiptsPrefix, key...)
	err := batch.Delete(keyFact)
	if err != nil {
		return err
	}
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// encapsulateReceipt encapsulates receipt with a wrapper for specify receipt structure version.
func encapsulateReceipt(receipt *types.Receipt, extra ...interface{}) ([]byte, error) {
	version := ReceiptVersion
	if receipt == nil {
		return nil, ErrEmptyPointer
	}
	if len(extra) > 0 {
		// Parse version
		if tmp, ok := extra[0].(string); ok {
			version = tmp
		}
	}
	receipt.Version = []byte(version)
	data, err := proto.Marshal(receipt)
	if err != nil {
		return nil, err
	}
	wrapper := &types.ReceiptWrapper{
		ReceiptVersion: []byte(version),
		Receipt:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// getReceiptFunc is a help function, gets the receipt from database with given txHash.
func getReceiptFunc(db db.Database, txHash common.Hash) *types.Receipt {
	data, _ := db.Get(append(ReceiptsPrefix, txHash[:]...))
	if len(data) == 0 {
		return nil
	}
	var receiptWrapper types.ReceiptWrapper
	err := proto.Unmarshal(data, &receiptWrapper)
	if err != nil {
		return nil
	}
	var receipt types.Receipt
	err = proto.Unmarshal(receiptWrapper.Receipt, &receipt)
	if err != nil {
		return nil
	}
	return &receipt
}
