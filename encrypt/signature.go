package encrypt

import (
	"math/big"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

//数字签名结构体
type Signature struct {
	R big.Int
	S big.Int
}
func newSignature(r,s big.Int) Signature{
	signature := Signature{
		R:r,
		S:s,
	}
	return signature
}
//实现string接口
func (s Signature)String() string{
	signature := s.R.Bytes()
	signature = append(signature, s.S.Bytes()...)
	return hex.EncodeToString(signature)
}

//签名序列化
func EncodeSignature(sig *Signature) string{
	r_str := sig.R.String()
	s_str := sig.S.String()
	sign_string  :=fmt.Sprintf(`{"r":"%s","s":"%s"}`,r_str,s_str)
	return sign_string
}
//签名反序列化
func DecodeSignature(sigString string,signature *Signature)error{
	var temp map[string]string
	json.Unmarshal([]byte(sigString),&temp)
	//fmt.Printf("%#v\n",temp["r"])
	r := new(big.Int)
	s := new(big.Int)
	r.SetString(temp["r"], 10) //base 10
	s.SetString(temp["s"], 10) //base 10
	(*signature).R = *r
	(*signature).S = *s
	return nil
}