package primitives

import (
	"crypto/sha256"
	"fmt"
	"hyperchain/common"
	"testing"
)

func TestECDSAVerifyTransport(t *testing.T) {
	msg := []byte("hyperchain")
	h := sha256.New()

	digest := make([]byte, 32)
	h.Write(msg)
	h.Sum(digest[:0])

	fmt.Println("hash", common.ToHex(digest))

	sign := common.Hex2Bytes("3046022100e94878e31e7ab4d707144ba17c516c13c7b10860fc09d79fda192aef35a95408022100d9031e82f7b40b4bb1a5922b92d3309676e3d4118df2fb34b4dda7fddb3cc8d3")
	pubkeyPem, err := GetConfig("/Users/chenquan/Workspace/java/encrypt/src/test/resources/cert/u.pub")
	if err != nil {
		t.Error(err)
	}

	pubkey, _ := ParsePubKey(pubkeyPem)

	b, e := ECDSAVerifyTransport(pubkey, msg, sign)
	t.Log(b)
	t.Log(e)

}