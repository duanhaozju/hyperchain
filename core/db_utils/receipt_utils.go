package db_utils

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
)

// GetReceipt returns a receipt by hash
func GetReceipt(namespace string, txHash common.Hash) *types.ReceiptTrans {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil
	}
	data, _ := db.Get(append(ReceiptsPrefix, txHash[:]...))
	if len(data) == 0 {
		return nil
	}
	var receiptWrapper types.ReceiptWrapper
	err = proto.Unmarshal(data, &receiptWrapper)
	if err != nil {
		logger.Errorf("GetReceipt err:", err)
		return nil
	}
	var receipt types.Receipt
	err = proto.Unmarshal(receiptWrapper.Receipt, &receipt)
	if err != nil {
		logger.Errorf("GetReceipt err:", err)
		return nil
	}
	return receipt.ToReceiptTrans()
}

// Persist receipt content to a batch, KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false
func PersistReceipt(batch db.Batch, receipt *types.Receipt, flush bool, sync bool) (error, []byte) {
	// check pointer value
	if receipt == nil || batch == nil {
		return EmptyPointerErr, nil
	}

	err, data := encapsulateReceipt(receipt)
	if err != nil {
		logger.Error("wrapper receipt failed.")
		return err, nil
	}
	if err := batch.Put(append(ReceiptsPrefix, receipt.TxHash...), data); err != nil {
		logger.Error("Put receipt data into database failed! error msg, ", err.Error())
		return err, nil
	}
	// flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

// encapsulateReceipt - encapsulate receipt with a wrapper for specify receipt structure version.
func encapsulateReceipt(receipt *types.Receipt) (error, []byte) {
	if receipt == nil {
		return EmptyPointerErr, nil
	}
	receipt.Version = []byte(ReceiptVersion)
	data, err := proto.Marshal(receipt)
	if err != nil {
		logger.Error("Invalid receipt struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	wrapper := &types.ReceiptWrapper{
		ReceiptVersion: []byte(ReceiptVersion),
		Receipt:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err != nil {
		logger.Error("Invalid receipt struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	return nil, data
}

// DeleteReceipt - delete receipt via tx hash.
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

// GetMarshalReceipt - marshal receipt with a specify receipt structure version.
func GetMarshalReceipt(receipt *types.Receipt) (error, []byte) {
	if receipt == nil {
		return EmptyPointerErr, nil
	}
	receipt.Version = []byte(ReceiptVersion)
	data, err := proto.Marshal(receipt)
	if err != nil {
		logger.Error("Invalid receipt struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	return nil, data
}
