package controllers

type JSONObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func NewJSONObject(data interface{}, err JSONError) *JSONObject {

	if err != nil {

		if err.Error() == "leveldb: not found" {
			return &JSONObject{
				Code:    200,
				Result:  nil,
				Message: "SUCCESS",
			}
		}

		return &JSONObject{
			Code:    err.Code(),
			Result:  err.Error(),
			Message: "FAILD",
		}
	}

	return &JSONObject{
		Code:    200,
		Result:  data,
		Message: "SUCCESS",
	}

}
