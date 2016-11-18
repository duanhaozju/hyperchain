package controllers
//
//import (
//	"hyperchain/hpc"
//	"math/big"
//	"hyperchain/core"
//	"fmt"
//	"strings"
//	"strconv"
//)
//
//func StringToNumber (data string) hpc.Number{
//	input := strings.TrimSpace(data)
//	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
//		input = input[1 : len(input)-1]
//	}
//
//	in := new(big.Int)
//	_, ok := in.SetString(input, 0)
//
//	latest_number := core.GetChainCopy().Height
//
//	if !ok { // test if user supplied string tag
//
//		strBlockNumber := input
//		if strBlockNumber == "latest" {
//			*n = *NewUint64ToBlockNumber(latest_number)
//			//*n = BlockNumber(latestBlockNumber)
//			return nil
//		}
//
//		if strBlockNumber == "earliest" {
//			*n = BlockNumber(earliestBlockNumber)
//			return nil
//		}
//
//		if strBlockNumber == "pending" {
//			*n = BlockNumber(pendingBlockNumber)
//			return nil
//		}
//
//		return fmt.Errorf(`invalid block number %s`, data)
//	}
//
//	if v, err := strconv.ParseUint(input, 0, 0);err != nil {
//		return fmt.Errorf("block number %v may be out of range",input)
//	} else if (v <= 0) {
//		return fmt.Errorf("block number can't be negative or zero, but get %v", input)
//	} else if v > latest_number{
//		return fmt.Errorf("block number is out of range, and now latest block number is %d", latest_number)
//	} else {
//		*n = *NewUint64ToBlockNumber(v)
//		return nil
//	}
//}