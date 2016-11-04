package certification

import (
	"github.com/op/go-logging"
	"crypto/x509"
	"encoding/pem"
)

var log *logging.Logger // package-level logger
//initialize the peers pool
func init() {
	log = logging.MustGetLogger("certification")
}

func VerifyCert() {

	// Verifying with a custom list of root certificates.
	const rootPEM = `
-----BEGIN CERTIFICATE-----
MIIDGjCCAsGgAwIBAgIBATAKBggqhkjOPQQDAjCBljELMAkGA1UEBhMCQ04xETAP
BgNVBAgMCFpoZWppYW5nMREwDwYDVQQHDAhIYW5nemhvdTETMBEGA1UECgwKaHlw
ZXJjaGFpbjENMAsGA1UECwwEcm9vdDEZMBcGA1UEAwwQcm9vdCBjZXJmaWNpdGlv
bjEiMCAGCSqGSIb3DQEJARYTYWRtaW5AaHlwZXJjaGFpbi5jbjAeFw0xNjExMDQx
MTA4NDdaFw0xNzExMDQxMTA4NDdaMIGBMQswCQYDVQQGEwJDTjERMA8GA1UECAwI
WmhlamlhbmcxEzARBgNVBAoMCmh5cGVyY2hhaW4xETAPBgNVBAsMCHN1YnJvb3Qx
MREwDwYDVQQDDAhzdWJyb290MTEkMCIGCSqGSIb3DQEJARYVc3Vicm9vdEBoeXBl
cmNoYWluLmNuMFYwEAYHKoZIzj0CAQYFK4EEAAoDQgAE5Wd2Sjuu3BIU5c6X1gEh
QaH3u+C/6PxRGuxo5wCGnxTw7UKLeC6TnmA6ckmX78y84cOj3lUnB1jDT8wZoQ+y
UqOCARQwggEQMAkGA1UdEwQCMAAwLAYJYIZIAYb4QgENBB8WHU9wZW5TU0wgR2Vu
ZXJhdGVkIENlcnRpZmljYXRlMB0GA1UdDgQWBBRAhhZUJRLU/rHYc1f/f+Rza3bi
4TCBtQYDVR0jBIGtMIGqoYGcpIGZMIGWMQswCQYDVQQGEwJDTjERMA8GA1UECAwI
WmhlamlhbmcxETAPBgNVBAcMCEhhbmd6aG91MRMwEQYDVQQKDApoeXBlcmNoYWlu
MQ0wCwYDVQQLDARyb290MRkwFwYDVQQDDBByb290IGNlcmZpY2l0aW9uMSIwIAYJ
KoZIhvcNAQkBFhNhZG1pbkBoeXBlcmNoYWluLmNuggkAiKJwcZ0RTgQwCgYIKoZI
zj0EAwIDRwAwRAIgZHbbvhAvrP8k4RUpah9yy1reZSprz+89Vpa8ovJBlSUCIHDa
DePGAyOw6gqKJl/yHQywd9MbRLYmvU7wje5pWZkT
-----END CERTIFICATE-----`

	const certPEM = `
-----BEGIN CERTIFICATE-----
MIICZDCCAgqgAwIBAgIBATAKBggqhkjOPQQDAjCBgTELMAkGA1UEBhMCQ04xETAP
BgNVBAgMCFpoZWppYW5nMRMwEQYDVQQKDApoeXBlcmNoYWluMREwDwYDVQQLDAhz
dWJyb290MTERMA8GA1UEAwwIc3Vicm9vdDExJDAiBgkqhkiG9w0BCQEWFXN1YnJv
b3RAaHlwZXJjaGFpbi5jbjAeFw0xNjExMDQxMTI3MzJaFw0xNzExMDQxMTI3MzJa
MHsxCzAJBgNVBAYTAkNOMREwDwYDVQQIDAhaaGVqaWFuZzETMBEGA1UECgwKaHlw
ZXJjaGFpbjEQMA4GA1UECwwHc3Vicm9vdDEOMAwGA1UEAwwFbm9kZTExIjAgBgkq
hkiG9w0BCQEWE25vZGUxQGh5cGVyY2hhaW4uY24wVjAQBgcqhkjOPQIBBgUrgQQA
CgNCAARdMBJ5gVF1yAmYRHca2uOtELyktCak4N9oT+gz5J+Cu3hVOEPw6QC50hck
l7fvTndIPHMzdM3DmxApglXaSrD+o3sweTAJBgNVHRMEAjAAMCwGCWCGSAGG+EIB
DQQfFh1PcGVuU1NMIEdlbmVyYXRlZCBDZXJ0aWZpY2F0ZTAdBgNVHQ4EFgQUL6Fd
0wWlt9QEDtwMTHRgiff2TKowHwYDVR0jBBgwFoAUQIYWVCUS1P6x2HNX/3/kc2t2
4uEwCgYIKoZIzj0EAwIDSAAwRQIhAKv+eZg/2HQ+2QbCYE/xipXd/H9K/9dwvWI7
TD6ZwVOdAiAWwN21LQ1ixkFi/TUpxuFMXzfwcfhQTOeEE7SIxEqY8g==
-----END CERTIFICATE-----`

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}

	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	opts := x509.VerifyOptions{
		DNSName: "mail.google.com",
		Roots:   roots,
	}

	if _, err := cert.Verify(opts); err != nil {
		panic("failed to verify certificate: " + err.Error())
	}
}
