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
	"encoding/base64"
)

func TestDes3(t *testing.T) {
	key := []byte("sfe023f_sefiel#fi32lf3e!")
	result, err := TripleDesEncrypt([]byte("polaris@studygol"), key)
	if err != nil {
		panic(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(result))
	origData, err := TripleDesDecrypt(result, key)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(origData))

}
