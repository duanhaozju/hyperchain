package controllers

import (
	"net/http"
	"hyperchain/hpc"
	"encoding/json"
	"github.com/gorilla/mux"
	"hyperchain/common"
)

//const (
//	HashLength    = 32
//)

// 处理请求：GET "/v1/blocks/latest"
func GetLatestBlock(w http.ResponseWriter, r *http.Request) {
	PublicBlockAPIInterface := hpc.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*hpc.PublicBlockAPI)

	block, err := PublicBlockAPI.LatestBlock()

	if err != nil {
		write(w, NewJSONObject(block, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(block, nil))
	}
}

// 处理请求：GET "/v1/blocks/{block_num_or_hash}"
func GetBlockByNumberOrHash(w http.ResponseWriter, r *http.Request) {
	PublicBlockAPIInterface := hpc.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*hpc.PublicBlockAPI)

	vars := mux.Vars(r)
	condition := vars["block_num_or_hash"]
	log.Error(condition)
	// verify whether a string can represent a valid hex-encoded hash or not.
	if (len(condition) == 2+2*HashLength && common.IsHex(condition)) || (len(condition) == 2*HashLength && common.IsHex("0x"+condition)) {
		block, err := PublicBlockAPI.GetBlockByHash(common.HexToHash(condition))
		log.Errorf("%#v",block)
		if err != nil {
			write(w, NewJSONObject(block, &callbackError{err.Error()}))
		} else {
			write(w, NewJSONObject(block, nil))
		}
	} else {
		var args hpc.BlockNumber

		if err := json.Unmarshal([]byte(condition), &args); err != nil {
			write(w, NewJSONObject(nil, &callbackError{err.Error()}))
		} else {
			log.Error(args)
			block, err := PublicBlockAPI.GetBlockByNumber(args)
			log.Errorf("%#v",block)
			if err != nil {
				write(w, NewJSONObject(block, &callbackError{err.Error()}))
			} else {
				write(w, NewJSONObject(block, nil))
			}
		}

	}
}

// 处理请求：POST "/v1/blocks"
func GetBlocks(w http.ResponseWriter, r *http.Request) {
	var args hpc.IntervalArgs

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)

	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	}

	PublicBlockAPIInterface := hpc.GetApiObjectByNamespace("block").Service
	PublicBlockAPI := PublicBlockAPIInterface.(*hpc.PublicBlockAPI)

	blocks, err := PublicBlockAPI.GetBlocks(args)
	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(blocks, nil))
	}
}