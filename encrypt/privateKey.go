package encrypt

import (
	"crypto/rand"
	"crypto/dsa"
	"fmt"
	"os"
	"encoding/json"
	"encoding/hex"
)

func GetPrivateKey()  dsa.PrivateKey{
	params := new(dsa.Parameters)
	// see http://golang.org/pkg/crypto/dsa/#ParameterSizes
	if err := dsa.GenerateParameters(params, rand.Reader, dsa.L1024N160); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	privatekey := new(dsa.PrivateKey)
	privatekey.PublicKey.Parameters = *params
	dsa.GenerateKey(privatekey, rand.Reader) // this generates a public & private key pair
	return *privatekey
}
//编码私钥并进行存储
func EncodePrivateKey(privateKey *dsa.PublicKey)string{
	data,_ := json.Marshal(*privateKey)
	hexdata := hex.EncodeToString(data)
	return hexdata
}
//解码已经编码的私钥
func DecodePrivateKey(privateKeyString string) (p dsa.PrivateKey){
	dehexdata,err:= hex.DecodeString(privateKeyString)
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
