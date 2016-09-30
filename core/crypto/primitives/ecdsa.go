// author: Lizhong kuang
// date: 2016-09-29
package primitives

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"math/big"
)

// ECDSASignature represents an ECDSA signature
type ECDSASignature struct {
	R, S *big.Int
}

// NewECDSAKey generates a new ECDSA Key
func NewECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(GetDefaultCurve(), rand.Reader)
}

// ECDSASignDirect signs
func ECDSASignDirect(signKey interface{}, msg []byte) (*big.Int, *big.Int, error) {
	temp := signKey.(*ecdsa.PrivateKey)
	h := Hash(msg)
	r, s, err := ecdsa.Sign(rand.Reader, temp, h)
	if err != nil {
		return nil, nil, err
	}

	return r, s, nil
}

// ECDSASign signs
func ECDSASign(signKey interface{}, msg []byte) ([]byte, error) {
	temp := signKey.(*ecdsa.PrivateKey)
	h := Hash(msg)
	r, s, err := ecdsa.Sign(rand.Reader, temp, h)
	if err != nil {
		return nil, err
	}

	//	R, _ := r.MarshalText()
	//	S, _ := s.MarshalText()
	//
	//	fmt.Printf("r [%s], s [%s]\n", R, S)

	raw, err := asn1.Marshal(ECDSASignature{r, s})
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// ECDSAVerify verifies
func ECDSAVerify(verKey interface{}, msg, signature []byte) (bool, error) {
	ecdsaSignature := new(ECDSASignature)
	_, err := asn1.Unmarshal(signature, ecdsaSignature)
	if err != nil {
		return false, nil
	}

	//	R, _ := ecdsaSignature.R.MarshalText()
	//	S, _ := ecdsaSignature.S.MarshalText()
	//	fmt.Printf("r [%s], s [%s]\n", R, S)

	temp := verKey.(*ecdsa.PublicKey)
	h := Hash(msg)
	return ecdsa.Verify(temp, h, ecdsaSignature.R, ecdsaSignature.S), nil
}

// VerifySignCapability tests signing capabilities
func VerifySignCapability(tempSK interface{}, certPK interface{}) error {
	/* TODO: reactive or remove
	msg := []byte("This is a message to be signed and verified by ECDSA!")

	sigma, err := ECDSASign(tempSK, msg)
	if err != nil {
		//		log.Errorf("Error signing [%s].", err.Error())

		return err
	}

	ok, err := ECDSAVerify(certPK, msg, sigma)
	if err != nil {
		//		log.Errorf("Error verifying [%s].", err.Error())

		return err
	}

	if !ok {
		//		log.Errorf("Signature not valid.")

		return errors.New("Signature not valid.")
	}

	//	log.Infof("Verifing signature capability...done")
	*/
	return nil
}