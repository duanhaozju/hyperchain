package buckettree

import (
	"hyperchain/hyperdb"
)
type rawKey []byte

func fetchDataNodeFromDB(dataKey *dataKey) (*dataNode, error) {
	db, err := hyperdb.GetLDBDatabase()
	// TODO is the key ok?
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
func fetchBucketNodeFromDB(accountID string,bucketKey *bucketKey) (*bucketNode, error) {
	db,_ := hyperdb.GetLDBDatabase()
	nodeKey := append(accountID,bucketKey.getEncodedBytes()...)
	nodeBytes, err := db.Get(nodeKey)
	if err != nil {
		return nil, err
	}
	if nodeBytes == nil {
		return nil, nil
	}
	return unmarshalBucketNode(bucketKey, nodeBytes), nil
}


// TODO how to featch datanodes iterate
func fetchDataNodesFromDBFor(bucketKey *bucketKey) (dataNodes, error) {
	logger.Debugf("Fetching from DB data nodes for bucket [%s]", bucketKey)
	//db,_ := hyperdb.GetLDBDatabase()
	return nil, nil
}
