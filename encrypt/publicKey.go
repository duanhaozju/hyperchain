package encrypt

import (
	"crypto/dsa"
	"encoding/json"
	"encoding/hex"
	"fmt"
)

//得到公钥
func GetPublicKey(privateKey dsa.PrivateKey) dsa.PublicKey {
	return privateKey.PublicKey
}

//编码私钥并进行存储
func EncodePublicKey(publickey *dsa.PublicKey)string{
	data,_ := json.Marshal(*publickey)
	hexdata := hex.EncodeToString(data)
	return hexdata
}
//解码已经编码的私钥
func DecodePublicKey(publicKeyString string) (p dsa.PublicKey){
	dehexdata,err:= hex.DecodeString(publicKeyString)
	if err!=nil{
		fmt.Errorf("Error: %v",err)
		return
	}
	err = json.Unmarshal(dehexdata, &p)
	if err!=nil{
		fmt.Errorf("Error: %v",err)
		return
	}
	return p
}

