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
	"bytes"
	"reflect"
	"testing"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/hyperdb/mdb"
)

func TestGetReceipt(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	receipt := types.Receipt{}
	receipt.TxHash = common.BytesToHash(receipt.TxHash).Bytes()
	PersistReceipt(db.NewBatch(), &receipt, true, true)
	dbReceipt := getReceiptFunc(db, common.BytesToHash(receipt.TxHash))
	if dbReceipt == nil || !reflect.DeepEqual(&receipt, dbReceipt) {
		t.Error("not exist in db or not deep equal exactly")
	}
}

func TestPersistReceiptWithVersion(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	receipt := types.Receipt{}
	receipt.TxHash = common.BytesToHash(receipt.TxHash).Bytes()
	PersistReceipt(db.NewBatch(), &receipt, true, true, "1.4")
	dbReceipt := getReceiptFunc(db, common.BytesToHash(receipt.TxHash))
	if dbReceipt == nil || bytes.Compare(dbReceipt.Version, []byte("1.4")) != 0 {
		t.Error("not exist in db or not deep equal exactly")
	}
}

func TestDeleteReceipt(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	receipt := types.Receipt{}
	receipt.TxHash = common.BytesToHash(receipt.TxHash).Bytes()
	PersistReceipt(db.NewBatch(), &receipt, true, true, "1.4")
	DeleteReceipt(db.NewBatch(), receipt.TxHash, true, true)

	if dbReceipt := getReceiptFunc(db, common.BytesToHash(receipt.TxHash)); dbReceipt != nil {
		t.Error("expect to be nil")
	}
}
