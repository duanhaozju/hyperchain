package test

import (
	"testing"
	"hyperchain/cmd/radar/core/api"
	"encoding/json"
	"io/ioutil"
	"github.com/syndtr/goleveldb/leveldb"
	"hyperchain/cmd/radar/contract"
	"hyperchain/cmd/radar/core/test"
	"fmt"
)

func TestSource2 (t *testing.T) {
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
	res, err := api.GetResult("source2.solc", db, "0346de490a2ebd37e1f6f216b1376c60f19cb9d6", originKeyOfMaps)

	var rightResult map[string][]string
	rightResult = make(map[string][]string)
	var temp = []string{
		"var1=1",
		"var2=2",
		"var3=3",
		"var4=4",
		"var5=\"a\"",
		"var6=\"b\"",
		"var7=\"c\"",
		"var8=\"abc\"",
		"var9=\"aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffgggggggggg\"",
		"userMap[\"003\"][\"004\"]=6",
		"userMap[\"001\"][\"002\"]=5",
		"bankMap[\"10086\"]=[7,8,9]",
		"bankMap[\"10087\"]=[10,11,12,13,14]",
		"array1=[100,200,0]",
		"array2=[300,400]",
		"array3=[[500,0],[600,700],[0,0]]",
		"array4=[[700,800],[900,1000,1100]]",
		"strArray1=[\"a\",\"bbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffgggggggggg\",\"\"]",
		"strArray2=[\"b\",\"aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffgggggggggg\"]",
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