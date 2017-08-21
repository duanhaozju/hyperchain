//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import (
	"encoding/base64"
	"crypto/md5"
	"encoding/hex"
)


func byteToString(data []byte) (re string) {

	re = base64.StdEncoding.EncodeToString(data)
	return
}


func hash(batch *TxHashBatch) string {
	h := md5.New()
	for _, hash := range batch.TxHashList {
		h.Write([]byte(hash))
	}
	return hex.EncodeToString(h.Sum(nil))
}