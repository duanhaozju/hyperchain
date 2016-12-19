package bucket

import (
	"hyperchain/hyperdb"
)
type rawKey []byte

// TODO test
func fetchDataNodeFromDB(dataKey *dataKey) (*dataNode, error) {
	db, err := hyperdb.GetLDBDatabase()
	nodeBytes, err := db.Get(dataKey.getEncodedBytes())
	if err != nil {
		return nil, err
	}
	if nodeBytes == nil {
		logger.Debug("nodeBytes from db is nil")
	} else if len(nodeBytes) == 0 {
		logger.Debug("nodeBytes from db is an empty array")
	}
	// key does not exist
	if nodeBytes == nil {
		return nil, nil
	}
	return unmarshalDataNode(dataKey, nodeBytes), nil
}

// TODO test
func fetchBucketNodeFromDB(treePrefix string,bucketKey *bucketKey) (*bucketNode, error) {
	db,_ := hyperdb.GetLDBDatabase()
	nodeKey := bucketKey.getEncodedBytes(treePrefix)
	nodeBytes, err := db.Get(nodeKey)
	if err.Error() == "leveldb: not found"{
		return nil,nil
	}
	if err != nil {
		return nil, err
	}
	if nodeBytes == nil {
		return nil, nil
	}
	return unmarshalBucketNode(bucketKey, nodeBytes), nil
}


// TODO it need to be tested
func fetchDataNodesFromDBFor(treePrefix string,bucketKey *bucketKey) (dataNodes, error) {
	db,_ := hyperdb.GetLDBDatabase()
	iter := db.NewIterator()

	minimumDataKeyBytes := minimumPossibleDataKeyBytesFor(bucketKey,treePrefix)

	var dataNodes dataNodes
	iter.Seek(minimumDataKeyBytes)

	for ; iter.Valid(); iter.Next() {
		keyBytes := iter.Key()
		valueBytes := iter.Value()

		dataKey := newDataKeyFromEncodedBytes(keyBytes)
		logger.Debugf("Retrieved data key [%s] from DB for bucket [%s]", dataKey, bucketKey)
		if !dataKey.getBucketKey().equals(bucketKey) {
			logger.Debugf("Data key [%s] from DB does not belong to bucket = [%s]. Stopping further iteration and returning results [%v]", dataKey, bucketKey, dataNodes)
			return dataNodes, nil
		}
		dataNode := unmarshalDataNode(dataKey, valueBytes)
		logger.Debugf("Data node [%s] from DB belongs to bucket = [%s]. Including the key in results...", dataNode, bucketKey)
		dataNodes = append(dataNodes, dataNode)
	}
	logger.Debugf("Returning results [%v]", dataNodes)
	return dataNodes, nil
}
