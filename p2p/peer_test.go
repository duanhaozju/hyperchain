package p2p

import (
	"fmt"
	"hyperchain/core/crypto/primitives"
	pb "hyperchain/p2p/peermessage"
	"testing"
)

func TestHandShake(t *testing.T) {
	eca, getErr1 := primitives.GetConfig("../config/cert/server/eca.cert")
	if getErr1 != nil {
		log.Error("cannot read ecert.", getErr1)
	}
	ecertBtye := []byte(eca)

	rca, getErr2 := primitives.GetConfig("../config/cert/server/rca.cert")
	if getErr2 != nil {
		log.Error("cannot read ecert.", getErr2)
	}
	rcertByte := []byte(rca)

	signature := pb.Signature{
		Ecert: ecertBtye,
		Rcert: rcertByte,
	}

	var test map[string]int

	test = make(map[string]int)
	str := "abc"
	test[str] = 123
	fmt.Println(test[str])

	fmt.Println("```````````````")
	//var b []byte

	fmt.Println(signature.Ecert)

	fmt.Println("------------------")

	fmt.Println(signature.Rcert)

	fmt.Println("------------------")

	fmt.Println(signature.Signature)
}

func TestSetPublikey(t *testing.T) {

}
