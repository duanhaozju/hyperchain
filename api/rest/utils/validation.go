package utils

import (
	"encoding/json"
	"hyperchain/api"
	"hyperchain/common"
	"strconv"
)

// CheckIntervalArgs
func CheckIntervalArgs(from, to string) (hpc.IntervalArgs, error) {

	var intervalArgs hpc.IntervalArgs

	jsonStr := "{\"from\":\"" + from + "\",\"to\":\"" + to + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &intervalArgs)

	if err != nil {
		return hpc.IntervalArgs{}, err
	}

	return intervalArgs, nil
}

func CheckIntervalTimeArgs(start, end string) (hpc.IntervalTime, error) {
	startTime, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return hpc.IntervalTime{}, err
	}

	endTime, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return hpc.IntervalTime{}, err
	}

	return hpc.IntervalTime{
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

func CheckBlockNumber(number string) (hpc.BlockNumber, error) {
	jsonObj := struct {
		BlkNum hpc.BlockNumber
	}{}

	jsonStr := "{\"blkNum\":\"" + number + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return hpc.BlockNumber(0), err
	}

	return jsonObj.BlkNum, nil
}

func CheckNumber(number string) (hpc.Number, error) {
	jsonObj := struct {
		Num hpc.Number
	}{}

	jsonStr := "{\"num\":\"" + number + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return hpc.Number(0), err
	}

	return jsonObj.Num, nil
}

func CheckBlkNumAndIndexParams(blkNum, index string) (hpc.BlockNumber, hpc.Number, error) {
	jsonObj := struct {
		BlkNum hpc.BlockNumber
		Index  hpc.Number
	}{}

	jsonStr := "{\"blkNum\":\"" + blkNum + "\",\"index\":\"" + index + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return hpc.BlockNumber(0), hpc.Number(0), err
	}

	return jsonObj.BlkNum, jsonObj.Index, nil
}

func CheckBlkHashAndIndexParams(blkHash, index string) (common.Hash, hpc.Number, error) {
	jsonObj := struct {
		BlkHash common.Hash
		Index   hpc.Number
	}{}

	jsonStr := "{\"blkHash\":\"" + blkHash + "\",\"index\":\"" + index + "\"}"

	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return common.Hash{}, hpc.Number(0), err
	}

	return jsonObj.BlkHash, jsonObj.Index, nil
}
