package hyperencrypt

import (
	"encoding/json"
	"hyperchain-alpha/common"
	"hyperchain-alpha/hypercrypto/sha3"

)


func Fix32Hash(x interface{}) (h common.Hash) {
	serialize_data,err := json.Marshal(x)
	if err!=nil{
		panic(err)
	}
	hw := sha3.NewKeccak256()
	hw.Write(serialize_data)
	hw.Sum(h[:0])
	return h
}
