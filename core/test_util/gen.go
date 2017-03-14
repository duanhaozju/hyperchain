package test_util

import (
	"hyperchain/core/types"
	"os/exec"
	"os"
	"path"
	"strings"
	"hyperchain/common"
	"strconv"
	"github.com/golang/protobuf/proto"
)


func GenTransferTransactionRandomly() *types.Transaction {
	owd, _ := os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/core/test_util"))
	defer os.Chdir(owd)
	ret, _ := exec.Command("./agile", "-o", "transaction").Output()
	args := strings.Split(string(ret), "\n")
	timestamp, _ := strconv.ParseInt(strings.Split(args[4], " ")[1], 10, 64)
	nonce, _ := strconv.ParseInt(strings.Split(args[5], " ")[1], 10, 64)
	amount, _ := strconv.ParseInt(strings.Split(args[6], " ")[1], 10, 64)
	value, _ := proto.Marshal(types.NewTransactionValue(10000, 10000, amount, nil, false))
	return &types.Transaction{
		From:      common.Hex2Bytes(strings.Split(args[2], " ")[1]),
		To:        common.Hex2Bytes(strings.Split(args[3], " ")[1]),
		Value:     value,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: common.Hex2Bytes(strings.Split(args[7], " ")[1]),
	}
}
