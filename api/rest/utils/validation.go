package utils

import (
	"encoding/json"
	"github.com/hyperchain/hyperchain/api"
	"github.com/hyperchain/hyperchain/common"
	"strconv"
)

// CheckIntervalArgs
func CheckIntervalArgs(from, to string) (api.IntervalArgs, error) {

	var intervalArgs api.IntervalArgs

	jsonStr := "{\"from\":\"" + from + "\",\"to\":\"" + to + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &intervalArgs)

	if err != nil {
		return api.IntervalArgs{}, err
	}

	return intervalArgs, nil
}

func CheckIntervalTimeArgs(start, end string) (api.IntervalTime, error) {
	startTime, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return api.IntervalTime{}, err
	}

	endTime, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return api.IntervalTime{}, err
	}

	return api.IntervalTime{
		StartTime: startTime,
		Endtime:   endTime,
	}, nil
}

func CheckHash(hash string) (common.Hash, error) {

	jsonObj := struct {
		Hash common.Hash
	}{}

	jsonStr := "{\"hash\":\"" + hash + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return common.Hash{}, err
	}

	return jsonObj.Hash, nil

}

func CheckAddress(address string) (common.Address, error) {

	jsonObj := struct {
		Address common.Address
	}{}

	jsonStr := "{\"address\":\"" + address + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return common.Address{}, err
	}

	return jsonObj.Address, nil
}

func CheckBlockNumber(number string) (api.BlockNumber, error) {
	jsonObj := struct {
		BlkNum api.BlockNumber
	}{}

	jsonStr := "{\"blkNum\":\"" + number + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return api.BlockNumber(0), err
	}

	return jsonObj.BlkNum, nil
}

func CheckNumber(number string) (api.Number, error) {
	jsonObj := struct {
		Num api.Number
	}{}

	jsonStr := "{\"num\":\"" + number + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return api.Number(0), err
	}

	return jsonObj.Num, nil
}

func CheckBlkNumAndIndexParams(blkNum, index string) (api.BlockNumber, api.Number, error) {
	jsonObj := struct {
		BlkNum api.BlockNumber
		Index  api.Number
	}{}

	jsonStr := "{\"blkNum\":\"" + blkNum + "\",\"index\":\"" + index + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return api.BlockNumber(0), api.Number(0), err
	}

	return jsonObj.BlkNum, jsonObj.Index, nil
}

func CheckBlkHashAndIndexParams(blkHash, index string) (common.Hash, api.Number, error) {
	jsonObj := struct {
		BlkHash common.Hash
		Index   api.Number
	}{}

	jsonStr := "{\"blkHash\":\"" + blkHash + "\",\"index\":\"" + index + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return common.Hash{}, api.Number(0), err
	}

	return jsonObj.BlkHash, jsonObj.Index, nil
}
