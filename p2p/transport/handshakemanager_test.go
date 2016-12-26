package transport

import (
	"testing"
	//"hyperchain/p2p/transport"
	"fmt"
	//s"google.golang.org/grpc/transport"
)

func TestNewHandShakeMangerNew(t *testing.T) {
	nHSM := NewHandShakeMangerNew()
	fmt.Println(nHSM.privateKey)
	fmt.Println(nHSM.publicKey)
}

func TestNewGetLocalPublicKey(t *testing.T)  {
	nHSM := NewHandShakeMangerNew()


	fmt.Println(nHSM.GetLocalPublicKey())
	//fmt.Println(nHSM)
}

func TestGenerateSecret(t *testing.T){
	nHSM := NewHandShakeMangerNew()
	hsm := NewHandShakeManger()

	remoteKey := hsm.GetLocalPublicKey()

	nHSM.GenerateSecret(remoteKey,"1234567")

	fmt.Println(nHSM.GetSecret("1234567"))
	//fmt.Println(nHSM.secrets["1234567"])
}