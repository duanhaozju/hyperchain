package controllers

import (
	"net/http"
	"hyperchain/hpc"
	"encoding/json"
)

// 处理请求： POST "/v1/contracts/compile"
func CompileContract(w http.ResponseWriter, r *http.Request) {
	var args string

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)

	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	}

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	hash, err := PublicContractAPI.CompileContract(args)

	//write(w, hash, err)

	if err != nil {
		write(w, NewJSONObject(hash, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(hash, nil))
	}
}

// 处理请求：POST "/v1/contracts/deploy"
func DeployContract(w http.ResponseWriter, r *http.Request) {
	var args hpc.SendTxArgs

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)

	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	}

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	hash, err := PublicContractAPI.DeployContract(args)

	if err != nil {
		write(w, NewJSONObject(hash, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(hash, nil))
	}
}

// 处理请求：POST "/v1/contracts/invoke"
func InvokeContract(w http.ResponseWriter, r *http.Request) {
	var args hpc.SendTxArgs

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)

	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	}

	PublicContractAPIInterface := hpc.GetApiObjectByNamespace("contract").Service
	PublicContractAPI := PublicContractAPIInterface.(*hpc.PublicContractAPI)

	hash, err := PublicContractAPI.InvokeContract(args)

	if err != nil {
		write(w, NewJSONObject(hash, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(hash, nil))
	}
}
