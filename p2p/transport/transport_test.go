// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:59
// last Modified Author: chenquan
// change log: add a header comment of this file
//

package transport

import (
	"testing"

	"fmt"
	"time"
	"encoding/hex"
)

func TestDes3(t *testing.T) {
	first:=time.Now().UnixNano()
	fmt.Println(first)
	for i := 0; i < 700; i += 1 {
		key := []byte("sfe023f_sefiel#fi32lf3e!")
		result, err := TripleDesEncrypt([]byte("polaris@studygol"), key)
		if err != nil {
			panic(err)
		}
		//fmt.Println(base64.StdEncoding.EncodeToString(result))
		origData, err := TripleDesDecrypt(result, key)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(origData))
	}
	now:=time.Now().UnixNano()
	fmt.Println((now-first)/int64(time.Millisecond))



	//var hSM0, hSM1 HandShakeManager

	//双方建立新的握手
	/*hSM0 := NewHandShakeManger()
	hSM1 := NewHandShakeManger()

	//双方根据对方的公钥生成共享密钥
	hSM0.GenerateSecret(hSM1.GetLocalPublicKey())
	hSM1.GenerateSecret(hSM0.GetLocalPublicKey())

	//双方利用共享密钥加密和解密消息
	println(string(hSM1.DecWithSecret(hSM0.EncWithSecret([]byte("hello, message length is not limited")))))
	println(string(hSM0.DecWithSecret(hSM1.EncWithSecret([]byte("ok, I got it")))))*/

}

var HSM1 *HandShakeManager
var HSM2 *HandShakeManager
func init(){
	HSM1 = NewHandShakeManger()
	HSM2 = NewHandShakeManger()
	pbk1 := HSM1.GetLocalPublicKey()
	pbk2 := HSM2.GetLocalPublicKey()
	HSM1.GenerateSecret(pbk2,"2")
	HSM2.GenerateSecret(pbk1,"1")
}

func TestHandShakeManager_DecWithSecret(t *testing.T) {
	enctrypted := HSM1.EncWithSecret([]byte("HELLO"),"2")
	t.Log("加密之后信息",hex.EncodeToString(enctrypted))
	decrypted := HSM2.DecWithSecret(enctrypted,"1")
	fmt.Println(string(decrypted))
}
