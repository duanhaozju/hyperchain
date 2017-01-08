/**
author:张珂杰
date:16-12-18
verify ecert and signature
 */

package primitives

import (
	"io/ioutil"
	"crypto/x509"
	"encoding/pem"
	//"fmt"
	"fmt"
	//"sync"
	"github.com/pkg/errors"
)

//读取config文件
func GetConfig(path string) (string,error){
	content,err := ioutil.ReadFile(path)

	if err!=nil {
		//fmt.Println()
		return "",err
	}

	return string(content),nil

}

//解析证书
func ParseCertificate(ECert string) (*x509.Certificate,error){
	block,_ := pem.Decode([]byte(ECert))

	if block==nil {
		fmt.Println("failed to parse certificate PEM")
		return nil,errors.New("failed to parse certificate PEM")
	}

	cert,err := x509.ParseCertificate(block.Bytes)

	if err!=nil{
		fmt.Println("faile to parse certificate")
		return nil,errors.New("faile to parse certificate")
	}

	return cert,nil
}

//验证证书中的签名
func VerifySignature(cert *x509.Certificate,ca *x509.Certificate) bool{

	err:=cert.CheckSignatureFrom(ca)
	if err==nil{
		return true
	}else {
		return false
	}
}


func ParseKey(derPri string)(interface{},error){
	block,_ := pem.Decode([]byte(derPri))


	pri,err1 := DERToPrivateKey(block.Bytes)

	if err1!=nil{
		return nil,err1
	}

	return pri,nil
}

func ParsePubKey(pubstr string)(interface{},error){
	//todo finish the public key parse
	return nil,nil

}
