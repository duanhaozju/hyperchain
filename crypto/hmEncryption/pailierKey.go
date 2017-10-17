//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hmEncryption

import (
	"crypto/rand"
	"math/big"
)

type PaillierKey struct {
	PaillierPrivatekey
	PaillierPublickey
}

type PaillierPublickey struct {
	//N=p*q
	N *big.Int

	//nsquare=n^2
	Nsquare *big.Int

	//G使得 gcd(L(G^lambda mod N^2),N)=1
	G *big.Int
}

type PaillierPrivatekey struct {
	//P,Q are two larger primes
	P *big.Int

	Q *big.Int

	//Lambda=lcm(P-1,Q-1)
	Lambda *big.Int
}

func Generate_paillierKey() (*PaillierKey, error) {
	var publickey PaillierPublickey
	var privatekey PaillierPrivatekey
	var key PaillierKey
	var final_key *PaillierKey
	BigONE := new(big.Int)
	BigONE = BigONE.SetInt64(1)

	p, err1 := rand.Prime(rand.Reader, 32)
	if err1 != nil {
		return nil, err1
	}
	q, err2 := rand.Prime(rand.Reader, 32)
	if err2 != nil {
		return nil, err2
	}

	privatekey.P = p
	privatekey.Q = q

	N := new(big.Int)
	N = N.Mul(p, q)
	publickey.N = N

	//这里G的确定还需要改进
	G := new(big.Int)
	rad, err3 := rand.Prime(rand.Reader, 256)
	//	ggg := rad.Bytes()
	//	fmt.Println(len(ggg))
	if err3 != nil {
		return nil, err3
	}
	publickey.G = G.Set(rad)
	nsquare := new(big.Int)
	publickey.Nsquare = nsquare.Mul(N, N)

	p_1 := new(big.Int)
	q_1 := new(big.Int)
	p_1 = p_1.Sub(p, BigONE)
	q_1 = q_1.Sub(q, BigONE)

	//求最大公约数
	gcd := new(big.Int)
	gcd = gcd.GCD(new(big.Int), new(big.Int), p_1, q_1)

	//求最小公倍数
	mutiply := new(big.Int)
	mutiply = mutiply.Mul(p_1, q_1)
	Lambda := new(big.Int)
	privatekey.Lambda = Lambda.Div(mutiply, gcd)

	//给key赋初值
	key.PaillierPrivatekey = privatekey
	key.PaillierPublickey = publickey

	//check whether G is good
	modpow := new(big.Int)
	modpow = modpow.Exp(key.G, key.Lambda, key.Nsquare)
	modpow = modpow.Sub(modpow, BigONE)
	modpow = modpow.Div(modpow, key.N)
	modpow = modpow.GCD(new(big.Int), new(big.Int), modpow, key.N)
	if modpow.Cmp(BigONE) != 0 {
		//fmt.Println("G is not good,get G again")
		final_key, _ = Generate_paillierKey()
	}
	if modpow.Cmp(BigONE) == 0 {
		final_key = &key
	}

	return final_key, nil
}

func (publickey *PaillierPublickey) Paillier_Encrypto(message *big.Int, rand *big.Int) (*big.Int, error) {
	//make a rand prime use to encryto message
	//	r, err := rand.Prime(rand.Reader, 256)
	//	if err != nil {
	//		return nil, err
	//	}
	modpow_1 := new(big.Int)
	modpow_1 = modpow_1.Exp(publickey.G, message, publickey.Nsquare)

	modpow_2 := new(big.Int)
	modpow_2 = modpow_2.Exp(rand, publickey.N, publickey.Nsquare)

	mutiply := new(big.Int)
	mutiply = mutiply.Mul(modpow_1, modpow_2)

	enmessage := mutiply.Mod(mutiply, publickey.Nsquare)
	return enmessage, nil
}

func (key *PaillierKey) Paillier_Decrypto(em *big.Int) *big.Int {
	BigONE := big.NewInt(1)
	modpow_1 := new(big.Int)
	modpow_1 = modpow_1.Exp(key.G, key.Lambda, key.Nsquare)
	modpow_1 = modpow_1.Sub(modpow_1, BigONE)
	modpow_1 = modpow_1.Div(modpow_1, key.N)
	modpow_1 = modpow_1.ModInverse(modpow_1, key.N)

	modpow_2 := new(big.Int)
	modpow_2 = modpow_2.Exp(em, key.Lambda, key.Nsquare)
	modpow_2 = modpow_2.Sub(modpow_2, BigONE)
	modpow_2 = modpow_2.Div(modpow_2, key.N)

	message := new(big.Int)
	message = message.Mul(modpow_1, modpow_2)
	message = message.Mod(message, key.N)

	return message

}

func (publickey *PaillierPublickey) Paillier_Cipher_add(em1 *big.Int, em2 *big.Int) *big.Int {
	mutiply := new(big.Int)
	mutiply = mutiply.Mul(em1, em2)
	mutiply_modnsquare := new(big.Int)
	mutiply_modnsquare = mutiply_modnsquare.Mod(mutiply, publickey.Nsquare)

	return mutiply_modnsquare
}
