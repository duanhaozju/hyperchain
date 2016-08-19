package encrypt

import "crypto/sha256"
//定义hash算法确保全局HASH结果一致
func GetHash(toHashString []byte) []byte{
	h := sha256.New() 	// 实例化SHA256容器
	h.Write(toHashString)
	signhash := h.Sum(nil)
	return signhash
}
