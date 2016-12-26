/**
author:Zhang Kejie
log:Test Create Cert.
 */

package primitives

import (
	"testing"
	"fmt"
	"encoding/asn1"
	"math/big"
	"crypto/x509/pkix"
	"time"
	//"encoding/pem"
	//"crypto/x509"
	"io/ioutil"
	//"crypto/ecdsa"
	"encoding/pem"
	"crypto/x509"
	//"os"
	//"encoding/json"
	"os"
)

/*const  ECert  = `-----BEGIN CERTIFICATE-----
MIICPzCCAcSgAwIBAgIBATAKBggqhkjOPQQDAzBQMRMwEQYDVQQKEwpIeXBlcmNo
YWluMRswGQYDVQQDExJ0ZXN0Lmh5cGVyY2hhaW4uY24xDzANBgNVBCoTBkdvcGhl
cjELMAkGA1UEBhMCWkgwHhcNMTYxMjE4MTAzMjE3WhcNMTcxMjE4MTEzMjE3WjBQ
MRMwEQYDVQQKEwpIeXBlcmNoYWluMRswGQYDVQQDExJ0ZXN0Lmh5cGVyY2hhaW4u
Y24xDzANBgNVBCoTBkdvcGhlcjELMAkGA1UEBhMCWkgwdjAQBgcqhkjOPQIBBgUr
gQQAIgNiAASwVJ1VpSDxQexQMpsHnNAHEJzZ26+G2+EVyN1ZJY1UyR7aqbl4oHC+
OV0IsmD3AI69HjpeHIN5AmQi1wdrsgwIIUYfigeO6oEo+HX2YqY27MfRgzfGC8jo
wjHlOfbLvpWjcjBwMA4GA1UdDwEB/wQEAwICBDAmBgNVHSUEHzAdBggrBgEFBQcD
AgYIKwYBBQUHAwEGAioDBgOBCwEwDwYDVR0TAQH/BAUwAwEB/zANBgNVHQ4EBgQE
AQIDBDAWBgMqAwQED2V4dHJhIGV4dGVuc2lvbjAKBggqhkjOPQQDAwNpADBmAjEA
u+A71SOUNIZ24osq5/Qvn9p5e0i7w95mAvXkPgIkB0CIImb7I4MJNDzhP6PmjN0w
AjEA9Zib3NJ0uIRIE5oylQAqPUuQA0eNMnf2qWuQwxc7hAHBtw9lNkK0/mj2SJbH
YNjt
-----END CERTIFICATE-----`*/


type dsaSignature struct {
	R, S *big.Int
}

type ecdsaSignature dsaSignature

type validity struct {
	NotBefore, NotAfter time.Time
}


type publicKeyInfo struct {
	Raw       asn1.RawContent
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}

type tbsCertificate struct {
	Raw                asn1.RawContent
	Version            int `asn1:"optional,explicit,default:0,tag:0"`
	SerialNumber       *big.Int
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Issuer             asn1.RawValue
	Validity           validity
	Subject            asn1.RawValue
	PublicKey          publicKeyInfo
	UniqueId           asn1.BitString   `asn1:"optional,tag:1"`
	SubjectUniqueId    asn1.BitString   `asn1:"optional,tag:2"`
	Extensions         []pkix.Extension `asn1:"optional,explicit,tag:3"`
}

func TestECACert(t *testing.T) {
	der,_,_ := NewSelfSignedCert()
	//fmt.Println("PrivateKey:" + string(key));
	pem := DERCertToPEM(der)
	file,_ := os.Create("rcert.cert")
	file.WriteString(string(pem))
	//fmt.Println(string(pem))
}

func TestDecode(t *testing.T){

	fileContent,err := ioutil.ReadFile("../../../config/cert/server/eca.cert")

	if(err!=nil){
		panic(err)
	}


	//fmt.Println([]byte(fileContent))

	ECert := string(fileContent)

	block,_ := pem.Decode([]byte(ECert))

	cert1,_ := x509.ParseCertificate(block.Bytes)

	/*if err==nil {
		fmt.Println(cert1.NotAfter.String())
	}*/
	//ecd := new(ECDSASignature)
	//pub,ok := cert1.PublicKey.(*(ecdsa.PublicKey))
	//if ok {
		//fmt.Printf("!!!!!!!!!!")
	//}
	//check := ecdsa.Verify(pub,cert1.RawTBSCertificate,ecd.R,ecd.S)
	//fmt.Println(check)
	err1 := cert1.CheckSignature(cert1.SignatureAlgorithm,cert1.RawTBSCertificate,cert1.Signature)
	fmt.Println(err1)
	//der,_,_ := NewSelfSignedCert()

	//spem := DERCertToPEM(der)

	//asnDer,_ := DERToX509Certificate(der)
	//cert := *(x509.ParseCertificate(asnDer))

	//var cert *x509.
	//fmt.Println(big.Int(cert.PublicKey.PublicKey.Bytes));

	//pem := DERCertToPEM(der)
	//cert,err := PEMtoCertificate(pem)
	//if(err != nil){
	//	fmt.Println("Error")
	//}
	//fmt.Println(cert.RawSubjectPublicKeyInfo)
}

func TestTime(t *testing.T)  {
	time1 := time.Now().Add(8760*time.Hour).String()
	fmt.Printf(time1)
}

func TestKey(t *testing.T){
	pri,_ := NewECDSAKey()

	fmt.Println(pri)

	//fmt.Println(json)
	var block pem.Block
	block.Type="ECDSA PRIVATE KEY"
	der,_ := PrivateKeyToDER(pri)
	block.Bytes = der
	file,_ := os.Create("a.priv")
	pem.Encode(file,&block)
	//fmt.Println(*pri)
}

func TestParseKey(t *testing.T){
	content,_ := ioutil.ReadFile("../../../config/cert/server/eca.priv")

	privateKey := string(content)

	block,_ := pem.Decode([]byte(privateKey));

	//var pri ecdsa.PrivateKey

	pri,err := DERToPrivateKey(block.Bytes)

	if err != nil {
		fmt.Println(err)
	}else {
		fmt.Println(pri)
	}

}


//测试签名
func TestPayload(t *testing.T){

	ee := NewEcdsaEncrypto("ecdsa")
	payload := []byte{1,2,3}

	content,_ := GetConfig("../../../config/cert/server/eca.priv")


	pri,_ := ParseKey(content)


	//fmt.Println(pri)
	sign,_ := ee.Sign(payload,pri)

	//fmt.Println(sign)

	ECert,_ := GetConfig("../../../config/cert/server/eca.cert")

	//block1,_ := pem.Decode([]byte(ECert))

	cert1 := ParseCertificate(ECert)

	pub := cert1.PublicKey
	//fmt.Println(pub)
//s256:= crypto.NewKeccak256Hash("Keccak256")
//	hash2 := s256.Hash(payload)
	//result,err:=ee.VerifySign(pub,payload,sign)
	result,err := ECDSAVerify(pub,payload,sign)

	if err!=nil {
		fmt.Println(err)
	}else {
		fmt.Println(result)
	}

}

func TestGetCert(t *testing.T){
	cert,err := GetConfig("../../../config/cert/server/eca.cert1")

	if err != nil{
		fmt.Println(err)
		fmt.Println("1231231")
		return
	}
	byteCert := []byte(cert)

	//fmt.Println(byteCert)

	ecrt := ParseCertificate(string(byteCert))

	//fmt.Println(ecrt)

	bol := VerifySignature(ecrt)

	fmt.Println(bol)
}

