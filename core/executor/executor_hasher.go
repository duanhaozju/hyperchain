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
package executor

import (
	"hyperchain/common"
	er "hyperchain/core/errors"
	"hyperchain/core/ledger/codec"
	"hyperchain/core/ledger/codec/consensus"
	"hyperchain/core/ledger/codec/v1_2"
	"hyperchain/core/types"
)

// A tool use to calculate hash of transactions, receipts.
type Hasher struct {
	transactionCalculator interface{} // a batch of transactions calculator
	receiptCalculator     interface{} // a batch of receipts calculator
	transactionBuffer     [][]byte    // transaction buffer
	receiptBuffer         [][]byte    // receipt buffer
}

// initTransactionHashCalculator - reset transaction buffer.
func (executor *Executor) initTransactionHashCalculator() {
	executor.hasher.transactionBuffer = nil
}

// initReceiptHashCalculator - reset receipt buffer.
func (executor *Executor) initReceiptHashCalculator() {
	executor.hasher.receiptBuffer = nil
}

func (executor *Executor) initCalculator() {
	executor.initTransactionHashCalculator()
	executor.initReceiptHashCalculator()
}

// calculate a batch of transaction
// if flush flag is false, append serialized transaction data to the buffer,
// otherwise, make the whole buffer content as the input to generate a hash as a fingerprint.
func (executor *Executor) calculateTransactionsFingerprint(transaction *types.Transaction, flush bool) (common.Hash, error) {
	// short circuit if transaction pointer is empty
	if transaction == nil && flush == false {
		return common.Hash{}, er.EmptyPointerErr
	}

	if flush == false {
		var (
			data    []byte
			err     error
			encoder codec.Encoder
		)
		// Determine the serialization policy based on the data structure version tag.
		// The purpose is to be compatible with older version of the block chain data.
		switch string(transaction.Version) {
		case "1.0":
			fallthrough
		case "1.1":
			fallthrough
		case "1.2":
			encoder = new(v1_2.V1_2Encoder)
			data, err = encoder.EncodeTransaction(transaction)
		default:
			encoder = new(consensus.ConEncoder)
			data, err = encoder.EncodeTransaction(transaction)
		}
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] invalid receipt struct to marshal! error msg, ", executor.namespace, err.Error())
			return common.Hash{}, err
		}
		// put transaction to buffer temporarily
		executor.hasher.transactionBuffer = append(executor.hasher.transactionBuffer, data)
		return common.Hash{}, nil
	} else {
		// calculate hash together
		hash := executor.commonHash.Hash(executor.hasher.transactionBuffer)
		executor.hasher.transactionBuffer = nil
		return hash, nil
	}
	return common.Hash{}, nil
}

// calculate a batch of receipt
// if flush flag is false, append serialized receipt data to the buffer,
// otherwise, make the whole buffer content as the input to generate a hash as a fingerprint.
func (executor *Executor) calculateReceiptFingerprint(tx *types.Transaction, receipt *types.Receipt, flush bool) (common.Hash, error) {
	// short circuit if receipt pointer is empty
	if receipt == nil && flush == false {
		return common.Hash{}, er.EmptyPointerErr
	}
	if flush == false {
		var (
			data    []byte
			err     error
			encoder codec.Encoder
		)
		// Determine the serialization policy based on the data structure version tag.
		// The purpose is to be compatible with older version of the block chain data.
		switch string(tx.Version) {
		case "1.0":
			fallthrough
		case "1.1":
			fallthrough
		case "1.2":
			encoder = new(v1_2.V1_2Encoder)
			data, err = encoder.EncodeReceipt(receipt)
		default:
			encoder = new(consensus.ConEncoder)
			data, err = encoder.EncodeReceipt(receipt)
		}
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] invalid receipt struct to marshal! error msg, ", executor.namespace, err.Error())
			return common.Hash{}, err
		}
		// put transaction to buffer temporarily
		executor.hasher.receiptBuffer = append(executor.hasher.receiptBuffer, data)
		return common.Hash{}, nil
	} else {
		// calculate hash together
		hash := executor.commonHash.Hash(executor.hasher.receiptBuffer)
		executor.hasher.receiptBuffer = nil
		return hash, nil
	}
	return common.Hash{}, nil
}

// calculateValidationResultHash - calculate a hash to represent the validation result for consensus comparison.
func (executor *Executor) calculateValidationResultHash(merkleRoot, txRoot, receiptRoot []byte) common.Hash {
	return executor.commonHash.Hash([]interface{}{
		merkleRoot,
		txRoot,
		receiptRoot,
	})
}
