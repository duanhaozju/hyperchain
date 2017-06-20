package bucket

import (
	"errors"
	hdb "hyperchain/hyperdb/db"
	"hyperchain/common"
)

type rawKey []byte

// TODO test
func fetchBucketNodeFromDB(db hdb.Database, treePrefix string, bucketKey *BucketKey) (*BucketNode, error) {
	//nodeKey := bucketKey.getEncodedBytes(treePrefix)
	nodeKey := append([]byte(BucketNodePrefix), []byte(treePrefix)...)
	nodeKey = append(nodeKey, bucketKey.getEncodedBytes()...)
	nodeBytes, err := db.Get(nodeKey)

	if err != nil {
		if err.Error() == hdb.DB_NOT_FOUND.Error() {
			return nil, nil
		}
		return nil, err
	}
	if nodeBytes == nil {
		return nil, nil
	}
	return unmarshalBucketNode(bucketKey, nodeBytes), nil
}

// TODO it need to be tested
func fetchDataNodesFromDBByBucketKey(db hdb.Database, treePrefix string, bucketKey *BucketKey) (dataNodes DataNodes, err error) {
	dataNodesValue, err := db.Get(append([]byte(treePrefix), append([]byte(DataNodesPrefix), bucketKey.getEncodedBytes()...)...))
	log := common.GetLogger(db.Namespace(), "bucket")
	if err != nil {
		if err.Error() == hdb.DB_NOT_FOUND.Error() {
			return dataNodes, nil
		}
		log.Errorf("DB get bucketKey ", bucketKey, "error is", err)
		panic("Get bucketKey error from db error ")
	}
	if dataNodesValue == nil || len(dataNodesValue) <= len(DataNodesPrefix)+1 {
		return nil, errors.New("Data is nil")
	}

	err = UnmarshalDataNodes(treePrefix, bucketKey, dataNodesValue, &dataNodes)

	//if(bucketKey.level == 2 && bucketKey.bucketNumber == 13){
	//	log.Critical("writeBatch.get size is",dataNodes.Len())
	//	log.Critical("dataNodes marshal is",common.ToHex(dataNodes.Marshal()))
	//}

	if err != nil {
		log.Errorf("Marshal dataNodesValue error", err)
		panic("Get bucketKey error from db error ")
	}
	return dataNodes, nil
}
