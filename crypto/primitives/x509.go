//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
//lastchange:Zhangkejie
//changelog:modify the some cert's Messages.
//date:2016-12-07
package primitives

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"math/big"
	//"net"
	"time"

	"github.com/hyperchain/hyperchain/common"
	"os"
)

var (
	// TCertEncTCertIndex oid for TCertIndex
	TCertEncTCertIndex = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7}

	// TCertEncEnrollmentID is the ASN1 object identifier of the TCert index.
	TCertEncEnrollmentID = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 8}

	// TCertEncAttributesBase is the base ASN1 object identifier for attributes.
	// When generating an extension to include the attribute an index will be
	// appended to this Object Identifier.
	TCertEncAttributesBase = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6}

	// TCertAttributesHeaders is the ASN1 object identifier of attributes header.
	TCertAttributesHeaders = asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 9}
)

// DERToX509Certificate converts der to x509
func DERToX509Certificate(asn1Data []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(asn1Data)
}

// PEMtoCertificate converts pem to x509
func PEMtoCertificate(raw []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("No PEM block available")
	}

	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, errors.New("Not a valid CERTIFICATE PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// PEMtoDER converts pem to der
func PEMtoDER(raw []byte) ([]byte, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("No PEM block available")
	}

	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, errors.New("Not a valid CERTIFICATE PEM block")
	}

	return block.Bytes, nil
}

// PEMtoCertificateAndDER converts pem to x509 and der
func PEMtoCertificateAndDER(raw []byte) (*x509.Certificate, []byte, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, nil, errors.New("No PEM block available")
	}

	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, nil, errors.New("Not a valid CERTIFICATE PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, block.Bytes, nil
}

// DERCertToPEM converts der to pem
func DERCertToPEM(der []byte) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: der,
		},
	)
}

// GetCriticalExtension returns a requested critical extension. It also remove it from the list
// of unhandled critical extensions
func GetCriticalExtension(cert *x509.Certificate, oid asn1.ObjectIdentifier) ([]byte, error) {
	for i, ext := range cert.UnhandledCriticalExtensions {
		if common.IntArrayEquals(ext, oid) {
			cert.UnhandledCriticalExtensions = append(cert.UnhandledCriticalExtensions[:i], cert.UnhandledCriticalExtensions[i+1:]...)

			break
		}
	}

	for _, ext := range cert.Extensions {
		if common.IntArrayEquals(ext.Id, oid) {
			return ext.Value, nil
		}
	}

	return nil, errors.New("Failed retrieving extension.")
}

// NewSelfSignedCert create a self signed certificate
func NewSelfSignedCert() ([]byte, interface{}, error) {
	privKey, err := NewECDSAKey()

	//储存privateKey
	var block pem.Block
	block.Type = "ECDSA PRIVATE KEY"
	der, _ := PrivateKeyToDER(privKey)
	block.Bytes = der
	file, _ := os.Create("tca.priv")
	pem.Encode(file, &block)
	//--------------------------

	if err != nil {
		return nil, nil, err
	}

	testExtKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	testUnknownExtKeyUsage := []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}
	extraExtensionData := []byte("extra extension")
	commonName := "hyperchain.cn"
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Hyperchain"},
			Country:      []string{"CHN"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  []int{2, 5, 4, 42},
					Value: "Develop",
				},
				// This should override the Country, above.
				{
					Type:  []int{2, 5, 4, 6},
					Value: "ZH",
				},
			},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(876000 * time.Hour), //暂定证书有效期为100年

		SignatureAlgorithm: x509.ECDSAWithSHA384,

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageCertSign,

		ExtKeyUsage:        testExtKeyUsage,
		UnknownExtKeyUsage: testUnknownExtKeyUsage,

		BasicConstraintsValid: true,
		IsCA: true,

		//OCSPServer:            []string{"http://ocsp.example.com"},
		//IssuingCertificateURL: []string{"http://crt.example.com/ca1.crt"},

		//DNSNames:       []string{"test.example.com"},
		//EmailAddresses: []string{"gopher@golang.org"},
		//IPAddresses:    []net.IP{net.IPv4(127, 0, 0, 1).To4(), net.ParseIP("2001:4860:0:2001::68")},

		//PolicyIdentifiers:   []asn1.ObjectIdentifier{[]int{1, 2, 3}},
		//PermittedDNSDomains: []string{".example.com", "example.com"},

		//CRLDistributionPoints: []string{"http://crl1.example.com/ca1.crl", "http://crl2.example.com/ca1.crl"},

		ExtraExtensions: []pkix.Extension{
			{
				Id:    []int{1, 2, 3, 4},
				Value: extraExtensionData,
			},
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, err
	}

	return cert, privKey, nil
}

