package core

import (


	"hyperchain-alpha/common"
	"hyperchain-alpha/hyperdb"
	"github.com/ethereum/go-ethereum/logger/glog"
)




// WriteHeadBlockHash stores the head block's hash.
func WriteBlockHash(db hyperdb.Database, hash common.Hash) error {
	if err := db.Put(hash, hash.Bytes()); err != nil {
		glog.Fatalf("failed to store last block's hash into database: %v", err)
		return err
	}
	return nil
}
