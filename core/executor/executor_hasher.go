package executor

import (
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/db_utils/codec/v1.2"
	er "hyperchain/core/errors"
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
			data []byte
			err  error
		)
		// Determine the serialization policy based on the data structure version tag.
		// The purpose is to be compatible with older version of the block chain data.
		switch string(transaction.Version) {
		case "1.0":
			fallthrough
		case "1.1":
			fallthrough
		case "1.2":
			data, err = v1_2.EncodeTransaction(transaction)
		default:
			data, err = edb.EncodeTransaction(transaction)
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
			data []byte
			err  error
		)
		// Determine the serialization policy based on the data structure version tag.
		// The purpose is to be compatible with older version of the block chain data.
		switch string(tx.Version) {
		case "1.0":
			fallthrough
		case "1.1":
			fallthrough
		case "1.2":
			data, err = v1_2.EncodeReceipt(receipt)
		default:
			data, err = edb.EncodeReceipt(receipt)
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
