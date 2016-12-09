package buckettree

import (
)

func fetchDataNodeFromDB(dataKey *dataKey) (*dataNode, error) {
	/*
	openchainDB := db.GetDBHandle()
	nodeBytes, err := openchainDB.GetFromStateCF(dataKey.getEncodedBytes())
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
	return unmarshalDataNode(dataKey, nodeBytes), nil*/
	return nil,nil
}

// 1.
func fetchBucketNodeFromDB(bucketKey *bucketKey) (*bucketNode, error) {
	/*openchainDB := db.GetDBHandle()
	nodeBytes, err := openchainDB.GetFromStateCF(bucketKey.getEncodedBytes())
	if err != nil {
		return nil, err
	}
	if nodeBytes == nil {
		return nil, nil
	}
	return unmarshalBucketNode(bucketKey, nodeBytes), nil*/
	return nil,nil
}

type rawKey []byte

func fetchDataNodesFromDBFor(bucketKey *bucketKey) (dataNodes, error) {
	/*logger.Debugf("Fetching from DB data nodes for bucket [%s]", bucketKey)
	openchainDB := db.GetDBHandle()
	itr := openchainDB.GetStateCFIterator()
	defer itr.Close()
	minimumDataKeyBytes := minimumPossibleDataKeyBytesFor(bucketKey)

	var dataNodes dataNodes

	itr.Seek(minimumDataKeyBytes)

	for ; itr.Valid(); itr.Next() {

		// making a copy of key-value bytes because, underlying key bytes are reused by itr.
		// no need to free slices as iterator frees memory when closed.
		keyBytes := statemgmt.Copy(itr.Key().Data())
		valueBytes := statemgmt.Copy(itr.Value().Data())

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
	return dataNodes, nil*/
	return nil,nil
}
