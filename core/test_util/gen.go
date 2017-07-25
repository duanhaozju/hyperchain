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
	value, _ := proto.Marshal(types.NewTransactionValue(10000, 10000, amount, nil, 0, 0))
	return &types.Transaction{
		From:      common.Hex2Bytes(strings.Split(args[2], " ")[1]),
		To:        common.Hex2Bytes(strings.Split(args[3], " ")[1]),
		Value:     value,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: common.Hex2Bytes(strings.Split(args[7], " ")[1]),
	}
}

func GenInvalidTransferTransactionRandomly() *types.Transaction {
	owd, _ := os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/core/test_util"))
	defer os.Chdir(owd)
	ret, _ := exec.Command("./agile", "-o", "transaction", "-v", "1000000000000000").Output()
	args := strings.Split(string(ret), "\n")
	timestamp, _ := strconv.ParseInt(strings.Split(args[4], " ")[1], 10, 64)
	nonce, _ := strconv.ParseInt(strings.Split(args[5], " ")[1], 10, 64)
	amount, _ := strconv.ParseInt(strings.Split(args[6], " ")[1], 10, 64)
	value, _ := proto.Marshal(types.NewTransactionValue(10000, 10000, amount, nil, 0, 0))
	return &types.Transaction{
		From:      common.Hex2Bytes(strings.Split(args[2], " ")[1]),
		To:        common.Hex2Bytes(strings.Split(args[3], " ")[1]),
		Value:     value,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: common.Hex2Bytes(strings.Split(args[7], " ")[1]),
	}
}

func GenSignatureInvalidTransferTransactionRandomly() *types.Transaction {
	owd, _ := os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/core/test_util"))
	defer os.Chdir(owd)
	ret, _ := exec.Command("./agile", "-o", "transaction", "-v", "1000").Output()
	args := strings.Split(string(ret), "\n")
	timestamp, _ := strconv.ParseInt(strings.Split(args[4], " ")[1], 10, 64)
	nonce, _ := strconv.ParseInt(strings.Split(args[5], " ")[1], 10, 64)
	amount, _ := strconv.ParseInt(strings.Split(args[6], " ")[1], 10, 64)
	value, _ := proto.Marshal(types.NewTransactionValue(10000, 10000, amount, nil, 0, 0))
	return &types.Transaction{
		From:      common.Hex2Bytes(strings.Split(args[2], " ")[1]),
		To:        common.Hex2Bytes(strings.Split(args[3], " ")[1]),
		Value:     value,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: common.Hex2Bytes("0x11111111"),
	}
}

func DeployContract(from, contractBin string) *types.Transaction {
	owd, _ := os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/core/test_util"))
	defer os.Chdir(owd)
	ret, _ := exec.Command("./agile", "-o", "transaction", "-t", "1", "-l", contractBin, "-from", from).Output()
	args := strings.Split(string(ret), "\n")
	timestamp, _ := strconv.ParseInt(strings.Split(args[4], " ")[1], 10, 64)
	nonce, _ := strconv.ParseInt(strings.Split(args[5], " ")[1], 10, 64)
	payload := strings.Split(args[6], " ")[1]
	value, _ := proto.Marshal(types.NewTransactionValue(10000, 10000, 0, common.Hex2Bytes(payload), 0, 0))
	return &types.Transaction{
		From:      common.Hex2Bytes(strings.Split(args[2], " ")[1]),
		To:        nil,
		Timestamp: timestamp,
		Nonce:     nonce,
		Value: 	   value,
		Signature: common.Hex2Bytes(strings.Split(args[8], " ")[1]),
	}
}

func GenContractTransactionRandomly(from, to, methodBin string) *types.Transaction {
	owd, _ := os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/core/test_util"))
	defer os.Chdir(owd)
	ret, _ := exec.Command("./agile", "-o", "transaction", "-t", "1", "-l", methodBin, "-from", from, "-to", to).Output()
	args := strings.Split(string(ret), "\n")
	timestamp, _ := strconv.ParseInt(strings.Split(args[4], " ")[1], 10, 64)
	nonce, _ := strconv.ParseInt(strings.Split(args[5], " ")[1], 10, 64)
	payload := strings.Split(args[6], " ")[1]
	value, _ := proto.Marshal(types.NewTransactionValue(10000, 10000, 0, common.Hex2Bytes(payload), 0, 0))
	return &types.Transaction{
		From:      common.Hex2Bytes(strings.Split(args[2], " ")[1]),
		To:        common.Hex2Bytes(strings.Split(args[3], " ")[1]),
		Timestamp: timestamp,
		Nonce:     nonce,
		Value: 	   value,
		Signature: common.Hex2Bytes(strings.Split(args[7], " ")[1]),
	}
}