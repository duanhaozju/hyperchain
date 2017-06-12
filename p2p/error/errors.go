package error

import "reflect"

const (
	CertPemParseFaild = iota
	ECertVerifyFailed
	RCertVerifyFailed
	SignVerifyFailed
)

func New(errType int) error {
	switch errType {
	case SignVerifyFailed:
		return &ErrCertSignVerifyFailed{}
	case CertPemParseFaild:
		return &ErrCertPemParseFailed{}
	case ECertVerifyFailed:
		return &ErrECertVerifyFailed{}
	case RCertVerifyFailed:
		return &ErrRCertVerifyFailed{}
	default:
		return &ErrUnknown{}
	}
}

func Campare(origin interface{}, shouldBe int) bool {
	if origin == nil {
		return false
	}
	return reflect.TypeOf(origin) == reflect.TypeOf(New(shouldBe))
}

type ErrCertPemParseFailed struct{}

func (e ErrCertPemParseFailed) Error() string {
	return "Cert PEM parse failed."
}

type ErrECertVerifyFailed struct{}

func (e ErrECertVerifyFailed) Error() string {
	return "ECert Verify Failed."
}

type ErrRCertVerifyFailed struct{}

func (e ErrRCertVerifyFailed) Error() string {
	return "RCert Verify failed."
}

type ErrCertSignVerifyFailed struct{}

func (e ErrCertSignVerifyFailed) Error() string {
	return "Cert Signature Verified failed."
}

type ErrUnknown struct{}

func (e ErrUnknown) Error() string {
	return "unKnown error."
}
