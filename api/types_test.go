//author:zsx
//data:2016-11-2
package api

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/core"
	"github.com/hyperchain/hyperchain/hyperdb"
	"math"
	"strings"
	"testing"
)

func TestHex(t *testing.T) {
	n := Number(1024)
	ref := n.Hex()
	str := string("0x400")
	if !strings.EqualFold(ref, str) {
		t.Errorf("Number.Hex wrong")
	}
}
func TestMarshalJSON(t *testing.T) {
	n := Number(1024)
	ref, _ := n.MarshalJSON()
	n2, _ := json.Marshal(n.Hex())
	if len(ref) != len(n2) {
		t.Errorf("n.MarshalJSON wrong")
	}
	for a := 0; a < len(n2); a++ {
		if n2[a] != ref[a] {
			t.Errorf("n.MarshalJSON wrong")
		}
	}
}
func TestUnmarshalJSON(t *testing.T) {
	hyperdb.Setclose()
	core.InitDB("./build/keystore1", 8004)
	n := Number(1024)
	err := n.UnmarshalJSON([]byte{'a', 'b'})
	if !strings.EqualFold(err.Error(), "invalid number ab") {
		t.Errorf(err.Error())
		t.Errorf("UnmarshalJSONtest1 wrong")
	}

	err = n.UnmarshalJSON([]byte{'4', '5'})
	if err != nil {
		t.Errorf("UnmarshalJSON wrong")
		t.Error(err)
		if n != Number(45) {
			t.Errorf("UnmarshalJSONtest2 wrong")
		}
	}

	err = n.UnmarshalJSON([]byte{})
	fmt.Println(err)

	err = n.UnmarshalJSON([]byte{'-', '1'})
	if err.Error() != "number out of range" {
		t.Errorf("UnmarshalJSONtest4 wrong")
	}

}
func TestToInt64(t *testing.T) {
	n := Number(456)
	ref := n.Int64()
	if ref != int64(456) {
		t.Errorf("ToInt64 wrong")
	}
	var n2 *Number

	ref = n2.Int64()
	if ref != int64(0) {
		t.Errorf("ToInt64.1 wrong")
	}
}

func TestToInt(t *testing.T) {
	n := Number(456)
	ref := n.Int()
	if ref != int(456) {
		t.Errorf("ToInt.1 wrong")
	}
}

func Test_Block(t *testing.T) {
	n := BlockNumber(456)

	ref, err := n.Hex()
	if err != nil {
		t.Errorf("%v", err)
	}
	if ref != "0x1c8" {
		t.Errorf("BlockNumber.Hex() fail")
	}

	ref2, err := n.BlockNumberToUint64(math.MaxUint64)
	if err != nil {
		t.Errorf("%v", err)
	}
	if ref2 != uint64(456) {
		t.Errorf("BlockNumber.ToUint64() fail")
	}

	_, err := n.MarshalJSON()
	if err != nil {
		t.Errorf("BlockNumber.MarshalJSON() fail")
	}

	err = n.UnmarshalJSON([]byte{'4', '5'})
	if err != nil {
		t.Errorf("UnmarshalJSON wrong")
		t.Error(err)
		if n != BlockNumber(45) {
			t.Errorf("UnmarshalJSONtest2 wrong")
		}
	}
	err = n.UnmarshalJSON([]byte{'e', 'a', 'r', 'l', 'i', 'e', 's', 't'})
	if err != nil {
		t.Errorf("UnmarshalJSON wrong")
		t.Error(err)
	}
	if n != BlockNumber(2) {
		t.Errorf("UnmarshalJSONtest3 wrong")
	}

	err = n.UnmarshalJSON([]byte{'l', 'a', 't', 'e', 's', 't'})
	if err != nil {
		t.Errorf("UnmarshalJSON wrong")
		t.Error(err)
	}

	err = n.UnmarshalJSON([]byte{'p', 'e', 'n', 'd', 'i', 'n', 'g'})
	if err != nil {
		t.Errorf("UnmarshalJSON wrong")
		t.Error(err)
	}

	err = n.UnmarshalJSON([]byte{'a', 'b'})
	if err == nil {
		t.Errorf(err.Error())
		t.Errorf("UnmarshalJSONtest1 wrong")
	}
}
