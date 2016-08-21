package utils

import (
	"crypto/dsa"
	"hyperchain-alpha/encrypt"
	"path"
	"encoding/json"
	"io/ioutil"
	"fmt"
)
//公钥私钥生成工具类

type KeyPair struct{
	PriKey dsa.PrivateKey
	PubKey dsa.PublicKey
}

//账户
type Account map[string]KeyPair

type Accounts []Account

var usernames = []string{ "张三",
					"李四",
					"王五",
					"赵六",
					"钱七",
					}
const accountsFilePath = "./keystore.txt"

func GenKeypair(prefix string){
	var accounts Accounts
	//生成五个账户
	for _,username := range usernames{
		prikey :=  encrypt.GetPrivateKey()
		pubkey := encrypt.GetPublicKey(prikey)
		keypair := KeyPair{
			PriKey:prikey,
			PubKey:pubkey,
		}
		account := make(Account)
		account[username] = keypair
		accounts = append(accounts,account)
	}

	//2.写入文件
	accFilePath := path.Join(GetBasePath(),accountsFilePath)
	writeAccountsTOFile(accFilePath,accounts)
}

func writeAccountsTOFile(filepath string,accounts Accounts){
	fmt.Println("keystore路径",filepath)
	//for idx,account := range accounts{
	//
	//}
	//序列化
	dataString,err := json.Marshal(accounts)
	check(err)
	err= ioutil.WriteFile(filepath,dataString,0664)
	check(err)

}

func GetAccount()(Accounts,error){
	accFilePath := path.Join(GetBasePath(),accountsFilePath)
	accountData := readAccountFromFile(accFilePath)
	var accounts Accounts
	err := json.Unmarshal(accountData,&accounts)
	return accounts,err
}

func readAccountFromFile(filePath string) []byte{
	data,_ := ioutil.ReadFile(filePath)
	return data
}

func check(e error){
	if e != nil {
		panic(e)
	}
}

func GetGodAccount()Accounts{
	accFilePath := path.Join(GetBasePath(),"./godaccount")
	accountData := readAccountFromFile(accFilePath)
	var accounts Accounts
	json.Unmarshal(accountData,&accounts)
	return accounts
}