/**
author:Zhang Kejie
changelog:Test Create Cert.
 */

package primitives

import (
	"testing"
	"fmt"
)

func TestECACert(t *testing.T) {
	der,_,_ := NewSelfSignedCert()
	//fmt.Println("PrivateKey:" + string(key));
	pem := DERCertToPEM(der)
	fmt.Println(string(pem))

}

func TestDecode(t *testing.T){
	der,_,_ := NewSelfSignedCert()
	pem := DERCertToPEM(der)
	cert,err := PEMtoCertificate(pem)
	if(err != nil){
		fmt.Println("Error")
	}
	fmt.Println(cert.Subject.CommonName)
}
