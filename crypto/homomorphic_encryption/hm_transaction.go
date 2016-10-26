package homomorphic_encryption

import (
	"math/big"
)

//Verify the transaction's legality
//the first parameter is the encrypted old balance with the whole network publickey
//second parameter is the encrypted transfer amount with the whole network publickey
//third parameter is the encrypted new balance with the whole network publickey
func Node_Verify(whole_networkpublickey PaillierPublickey, oldBalance_hm []byte, transferAmount_hm []byte, newBalance_hm []byte) bool {
	var flag bool
	phm := New_Paillier_Hmencryption()
	sum, _ := phm.Calculator(&whole_networkpublickey, "paillier", transferAmount_hm, newBalance_hm)
	flag = Equaltwobytes(oldBalance_hm, sum)

	return flag
}

func Equaltwobytes(a []byte, b []byte) bool {
	var flag bool
	if len(a) == len(b) {
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				flag = false
				break
			}
			flag = true
		}

	} else {
		flag = false
	}
	return flag
}

//比较两个[]byte的实际大小
func Comparetwobytes(a []byte, b []byte) int {
	A := new(big.Int)
	B := new(big.Int)
	A = A.SetBytes(a)
	B = B.SetBytes(b)
	flag := A.Cmp(B)
	return flag
}
