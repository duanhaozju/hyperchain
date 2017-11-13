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
package bucket

import (
	"errors"
	hdb "github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
)

const (
	bucketPrefix     = "DataNodes"
	merkleNodePrefix = "BucketNode"
)

// PersistMerkleNode persists serialized merkle node to db writer.
// if the node's content is empty, it will be deleted from db.
// Otherwise, updation is omitted.
func PersistMerkleNode(prefix string, node *MerkleNode, dbw hdb.Batch) error {
	if node.deleted {
		return dbw.Delete(append([]byte(merkleNodePrefix), append([]byte(prefix), node.pos.encode()...)...))
	} else {
		return dbw.Put(append([]byte(merkleNodePrefix), append([]byte(prefix), node.pos.encode()...)...), node.encode())
	}
}

// RetrieveMerkleNode returns a merkle node if exists in database.
// The key is combined with rule:
//     key = <merkleNodePrefix> + <tree prefix> + <encoded postion>
func RetrieveMerkleNode(log *logging.Logger, db hdb.Database, prefix string, pos *Position) (*MerkleNode, error) {
	dbKey := append([]byte(merkleNodePrefix), []byte(prefix)...)
	dbKey = append(dbKey, pos.encode()...)
	blob, err := db.Get(dbKey)
	if err != nil {
		if err.Error() == hdb.DB_NOT_FOUND.Error() {
			return nil, nil
		}
		return nil, err
	}
	if blob == nil {
		return nil, nil
	}
	return decodeMerkleNode(log, pos, blob), nil
}

// PersistBucket persists serialized bucket data to db writer.
// if the bucket's content is empty, it will be deleted from db.
// Otherwise, updation is omitted.
func PersistBucket(prefix string, bucket Bucket, pos *Position, dbw hdb.Batch) error {
	if bucket == nil || len(bucket) == 0 {
		return dbw.Delete(append([]byte(prefix), append([]byte(bucketPrefix), pos.encode()...)...))
	} else {
		return dbw.Put(append([]byte(prefix), append([]byte(bucketPrefix), pos.encode()...)...), bucket.Encode())
	}
}

// RetrieveHashBucket returns a set of hash nodes belong to same bucket if exists in database.
// The key is combined with rule:
//     key = <tree prefix> + <hash node prefix> + <encoded postion>
func RetrieveBucket(log *logging.Logger, db hdb.Database, treePrefix string, pos *Position) (bucket Bucket, err error) {
	blob, err := db.Get(append([]byte(treePrefix), append([]byte(bucketPrefix), pos.encode()...)...))
	if err != nil {
		if err.Error() == hdb.DB_NOT_FOUND.Error() {
			return bucket, nil
		}
		panic("Retrieve hash bucket from db failed")
	}
	if blob == nil || len(blob) <= len(bucketPrefix)+1 {
		return nil, errors.New("Data is nil")
	}

	if err := DecodeBucket(treePrefix, pos, blob, &bucket); err != nil {
		panic("Get bucketKey error from db error ")
	}

	return bucket, nil
}
