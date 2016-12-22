//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hmEncryption

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"hyperchain/crypto/ecies"
	"io/ioutil"
	"math/big"
	//"fmt"
)

type pp struct {
	N1 []byte
	N2 []byte
	N3 []byte
}

//prepare the hm_transaction parameters
//oldBalance 16byte
//transferAmount 16byte
//illegal_balance_hm 32byte

func PreHmTransaction(oldBalance []byte, transferAmount []byte, illegal_balance_hm []byte, whole_networkpublickey PaillierPublickey) (bool, []byte, []byte) {
	//var flag bool
	newBalance_hm := make([]byte, 32)
	transferAmount_hm := make([]byte, 32)
	//newBalance_local := make([]byte, 16)
	//transferAmount_ecc := make([]byte, 129)

	oldBalanceFillbyte := make([]byte,8)
	transferAmountFillbyte := make([]byte,8)
	newBalanceFillByte := make([]byte,8)

	//suffix := make([]byte, 8)
	//transferAmount = append(transferAmount, suffix...)
	//mark := CompareTwoBytes(oldBalance, transferAmount)
	//
	////filling transferAmount become  a 16 bytes slice
	//if mark == -1 {
	//	flag = false
	//	return flag,newBalance_local, newBalance_hm, transferAmount_hm, transferAmount_ecc
	//} else {
	//
	//	r, _ := rand.Prime(rand.Reader, 48)
	//	suffix_pro := r.Bytes()
	//	//suffix_final := append(middle, suffix_pro...)
	//	transferAmount = append(transferAmount[:(16-len(suffix_pro))], suffix_pro...)
	//	if CompareTwoBytes(oldBalance, transferAmount) == -1 {
	//		temp := CutByte(oldBalance)
	//		transferAmount = append(transferAmount[:(16-len(temp))], temp...)
	//	}
	//	flag = true
	//}

	//将oldbalance 和　transferamount 填充成８个字节

	if len(oldBalance)<=8  {
		oldBalanceFillbyte = append(oldBalanceFillbyte[:(8-len(oldBalance))],oldBalance...)
	}
	if len(transferAmount)<=8 {
		transferAmountFillbyte = append(transferAmountFillbyte[:(8-len(transferAmount))],transferAmount...)
	}


	oldBalance_bigint := new(big.Int)
	oldBalance_bigint = oldBalance_bigint.SetBytes(oldBalanceFillbyte)
	transferAmount_bigint := new(big.Int)
	transferAmount_bigint = transferAmount_bigint.SetBytes(transferAmountFillbyte)

	if oldBalance_bigint.Cmp(transferAmount_bigint) ==-1 {
		return false,nil,nil,nil;
	}



	newBalance_bigint := new(big.Int)
	newBalance_bigint = newBalance_bigint.Sub(oldBalance_bigint, transferAmount_bigint)
	newBalance_byte := newBalance_bigint.Bytes()

	if len(newBalance_byte)<=8 {
		newBalanceFillByte = append(newBalanceFillByte[:(8-len(newBalance_byte))],newBalance_byte...)
	}

	//fmt.Println(len(newBalanceFillByte))
	//fmt.Println(len(transferAmountFillbyte))
	//fmt.Println(len(oldBalanceFillbyte))

	phm := New_Paillier_Hmencryption()
	//newBalance_local = append(newBalance_local[:16-len(newBalance_byte)], newBalance_byte...)

	transferAmount_hm, _ = phm.Encrypto_message(&whole_networkpublickey, transferAmountFillbyte)

	//check illegal_balance_hm whether exist or not
	if illegal_balance_hm == nil {
		newBalance_hm, _ = phm.Encrypto_message(&whole_networkpublickey, newBalance_byte)
		//fmt.Println(newBalance_hm)
	} else {
		newBalance_temp, _ := phm.Encrypto_message(&whole_networkpublickey, newBalance_byte)
		newBalance_hm, _ = phm.Calculator(&whole_networkpublickey, "paillier", newBalance_temp, illegal_balance_hm)
	}

	//ecdsa_encrypto the transferamount
	//ecies_publickey := ecies.ImportECDSAPublic(ecdsa_publickey)
	//transferAmount_ecc, _ = ecies.Encrypt(rand.Reader, ecies_publickey, transferAmountFillbyte, nil, nil)
	//fmt.Println(len(transferAmount_ecc))


	return true, newBalance_hm, transferAmount_hm

}

