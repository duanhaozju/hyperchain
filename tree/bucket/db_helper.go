package bucket

import (
	"hyperchain/hyperdb"
)
type rawKey []byte

// TODO test
func fetchDataNodeFromDB(dataKey *DataKey) (*DataNode, error) {
	db, err := hyperdb.GetLDBDatabase()
	nodeBytes, err := db.Get(dataKey.getEncodedBytes())
	nodeBytes = append([]byte(DataNodePrefix),nodeBytes...)
	if err != nil {
		return nil, err
	}
	if nodeBytes == nil {
		log.Debug("nodeBytes from db is nil")
	} else if len(nodeBytes) == 0 {
		log.Debug("nodeBytes from db is an empty array")
	}
	// key does not exist
	if nodeBytes == nil {
		return nil, nil
	}
	return unmarshalDataNode(dataKey, nodeBytes), nil
}

// TODO test
func fetchBucketNodeFromDB(treePrefix string,bucketKey *BucketKey) (*BucketNode, error) {
	db,_ := hyperdb.GetLDBDatabase()
	//nodeKey := bucketKey.getEncodedBytes(treePrefix)
	nodeKey := append([]byte(BucketNodePrefix),[]byte(treePrefix)...)
	nodeKey = append(nodeKey,bucketKey.getEncodedBytes()...)
	nodeBytes, err := db.Get(nodeKey)

	if err != nil {
		if err.Error() == "leveldb: not found"{
			return nil,nil
		}
		return nil, err
	}
	if nodeBytes == nil {
		return nil, nil
	}
	return unmarshalBucketNode(bucketKey, nodeBytes), nil
}

// TODO it need to be tested
func fetchDataNodesFromDB(treePrefix string) (DataNodes, error) {
	db,_ := hyperdb.GetLDBDatabase()

	minimumDataKeyBytes := append([]byte(DataNodePrefix))

	var dataNodes DataNodes
	// IMPORTANT return value obtained by iterator is sorted
	iter := db.NewIteratorWithPrefix(minimumDataKeyBytes)
	for iter.Next() {
		keyBytes := iter.Key()
		valueBytes := iter.Value()

		keyBytes = keyBytes[len(DataNodePrefix):]
		dataKey := newDataKeyFromEncodedBytes(keyBytes)
		dataNode := unmarshalDataNode(dataKey, valueBytes)
		log.Debugf("Data node [%s] from DB belongs to bucket = [%s]. Including the key in results...", dataNode, dataNode.dataKey.bucketKey)
		dataNodes = append(dataNodes, dataNode)
	}
	log.Debugf("Returning results [%v]", dataNodes)
	return dataNodes, nil
}



// TODO it need to be tested
func fetchDataNodesFromDBByBucketKey(treePrefix string,bucketKey *BucketKey) (dataNodes DataNodes, err error) {
	db,_ := hyperdb.GetLDBDatabase()
	dataNodesValue,err := db.Get(append([]byte(DataNodesPrefix),append([]byte(treePrefix),bucketKey.getEncodedBytes()...)...))
	if err != nil{
		if err.Error() == ErrNotFound.Error(){
			return dataNodes,nil
		}
		log.Errorf("DB get bucketKey ",bucketKey,"error is",err)
		panic("Get bucketKey error from db error ")
	}
	err = UnmarshalDataNodes(bucketKey,dataNodesValue,dataNodes)
	if err != nil{
		log.Errorf("Marshal dataNodesValue error",err)
		panic("Get bucketKey error from db error ")
	}
	return dataNodes, nil
}
