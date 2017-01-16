package transport

import (
	"testing"
	//"hyperchain/p2p/transport"
	"fmt"
	//s"google.golang.org/grpc/transport"
	//"hyperchain/core/crypto/primitives"
	//"hyperchain/core/crypto/primitives"
	"hyperchain/p2p/transport/ecdh"
	"crypto/elliptic"
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

func TestSetKey(t *testing.T)  {
	nHSM := NewHandShakeMangerNew();

	key := nHSM.GetLocalPublicKey()

	//pub,_ := nHSM.e.Unmarshal(key)
	e := ecdh.NewEllipticECDH(elliptic.P256())
	pub,_ := e.Unmarshal(key)
	//cert,_ := primitives.GetConfig("../../config/cert/server/eca.cert")
	//ecert := primitives.ParseCertificate(cert)
	//
	////(*ecert).PublicKey
	nHSM.SetSignPublicKey(pub,"123")

	fmt.Println(nHSM.signPublickey["123"])
}