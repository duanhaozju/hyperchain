package utils

import (
	"os"
	"strings"
	"path/filepath"
	"hyperchain/crypto/sha3"
	"strconv"
	"hyperchain/common"
)

func GetProjectPath() string{
	gopath := os.Getenv("GOPATH")
	if strings.Contains(gopath,":"){
		gopath = strings.Split(gopath,":")[0]
	}
	return gopath + "/src/hyperchain"
}

func GetCurrentDirectory() (string,error){
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "",err
	}
	return strings.Replace(dir, "\\", "/", -1),nil

}

func GetPeerHash(namespace string,id int) string{
	hasher := sha3.NewKeccak256()
	ids := strconv.Itoa(id)
	hasher.Write([]byte(namespace+ids))
	return common.ToHex(hasher.Sum(nil))
}
