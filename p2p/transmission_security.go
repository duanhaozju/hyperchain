package p2p

import (
	pb "hyperchain/p2p/peermessage"
	"hyperchain/admittance"
	"hyperchain/core/crypto/primitives"
)

func SignCert(msg *pb.Message,CM *admittance.CAManager){
	//TODO ensure not nil for params
	signature := new(pb.Signature)
	msg.Signature = signature
	pri := CM.GetECertPrivKey()
	ecdsaEncry := primitives.NewEcdsaEncrypto("ecdsa")
	sign, err := ecdsaEncry.Sign(msg.Payload, pri)
	if err != nil {
		log.Errorf("cannot sign the msg %v",msg)
	}
	msg.Signature.Signature = sign
	signature.ECert = CM.GetECertByte()
	signature.RCert = CM.GetRCertByte()
}
