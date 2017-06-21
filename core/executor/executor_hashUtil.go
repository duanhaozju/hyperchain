package executor

import (
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	er "hyperchain/core/errors"
)

type ExecutorHashUtil struct {
	transactionCalculator interface{} // a batch of transactions calculator
	receiptCalculator     interface{} // a batch of receipts calculator
	transactionBuffer     [][]byte    // transaction buffer
	receiptBuffer         [][]byte    // receipt buffer
}

func (executor *Executor) initTransactionHashCalculator() {
	executor.hashUtils.transactionBuffer = nil
}

func (executor *Executor) initReceiptHashCalculator() {
	executor.hashUtils.receiptBuffer = nil
}

func (executor *Executor) initCalculator() {
	executor.initTransactionHashCalculator()
	executor.initReceiptHashCalculator()
}

// calculate a batch of transaction
func (executor *Executor) calculateTransactionsFingerprint(transaction *types.Transaction, flush bool) (common.Hash, error) {
	if transaction == nil && flush == false {
		return common.Hash{}, er.EmptyPointerErr
	}
	if flush == false {
		err, data := edb.GetMarshalTransaction(transaction)
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] invalid transaction struct to marshal! error msg, ", executor.namespace, err.Error())
			return common.Hash{}, err
		}
		// put transaction to buffer temporarily
		executor.hashUtils.transactionBuffer = append(executor.hashUtils.transactionBuffer, data)
		return common.Hash{}, nil
	} else {
		// calculate hash together
		hash := executor.commonHash.Hash(executor.hashUtils.transactionBuffer)
		executor.hashUtils.transactionBuffer = nil
		return hash, nil
	}
	return common.Hash{}, nil
}

// calculate a batch of receipt
func (executor *Executor) calculateReceiptFingerprint(receipt *types.Receipt, flush bool) (common.Hash, error) {
	// 1. marshal receipt to byte slice
	if receipt == nil && flush == false {
		return common.Hash{}, er.EmptyPointerErr
	}
	if flush == false {
		err, data := edb.GetMarshalReceipt(receipt)
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] invalid receipt struct to marshal! error msg, ", executor.namespace, err.Error())
			return common.Hash{}, err
		}
		// put transaction to buffer temporarily
		executor.hashUtils.receiptBuffer = append(executor.hashUtils.receiptBuffer, data)
		return common.Hash{}, nil
	} else {
		// calculate hash together
		hash := executor.commonHash.Hash(executor.hashUtils.receiptBuffer)
		executor.hashUtils.receiptBuffer = nil
		return hash, nil
	}
	return common.Hash{}, nil
}

// calculateValidationResultHash - calculate a hash to represent a validation result for comparison.
func (executor *Executor) calculateValidationResultHash(merkleRoot, txRoot, receiptRoot []byte) common.Hash {
	return executor.commonHash.Hash([]interface{}{
		merkleRoot,
		txRoot,
		receiptRoot,
	})
}
