package encrypt

import (
	"math/big"
	"encoding/hex"
)

//数字签名结构体
type Signature struct {
	R big.Int
	S big.Int
}
func NewSignature(r,s big.Int) Signature{
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
