//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hmEncryption

import (
	//"crypto/ecdsa"
	//"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	//"hyperchain/crypto"
	//"github.com/astaxie/beego/toolbox"
)

func Generate_wholenetpublickey(g []byte, n []byte, nsquare []byte) PaillierPublickey {
	var publickey PaillierPublickey
	G1 := new(big.Int)
	n1 := new(big.Int)
	nsquare1 := new(big.Int)
	publickey.G = G1.SetBytes(g)
	publickey.N = n1.SetBytes(n)
	publickey.Nsquare = nsquare1.SetBytes(nsquare)
	return publickey
}

func Test_NodeVerify(t *testing.T) {
	old := big.NewInt(500)
	old_byte := old.Bytes()
	trans := big.NewInt(200)
	trans_byte := trans.Bytes()

	oldbalance := make([]byte, 16)
	oldbalance = append(oldbalance[:(16-len(old_byte))], old_byte...)

	transferamount := make([]byte, 16)
	transferamount = append(transferamount[:(16-len(trans_byte))], trans_byte...)

	G1 := []byte{195, 153, 245, 222, 195, 152, 112, 47, 129, 79, 150, 122, 83, 4, 237, 142, 251, 215, 241, 160, 24, 216, 246, 105, 237, 83, 182, 53, 214, 53, 139, 73}

	N1 := []byte{238, 223, 119, 206, 43, 23, 14, 2, 159, 122, 233, 6, 43, 53, 223, 47}
	nsquare1 := []byte{222, 228, 69, 213, 73, 95, 66, 80, 24, 124, 30, 174, 198, 90, 112, 198, 28, 116, 247, 151, 13, 92, 246, 31, 56, 225, 40, 131, 246, 8, 234, 161}
	publickey := Generate_wholenetpublickey(G1, N1, nsquare1)

	oldBalance_hm := []byte{139, 222, 135, 44, 112, 98, 183, 241, 13, 118, 121, 159, 232, 199, 15, 122, 142, 199, 189, 205, 14, 101, 203, 237, 113, 142, 85, 61, 15, 114, 205, 208}
	transferAmount_hm := []byte{33, 91, 137, 191, 96, 106, 118, 67, 185, 132, 43, 1, 104, 139, 5, 168, 242, 111, 122, 196, 29, 56, 118, 141, 111, 103, 112, 191, 17, 82, 21, 98}
	newBalance_hm := []byte{60, 17, 254, 228, 200, 71, 7, 13, 242, 62, 3, 241, 78, 200, 135, 145, 94, 162, 142, 172, 63, 58, 141, 153, 211, 230, 51, 98, 86, 97, 157, 94}
	flag := NodeVerify(publickey, oldBalance_hm, transferAmount_hm, newBalance_hm)
	fmt.Println(flag)
}

func Test_getput(t *testing.T) {
	G1 := []byte{195, 153, 245, 222, 195, 152, 112, 47, 129, 79, 150, 122, 83, 4, 237, 142, 251, 215, 241, 160, 24, 216, 246, 105, 237, 83, 182, 53, 214, 53, 139, 73}

	N1 := []byte{238, 223, 119, 206, 43, 23, 14, 2, 159, 122, 233, 6, 43, 53, 223, 47}
	nsquare1 := []byte{222, 228, 69, 213, 73, 95, 66, 80, 24, 124, 30, 174, 198, 90, 112, 198, 28, 116, 247, 151, 13, 92, 246, 31, 56, 225, 40, 131, 246, 8, 234, 161}
	publickey := Generate_wholenetpublickey(G1, N1, nsquare1)
	fmt.Println(publickey.G)
	fmt.Println(publickey.N)
	fmt.Println(publickey.Nsquare)
	err := PutWholeNetworkPublickey("G:/testgetput.txt", &publickey)
	p, _ := GetWholeNetworkPublickey("G:/testgetput.txt")
	fmt.Println(p.G)
	fmt.Println(p.N)
	fmt.Println(p.Nsquare)
	fmt.Println(err)

}

//func Test_hm_transacton(t *testing.T) {
//	old := big.NewInt(500)
//	old_byte := old.Bytes()
//	suffix_bigint := rand.Prime(rand.Reader, 64)
//	suffix := suffix_bigint.Bytes()

//	oldBalance := make([]byte, 16)
//	oldBalance = append(old_byte, oldBalance[len(old_byte):])
//	oldBalance = append(oldBalance[:(16-len(suffix))], suffix...)