func createCertByCa(ca *x509.Certificate, private interface{}) ([]byte, interface{}, error) {

	caPri := private.(*ecdsa.PrivateKey)
	privKey, err := NewECDSAKey()

	//储存privateKey
	var block pem.Block
	block.Type = "ECDSA PRIVATE KEY"
	der, _ := PrivateKeyToDER(privKey)
	block.Bytes = der
	file, _ := os.Create("tcert.priv")
	pem.Encode(file, &block)
	//--------------------------

	if err != nil {
		return nil, nil, err
	}

	testExtKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	testUnknownExtKeyUsage := []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}
	extraExtensionData := []byte("extra extension")
	commonName := "hyperchain.cn"
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Hyperchain"},
			Country:      []string{"CHN"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  []int{2, 5, 4, 42},
					Value: "Develop",
				},
				// This should override the Country, above.
				{
					Type:  []int{2, 5, 4, 6},
					Value: "ZH",
				},
			},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(876000 * time.Hour), //暂定证书有效期为100年

		SignatureAlgorithm: x509.ECDSAWithSHA384,

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageCertSign,

		ExtKeyUsage:        testExtKeyUsage,
		UnknownExtKeyUsage: testUnknownExtKeyUsage,

		BasicConstraintsValid: true,
		IsCA: true,

		ExtraExtensions: []pkix.Extension{
			{
				Id:    []int{1, 2, 3, 4},
				Value: extraExtensionData,
			},
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, ca, &privKey.PublicKey, caPri)
	if err != nil {
		return nil, nil, err
	}

	return cert, privKey, nil
}

func GenTCert(ca *x509.Certificate, privatekey interface{}, publicKey interface{}) ([]byte, error) {
	return generTcert(ca, privatekey, publicKey)
}

func generTcert(ca *x509.Certificate, private interface{}, publicKey interface{}) ([]byte, error) {
	caPri := private.(*ecdsa.PrivateKey)
	//privKey, err := NewECDSAKey()

	testExtKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	testUnknownExtKeyUsage := []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}
	extraExtensionData := []byte("extra extension")
	commonName := "hyperchain.cn"
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Hyperchain"},
			Country:      []string{"CHN"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  []int{2, 5, 4, 42},
					Value: "Develop",
				},
				// This should override the Country, above.
				{
					Type:  []int{2, 5, 4, 6},
					Value: "ZH",
				},
			},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(876000 * time.Hour), //暂定证书有效期为100年

		SignatureAlgorithm: x509.ECDSAWithSHA384,

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageCertSign,

		ExtKeyUsage:        testExtKeyUsage,
		UnknownExtKeyUsage: testUnknownExtKeyUsage,

		BasicConstraintsValid: true,
		IsCA: false,

		ExtraExtensions: []pkix.Extension{
			{
				Id:    []int{1, 2, 3, 4},
				Value: extraExtensionData,
			},
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, ca, publicKey, caPri)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// CheckCertPKAgainstSK checks certificate's publickey against the passed secret key
func CheckCertPKAgainstSK(x509Cert *x509.Certificate, privateKey interface{}) error {
	switch pub := x509Cert.PublicKey.(type) {
	case *rsa.PublicKey:
		priv, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return errors.New("Private key type does not match public key type")
		}
		if pub.N.Cmp(priv.N) != 0 {
			return errors.New("Private key does not match public key")
		}
	case *ecdsa.PublicKey:
		priv, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return errors.New("Private key type does not match public key type")

		}
		if pub.X.Cmp(priv.X) != 0 || pub.Y.Cmp(priv.Y) != 0 {
			return errors.New("Private key does not match public key")
		}
	default:
		return errors.New("Unknown public key algorithm")
	}

	return nil
}

// CheckCertAgainRoot check the validity of the passed certificate against the passed certPool
func CheckCertAgainRoot(x509Cert *x509.Certificate, certPool *x509.CertPool) ([][]*x509.Certificate, error) {
	opts := x509.VerifyOptions{
		// TODO		DNSName: "test.example.com",
		Roots: certPool,
	}

	return x509Cert.Verify(opts)
}
