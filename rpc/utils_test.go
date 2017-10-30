package jsonrpc

import (
	"testing"
	"reflect"
	"github.com/stretchr/testify/assert"
	"math/big"
	"encoding/json"
)

func TestIsHexNum(t *testing.T) {
	reply := big.NewInt(11)
	rtyp := reflect.TypeOf(reply)
	ifHexNum := isHexNum(rtyp)
	assert.True(t, ifHexNum)

	flt := big.NewFloat(1.5)
	rtyp = reflect.TypeOf(flt)

	ifHexNum = isHexNum(rtyp)
	assert.False(t, ifHexNum)

	ifHexNum = isHexNum(nil)
	assert.False(t, ifHexNum)
}

func TestSplitRawMessage(t *testing.T) {
	//req := `{"jsonrpc": "2.0", "method": "test_hello", "params": [], "id": 1, "namespace":"global"}`
	params := `["1,2"]`
	rawMsg := json.RawMessage(params)

	result, err := splitRawMessage(rawMsg)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "1", result[0])
	assert.Equal(t, "2", result[1])

	rawMsg = json.RawMessage("1,2")
	_, err = splitRawMessage(rawMsg)
	assert.NotNil(t, err)

	rawMsg = json.RawMessage("")
	_, err = splitRawMessage(rawMsg)
	assert.NotNil(t, err)
}
