//用于对外暴露加密算法
package encrypt

import (
	"crypto/rand"
	"crypto/dsa"
	"fmt"
	"encoding/hex"
	"encoding/json"
)

/**
	DSA数字签名算法
	一、算法中应用了下述参数：
	代码:
	p：一个素模数，其值满足: 2^(L-1) < p < 2^L，其中L是64的倍数，且满足 512 ≤ L ≤ 1024 。
	q：(p-1) 的素因子，其值满足 2^159 < q < 2^160 。
	g：g = powm(h, (p-1)/q, p) 。h为满足1< h < p - 1 的任意整数，从而有 powm(h, (p-1)/q, p) > 1 。
	x：私钥。 x为一个随机或伪随机生成的整数，其值满足 0 < x < q 。
	y：公钥。 y = powm(g, x, p)  。
	注：
	1、整数 p, q, g 可以公开，也可仅由一组特定用户共享。
	2、私钥 x 和 公钥 y 称为一个密钥对 (x,y) ， 私钥只能由签名者本人独自持有，公钥则可以公开发布。密钥对可以在一段时间内持续使用。
	3、符号约定： powm(x,y,z)表示 (x^y) mod z, 即 (x的y次方) 模 z 。"mod"为求模运算符; "^" 为幂运算符; 下文的"*"为乘法运算符。

 */

//签名算法
// privateKey 私钥, toSignData 需要加密的数据 ,返回一个数字签名sign和datahash 通过统一hash算法得到的值
func Sign(privateKey dsa.PrivateKey, toSignData []byte) (sign Signature,datahash []byte){
	// 签名算法
	// 签名算法为dsa，该算法要求提供待签名的信息 M （或者M的hash） 要求采用one-way Hash算法，在这里采用的是md5
	// 签名产生过程如下：
	// 1、产生一个随机数k，其值满足0< k < q 。
	// 2、计算 r := powm(g, k, p) mod q ，其值满足 r >0 。
	// 3、计算 s := ( k^(-1) ( SHA(M) + x*r) ) mod q  ，其值满足 s >0 。
	// 注：
	// 1、k^(-1) 表示整数k关于某个模数的逆元，并非指 k 的倒数。k在每次签名时都要重新生成。逆元：满足 (a*b) mod m =1 的a和b互为关于模数 m 的逆元，表示为 a=b^(-1) 或 b=a^(-1) , 如 (2*5) mod 3 = 1 ， 则 2 和 5 互为 模数3 的逆元。
	// 2、SHA(M )：M的Hash值，M为待签署的明文或明文的Hash值。SHA是Oneway-Hash函数，DSS中选用SHA1( FIPS180: Secure Hash Algorithm )，此时SHA(M) 为160bits长的数字串（16进制），可以参与数学计算（事实上SHA1函数生成的是字符串，因此必须把它转化为数字串）。
	// 3、最终的签名就是整数对( r , s )， 它们和 M 一起发送到验证方。
	// 4、尽管 r 和 s 为 0 的概率相当小，但只要有任何一个为0, 必须重新生成 k ，并重新计算 r 和 s 。

	// 1. 先取得一个hash
	signhash := GetHash(toSignData)

	//2. 用dsa算法进行签名，返回r,s整数对
	r, s, err := dsa.Sign(rand.Reader, &privateKey, signhash)
	if err != nil {
		fmt.Println(err)
	}

	signature := newSignature(*r,*s)

	//fmt.Printf("数字签名:\n%v\n",signature)
	return signature,signhash
}
// 签名验证方法
// publicKey 为公钥，toVerifiedMessage 需要检查的数据，sign 同时携带的数字签名
func Verify(publicKey dsa.PublicKey,toVerifiedMessage []byte,sign Signature) bool{
	//验证签名过程
	//我们用 ( r', s', M' ) 来表示验证方通过某种途径获得的签名结果，之所以这样表示是因为你不能保证你这个签名结果一定是发送方生成的真签名，相反有可能被人窜改过，甚至掉了包。为了描述简便，下面仍用 (r ,s ,M) 代替 (r', s', M') 。
	//为了验证( r, s, M )的签名是否确由发送方所签，验证方需要有 (g, p, q, y) ，验证过程如下：
	//
	// 1、计算 w := s^(-1) mod q
	// 2、计算 u1 := ( SHA( M ) * w ) mod q
	// 3、计算 u2 := ( r * w ) mod q
	// 4、计算 v := (( (g^u1) * (y^u2) ) mod p ) mod q
	// =( (g^u1 mod p)*(y^u2 mod p)  mod p ) mod q
	// =( powm(g, u1, p) * powm(y, u2, p) mod p ) mod q
	// 5、若 v 等于 r，则通过验证，否则验证失败。
	// 注：
	// 1、验证通过说明： 签名( r, s )有效，即(r , s , M )确为发送方的真实签名结果，真实性可以高度信任， M 未被窜改，为有效信息。
	// 2、验证失败说明：签名( r , s )无效，即(r, s , M) 不可靠，或者M被窜改过，或者签名是伪造的，或者M的签名有误，M 为无效信息。

	verifystatus := dsa.Verify(&publicKey, toVerifiedMessage, &sign.R, &sign.S)
	//fmt.Println(verifystatus) // should be true
	return verifystatus
}


func main() {
	//生成私钥
	privatekey := GetPrivateKey()
	fmt.Println("私钥 :")

	data,_ := json.Marshal(privatekey)

	hexdata := hex.EncodeToString(data)
	dehexdata,_:= hex.DecodeString(hexdata)

	fmt.Printf("%#v \n%#v\n%#v\n", privatekey,hex.EncodeToString(data),dehexdata)

	var p dsa.PrivateKey

	json.Unmarshal(dehexdata, &p)
	fmt.Printf("%#v \n", p)


	//取得公钥
	publickey := GetPublicKey(privatekey)
	fmt.Println("公钥:")
	fmt.Printf("%x \n",publickey)

	signature,signhash := Sign(privatekey,[]byte("123123"))
	fmt.Println(Verify(publickey,signhash,signature)) // should be true
}