//Verify the transaction's legality
//the first parameter is the encrypted old balance with the whole network publickey
//second parameter is the encrypted transfer amount with the whole network publickey
//third parameter is the encrypted new balance with the whole network publickey
func NodeVerify(whole_networkpublickey PaillierPublickey, oldBalance_hm []byte, transferAmount_hm []byte, newBalance_hm []byte) bool {
	var flag bool
	phm := New_Paillier_Hmencryption()
	sum, _ := phm.Calculator(&whole_networkpublickey, "paillier", transferAmount_hm, newBalance_hm)
	flag = EqualTwoBytes(oldBalance_hm, sum)

	return flag
}

//destination verify whether the amount is right or not
func DestinationVerify(transferAmount_hm []byte, transferAmount_ecc []byte, ecdsa_privatekey *ecdsa.PrivateKey, whole_networkpublickey PaillierPublickey) bool {

	ecies_privatekey := ecies.ImportECDSA(ecdsa_privatekey)
	transferAmount, _ := ecies_privatekey.Decrypt(rand.Reader, transferAmount_ecc, nil, nil)
	phm := New_Paillier_Hmencryption()
	transferAmount_hm_verify, _ := phm.Encrypto_message(&whole_networkpublickey, transferAmount)
	flag := EqualTwoBytes(transferAmount_hm, transferAmount_hm_verify)

	return flag
}

//put whole_networkpublickey to a file
func PutWholeNetworkPublickey(file string, whole_networkpublickey *PaillierPublickey) error {
	var publickey pp
	publickey.N1 = whole_networkpublickey.G.Bytes()
	publickey.N2 = whole_networkpublickey.N.Bytes()
	publickey.N3 = whole_networkpublickey.Nsquare.Bytes()

	data, _ := Encode(publickey)
	stringdata := hex.EncodeToString(data)

	return ioutil.WriteFile(file, []byte(stringdata), 0600)
}

//get wholenetworkpublickey from a file
func GetWholeNetworkPublickey(file string) (*PaillierPublickey, error) {

	var pailpub PaillierPublickey
	var publickey pp

	number, _ := Getnumber(file)
	err := Decode(number, &publickey)

	a1 := new(big.Int)
	a2 := new(big.Int)
	a3 := new(big.Int)
	pailpub.G = a1.SetBytes(publickey.N1)
	pailpub.N = a2.SetBytes(publickey.N2)
	pailpub.Nsquare = a3.SetBytes(publickey.N3)

	return &pailpub, err
}

func EqualTwoBytes(a []byte, b []byte) bool {
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

//compare two bytes
//flag=-1,a<b; flag=0,a=b; flag=1,a>b
func CompareTwoBytes(a []byte, b []byte) int {
	A := new(big.Int)
	B := new(big.Int)
	A = A.SetBytes(a)
	B = B.SetBytes(b)
	flag := A.Cmp(B)
	return flag
}

//get a []byte's last 8 byte and div 3
func CutByte(a []byte) []byte {
	temp := make([]byte, 8)
	if len(a) > 8 {
		copy(temp, a[len(a)-8:])
		big_temp := new(big.Int)
		big_temp = big_temp.SetBytes(temp)
		big_three := big.NewInt(3)
		div_big_temp := new(big.Int)
		div_big_temp = div_big_temp.Div(big_temp, big_three)
		div_big_temp_byte := div_big_temp.Bytes()
		temp = []byte{}
		temp = append(temp[:8-len(div_big_temp_byte)], div_big_temp_byte...)
	}
	return temp
}
