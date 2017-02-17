package controllers

import (
	"hyperchain/api"
)

type JSONObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func NewJSONObject(data interface{}, err hpc.JSONRPCError) *JSONObject {

	if err != nil {

		return &JSONObject{
			Code:    err.Code(),
			Message: err.Error(),
			Result:  nil,
		}
	}

	return &JSONObject{
		Code:    0,
		Message: "SUCCESS",
		Result:  data,
	}

}
