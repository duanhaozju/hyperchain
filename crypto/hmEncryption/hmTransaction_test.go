//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hmEncryption

import (
	//"crypto/ecdsa"
	//"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	//"github.com/hyperchain/hyperchain/crypto"
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
func Test_hmTransaction(t *testing.T) {
	paillerkey, _ := Generate_paillierKey()
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
	bign.SetString(wholepublickey.N.String(), 10)
	fmt.Println(*wholepublickey.N)
	fmt.Println(*bign)

	//paillerprivate := paillerkey.PaillierPrivatekey

	old := big.NewInt(100)
	old_byte := old.Bytes()
	fmt.Println(len(old_byte))
	trans := big.NewInt(10)
	trans_byte := trans.Bytes()

	phm := New_Paillier_Hmencryption()
	hmoldBalance, _ := phm.Encrypto_message(&wholepublickey, old_byte)
	realold := phm.Decrypto_Ciphertext(paillerkey, hmoldBalance)
	fmt.Println(hmoldBalance)
	fmt.Println(old_byte)
	fmt.Println(realold)

	//ecdsa_private,_:=crypto.GenerateKey()
	//ecdsa_publickey := ecdsa_private.PublicKey

	//HmillegalBanlance := make([]byte,32)

	index, hmnewbalance, hmTransferAmount := PreHmTransaction(old_byte, trans_byte, nil, wholepublickey)
	fmt.Println(index)
	fmt.Println(hmnewbalance)
	fmt.Println(hmTransferAmount)
	//fmt.Println(len(transferamounecc))
	fmt.Println()

	testbig := new(big.Int)
	testbig.SetBytes(hmnewbalance)
	fmt.Println(testbig)
	fmt.Println(testbig.BitLen())

	sum, _ := phm.Calculator(&wholepublickey, "paillier", hmnewbalance, hmTransferAmount)
	desum := phm.Decrypto_Ciphertext(paillerkey, sum)

	fmt.Println(len(sum))
	fmt.Println(desum)

	flag := EqualTwoBytes(hmoldBalance, sum)

	fmt.Println(flag)

}

func Test_hmencryption(t *testing.T) {
	//keypair := new(PaillierKey)
	private := new(PaillierPrivatekey)
	public := new(PaillierPublickey)
	private.P, _ = new(big.Int).SetString("4168938673", 10)
	private.Q, _ = new(big.Int).SetString("3433902229", 10)
	private.Lambda, _ = new(big.Int).SetString("132553035131260752", 10)

	public.N, _ = new(big.Int).SetString("14315727801779002117", 10)
	public.G, _ = new(big.Int).SetString("90976693534933209671098397317966944738726332459523400324197777885595356310417", 10)
	public.Nsquare, _ = new(big.Int).SetString("204940062494628260128356353732290481689", 10)
	fmt.Println(len("204940062494628260128356353732290481689"))
	old := big.NewInt(100)
	old_byte := old.Bytes()
	//fmt.Println(len(old_byte))
	trans := big.NewInt(90)
	trans_byte := trans.Bytes()

	phm := New_Paillier_Hmencryption()
	hmoldBalance, _ := phm.Encrypto_message(public, old_byte)
	hmoldbig := new(big.Int).SetBytes(hmoldBalance)
	fmt.Println(hmoldbig)

	index, hmnewbalance, hmTransferAmount := PreHmTransaction(old_byte, trans_byte, nil, *public)

	sum, _ := phm.Calculator(public, "paillier", hmnewbalance, hmTransferAmount)

	hmsum := new(big.Int).SetBytes(sum)

	hmnew := new(big.Int).SetBytes(hmnewbalance)
	hmtrf := new(big.Int).SetBytes(hmTransferAmount)

	fmt.Println(index)
	fmt.Println(hmnew)
	fmt.Println(hmtrf)
	fmt.Println(hmsum)

}

func Test_Big(t *testing.T) {
	transfer := new(big.Int)

	transfer.SetString("183003357837956770250501280563466144409", 10)
	fmt.Println(transfer)
	fmt.Println(len(transfer.String()))
	remain, _ := new(big.Int).SetString("104421979747789724691932366515971813173", 10)
	fmt.Println(remain)
	fmt.Println(len(remain.String()))
	mul := new(big.Int)
	mul = mul.Mul(transfer, remain)

	nsquare, _ := new(big.Int).SetString("204940062494628260128356353732290481689", 10)

	mod := new(big.Int)
	mod = mod.Mod(mul, nsquare)
	fmt.Println("201210379221173114992221939430310283117")
	fmt.Println(len("201210379221173114992221939430310283117"))
	fmt.Println(mul)
	fmt.Println(len(mul.String()))
	fmt.Println(mod)
}
