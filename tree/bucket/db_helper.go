package bucket

import (
	"hyperchain/hyperdb"
)
type rawKey []byte

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
func fetchDataNodesFromDBByBucketKey(treePrefix string,bucketKey *BucketKey) (dataNodes DataNodes, err error) {
	db,_ := hyperdb.GetLDBDatabase()
	dataNodesValue,err := db.Get(append([]byte(treePrefix),append([]byte(DataNodesPrefix),bucketKey.getEncodedBytes()...)...))
	if err != nil{
		if err.Error() == ErrNotFound.Error(){
			return dataNodes,nil
		}
		log.Errorf("DB get bucketKey ",bucketKey,"error is",err)
		panic("Get bucketKey error from db error ")
	}
	err = UnmarshalDataNodes(bucketKey,dataNodesValue,&dataNodes)
	if err != nil{
		log.Errorf("Marshal dataNodesValue error",err)
		panic("Get bucketKey error from db error ")
	}
	return dataNodes, nil
}
