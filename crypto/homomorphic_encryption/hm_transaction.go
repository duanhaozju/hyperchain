package homomorphic_encryption

import (
	"crypto/rand"
	"math/big"
)

//对传递给B的转账金额目前进行明文传输，所以不需要进行ecc的加密。传进来的参数目前也不需要ecc公钥
func Pre_transaction(oldBalance []byte, transferAmount []byte, illegal_balance_hm []byte, whole_networkpublickey PaillierPublickey) (bool, []byte, []byte, []byte, []byte) {
	var flag bool
	newBalance_hm := make([]byte, 32)
	transferAmount_hm := make([]byte, 32)
	newBalance_local := make([]byte, 16)
	transferAmount_ecc := make([]byte, 129)

	suffix := make([]byte, 8)
	//	suffix := []byte{}
	transferAmount = append(transferAmount[:(16-len(suffix))], suffix...)
	mark := Comparetwobytes(oldBalance, transferAmount)

	//填充transferAmount为16个字节
	if mark == -1 {
		flag = false
		return flag, newBalance_local, newBalance_hm, transferAmount_hm, transferAmount_ecc
	} else {

		//middle := []byte{0, 0}
		r, _ := rand.Prime(rand.Reader, 48)
		suffix_pro := r.Bytes()
		//suffix_final := append(middle, suffix_pro...)
		transferAmount = append(transferAmount[:(16-len(suffix_pro))], suffix_pro...)
		if Comparetwobytes(oldBalance, transferAmount) == -1 {
			temp := Cutbyte(oldBalance)
			transferAmount = append(transferAmount[:(16-len(temp))], temp...)
		}
		flag = true
	}

	oldBalance_bigint := new(big.Int)
	oldBalance_bigint = oldBalance_bigint.SetBytes(oldBalance)
	transferAmount_bigint := new(big.Int)
	transferAmount_bigint = transferAmount_bigint.SetBytes(transferAmount)

	newBalance_bigint := new(big.Int)
	newBalance_bigint = newBalance_bigint.Sub(oldBalance_bigint, transferAmount_bigint)
	newBalance_byte := newBalance_bigint.Bytes()

	phm := New_Paillier_Hmencryption()
	copy(newBalance_local, newBalance_byte)

	transferAmount_hm, _ = phm.Encrypto_message(&whole_networkpublickey, transferAmount)
	//判断是否具有非法同态金额,同时进行加密
	if illegal_balance_hm == nil {
		newBalance_hm, _ = phm.Encrypto_message(&whole_networkpublickey, newBalance_byte)
	} else {
		newBalance_temp, _ := phm.Encrypto_message(&whole_networkpublickey, newBalance_byte)
		newBalance_hm, _ = phm.Calculator(&whole_networkpublickey, "paillier", newBalance_temp, illegal_balance_hm)
	}

	copy(transferAmount_ecc, transferAmount)
	return flag, newBalance_local, newBalance_hm, transferAmount_hm, transferAmount_ecc

}

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
//flag=-1,a<b; flag=0,a=b; flag=1,a>b
func Comparetwobytes(a []byte, b []byte) int {
	A := new(big.Int)
	B := new(big.Int)
	A = A.SetBytes(a)
	B = B.SetBytes(b)
	flag := A.Cmp(B)
	return flag
}

//取[]byte数组的后8个字节的1/3取整
func Cutbyte(a []byte) []byte {
	temp := make([]byte, 8)
	//	temp := []byte{}
	if len(a) > 8 {
		copy(temp, a[len(a)-8:])
		big_temp := new(big.Int)
		big_temp = big_temp.SetBytes(temp)
		big_three := big.NewInt(3)
		div_big_temp := new(big.Int)
		div_big_temp = div_big_temp.Div(div_big_temp, big_three)
		div_big_temp_byte := div_big_temp.Bytes()
		temp = []byte{}
		temp = append(temp[:8-len(div_big_temp_byte)], div_big_temp_byte...)
	}
	return temp
}
