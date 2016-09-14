// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:59
// last Modified Author: chenquan
// change log: add a header comment of this file
//

package transport

import (
	"testing"


)

func TestDes3(t *testing.T) {
/*	key := []byte("sfe023f_sefiel#fi32lf3e!")
	result, err := TripleDesEncrypt([]byte("polaris@studygol"), key)
	if err != nil {
		panic(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(result))
	origData, err := TripleDesDecrypt(result, key)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(origData))*/


	var hSM0, hSM1 handShakeManager

	//双方建立新的握手
	hSM0.newHandShakeManger()
	hSM1.newHandShakeManger()

	//双方根据对方的公钥生成共享密钥
	hSM0.generateSecret(hSM1.getLocalPublicKey())
	hSM1.generateSecret(hSM0.getLocalPublicKey())

	//双方利用共享密钥加密和解密消息
	println(string(hSM1.decWithSecret(hSM0.encWithSecret([]byte("hello, message length is not limited")))))
	println(string(hSM0.decWithSecret(hSM1.encWithSecret([]byte("ok, I got it")))))

}
