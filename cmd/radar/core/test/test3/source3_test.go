package test

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/cmd/radar/contract"
	"github.com/hyperchain/hyperchain/cmd/radar/core/api"
	"github.com/hyperchain/hyperchain/cmd/radar/core/test"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"testing"
)

func TestSource3(t *testing.T) {
	db, err := leveldb.OpenFile("blockchain", nil)
	defer db.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	var originKeyOfMaps map[string]map[string][][]string
	originKeyOfMaps = make(map[string]map[string][][]string)
	content, err := ioutil.ReadFile("key")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var temp1 contract.Keys
	err = json.Unmarshal(content, &temp1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	originKeyOfMaps[temp1.ContractName] = temp1.Key
	res, err := api.GetResult("source3.solc", db, "0346de490a2ebd37e1f6f216b1376c60f19cb9d6", originKeyOfMaps)

	var rightResult map[string][]string
	rightResult = make(map[string][]string)
	var temp = []string{
		"user={id:1,age:[3,4],desc:[\"b\",\"aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffgggggggggg\"]}",
		"temp=[\"b\",\"aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffgggggggggg\"]",
		"users=[{id:5,age:[7,8],desc:[\"b\",\"aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffgggggggggg\"]},{id:0,age:[0,0],desc:[]}]",
	}
	rightResult["Demo"] = temp

	if err != nil {
		t.Error(err.Error())
	} else {
		if test.Judge(res, rightResult) == false {
			t.Error("sorry, not equal!")
		}
	}

}
