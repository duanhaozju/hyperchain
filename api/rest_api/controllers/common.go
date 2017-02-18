package controllers

import (
	"hyperchain/api"
)

type JSONObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func NewJSONObject(data interface{}, err error) *JSONObject {

	if err != nil {

		jsonrpcError := err.(hpc.JSONRPCError)

		return &JSONObject{
			Code:    jsonrpcError.Code(),
			Message: jsonrpcError.Error(),
			Result:  nil,
		}
	}

	return &JSONObject{
		Code:    0,
		Message: "SUCCESS",
		Result:  data,
	}

}
