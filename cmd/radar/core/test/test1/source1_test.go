package test

import (
	"testing"
	"hyperchain/cmd/radar/core/api"
	"hyperchain/cmd/radar/core/test"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"encoding/json"
	"io/ioutil"
	"hyperchain/cmd/radar/contract"
)
func TestSource1 (t *testing.T) {

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
	res, err := api.GetResult("source1.solc", db, "0346de490a2ebd37e1f6f216b1376c60f19cb9d6", originKeyOfMaps)

	var rightResult map[string][]string
	rightResult = make(map[string][]string)
	var temp = []string{"bank1={bankID:\"001\",bankName:\"CIBC\",balance:10000,bankState:ABNORMAL}",
		"bank2={bankID:\"002\",bankName:\"CBC\",balance:354,bankState:ABNORMAL}",
		"bankMap[\"001\"]={bankID:\"001\",bankName:\"CIBC\",balance:10000,bankState:ABNORMAL}",
		"bankMap[\"002\"]={bankID:\"002\",bankName:\"CBC\",balance:354,bankState:ABNORMAL}",
		"balance=[10,11,12]",
		"data=[[13,14],[15,16,17,18]]"}
	rightResult["Demo"] = temp

	if err != nil {
		t.Error(err.Error())
	} else {
		if test.Judge(res, rightResult) == false {
			t.Error("sorry, not equal!")
		}
	}
}