//	transfer := big.NewInt(200)
//	transfer_byte := transfer.Bytes()
//	transferAmount := make([]byte, 8)
//	transferAmount = append(transfer_byte, transferAmount[len(transfer_byte):])

//	//get whole_networkpublickey
//	G1 := []byte{195, 153, 245, 222, 195, 152, 112, 47, 129, 79, 150, 122, 83, 4, 237, 142, 251, 215, 241, 160, 24, 216, 246, 105, 237, 83, 182, 53, 214, 53, 139, 73}
//	N1 := []byte{238, 223, 119, 206, 43, 23, 14, 2, 159, 122, 233, 6, 43, 53, 223}
//	nsquare1 := []byte{222, 228, 69, 213, 73, 95, 66, 80, 24, 124, 30, 174, 198, 90, 112, 198, 28, 116, 247, 151, 13, 92, 246, 31, 56, 225, 40, 131, 246, 8, 234, 161}
//	publickey := Generate_wholenetpublickey(G1, N1, nsquare1)

//	ee := New

//}
//PreTransaction(oldBalance []byte, transferAmount []byte, illegal_balance_hm []byte,
// whole_networkpublickey PaillierPublickey, ecdsa_publickey *ecdsa.PublicKey) (bool, []byte, []byte, []byte)
//newBalance_hm, transferAmount_hm, transferAmount_ecc
//NodeVerify(whole_networkpublickey PaillierPublickey, oldBalance_hm []byte, transferAmount_hm []byte, newBalance_hm []byte) bool
func Test_hmTransaction(t *testing.T){
	paillerkey,_:= Generate_paillierKey()
	wholepublickey := paillerkey.PaillierPublickey
	wholeprivate := paillerkey.PaillierPrivatekey

	fmt.Println(wholeprivate.P.String())
	fmt.Println(wholeprivate.Q.String())
	fmt.Println(wholeprivate.Lambda.String())
	fmt.Println()

	fmt.Println(wholepublickey.N.String())
	fmt.Println(wholepublickey.Nsquare.String())
	fmt.Println(wholepublickey.G.String())
	fmt.Println()

	bign := new(big.Int)
	bign.SetString(wholepublickey.N.String(),10)
	fmt.Println(*wholepublickey.N)
	fmt.Println(*bign)


	//paillerprivate := paillerkey.PaillierPrivatekey

	old := big.NewInt(6000000000000000000)
	old_byte := old.Bytes()
	fmt.Println(len(old_byte))
	trans := big.NewInt(300000000000000)
	trans_byte := trans.Bytes()

	phm := New_Paillier_Hmencryption()
	hmoldBalance,_:= phm.Encrypto_message(&wholepublickey,old_byte)
	realold := phm.Decrypto_Ciphertext(paillerkey,hmoldBalance)
	fmt.Println(hmoldBalance)
	fmt.Println(old_byte)
	fmt.Println(realold)

	//ecdsa_private,_:=crypto.GenerateKey()
	//ecdsa_publickey := ecdsa_private.PublicKey

	//HmillegalBanlance := make([]byte,32)

	index,hmnewbalance,hmTransferAmount:= PreHmTransaction(old_byte,trans_byte,nil,wholepublickey)
	fmt.Println(index)
	fmt.Println(hmnewbalance)
	fmt.Println(hmTransferAmount)
	//fmt.Println(len(transferamounecc))

	testbig := new(big.Int)
	//testbig.SetBytes(transferamounecc)
	ss := testbig.String()
	fmt.Println(len(ss))



	sum,_ :=phm.Calculator(&wholepublickey,"paillier",hmnewbalance,hmTransferAmount)
	desum := phm.Decrypto_Ciphertext(paillerkey,sum)

	fmt.Println(len(sum))
	fmt.Println(desum)


	flag := EqualTwoBytes(hmoldBalance,sum)

	fmt.Println(flag)

}

func Test_Big(t *testing.T){
	old := big.NewInt(500)
	fmt.Println(old)

	old_byte := old.Bytes()
	//将byte[]转为大整数
	big := new(big.Int)
	big.SetBytes(old_byte)
	fmt.Println(big)

	//bigint类型转字符串类型，10进制
	ss := old.String()
	//将字符串类型按10进制转为大整数类型

	bigint,_:=big.SetString(ss,10)
	fmt.Println(ss)
	fmt.Println(bigint)
}


//func Test_createccpub(t *testing.T){
//	x := new(big.Int)
//	y := new(big.Int)
//
//
//
//}

