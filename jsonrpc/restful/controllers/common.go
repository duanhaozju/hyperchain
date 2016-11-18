package controllers

import (
	"net/http"
	"encoding/json"
)

type JSONObject struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Result interface{} `json:"result"`
}

func NewJSONObject (data interface{}, err JSONError) *JSONObject{

	if err != nil {

		if err.Error() == "leveldb: not found" {
			return &JSONObject{
				Code: 200,
				Result: nil,
				Message: "SUCCESS",
			}
		}

		return &JSONObject{
			Code: err.Code(),
			Result: err.Error(),
			Message: "FAILD",
		}
	}

	return &JSONObject{
		Code: 200,
		Result: data,
		Message: "SUCCESS",
	}


}

func write(w http.ResponseWriter, data interface{}) {
	//
	//if err != nil {
	//	json.NewEncoder(w).Encode(&callbackError{err.Error()})
	//} else {
	//	json.NewEncoder(w).Encode(data)
	//}
	json.NewEncoder(w).Encode(data)
}
