package types

import (
	"testing"
	"hyperchain-alpha/utils"
	"hyperchain-alpha/encrypt"
	"time"
	"strconv"
	"github.com/stretchr/testify/assert"
)

//测试的时候需要修改文件路径
func TestTransaction_IsValidSign(t *testing.T) {
	accounts,_ := utils.GetAccount()
	t.Log(utils.GetGodAccount())
	godPubkey := utils.GetGodAccount()[0]["god"].PubKey
	godPrikey := utils.GetGodAccount()[0]["god"].PriKey
	zhangsanPubkey := accounts[0]["张三"].PubKey
	var tran = Transaction{
		From:encrypt.EncodePublicKey(&godPubkey),
		To:encrypt.EncodePublicKey(&zhangsanPubkey),
		Value:1000,
		TimeStamp:time.Now().Unix(),
	}
	signature,_ := encrypt.Sign(godPrikey,[]byte(tran.From + tran.To + strconv.Itoa(tran.Value) + strconv.FormatInt(tran.TimeStamp,10)))
	tran.Signature = signature
	t.Log("签名结束",tran)
	if assert.IsType(t,encrypt.Signature{},tran.Signature){

	}else{
		t.Error("签名signature非signature类型")

	}

	if encrypt.Verify(encrypt.DecodePublicKey(tran.From),[]byte(tran.Hash()),tran.Signature){
		t.Log("签名验证通过")
	}else{
		t.Error("签名验证不通过")
	}
}