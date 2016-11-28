//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package utils

import "errors"

var (
	// ErrRegistrationRequired Registration to the Membership Service required.
	ErrRegistrationRequired = errors.New("Registration to the Membership Service required.")

	// ErrNotInitialized Initialization required
	ErrNotInitialized = errors.New("Initialization required.")

	// ErrAlreadyInitialized Already initialized
	ErrAlreadyInitialized = errors.New("Already initialized.")

	// ErrAlreadyRegistered Already registered
	ErrAlreadyRegistered = errors.New("Already registered.")

	// ErrTransactionMissingCert Transaction missing certificate or signature
	ErrTransactionMissingCert = errors.New("Transaction missing certificate or signature.")

	// ErrInvalidTransactionSignature Invalid Transaction Signature
	ErrInvalidTransactionSignature = errors.New("Invalid Transaction Signature.")

	// ErrTransactionCertificate Missing Transaction Certificate
	ErrTransactionCertificate = errors.New("Missing Transaction Certificate.")

	// ErrTransactionSignature Missing Transaction Signature
	ErrTransactionSignature = errors.New("Missing Transaction Signature.")

	// ErrInvalidSignature Invalid Signature
	ErrInvalidSignature = errors.New("Invalid Signature.")

	// ErrInvalidKey Invalid key
	ErrInvalidKey = errors.New("Invalid key.")

	// ErrInvalidReference Invalid reference
	ErrInvalidReference = errors.New("Invalid reference.")

	// ErrNilArgument Invalid reference
	ErrNilArgument = errors.New("Nil argument.")

	// ErrNotImplemented Not implemented
	ErrNotImplemented = errors.New("Not implemented.")

	// ErrKeyStoreAlreadyInitialized Keystore already Initilized
	ErrKeyStoreAlreadyInitialized = errors.New("Keystore already Initilized.")

	// ErrEncrypt Encryption failed
	ErrEncrypt = errors.New("Encryption failed.")

	// ErrDecrypt Decryption failed
	ErrDecrypt = errors.New("Decryption failed.")

	// ErrDifferentChaincodeID ChaincodeIDs are different
	ErrDifferentChaincodeID = errors.New("ChaincodeIDs are different.")

	// ErrDifferrentConfidentialityProtocolVersion different confidentiality protocol versions
	ErrDifferrentConfidentialityProtocolVersion = errors.New("Confidentiality protocol versions are different.")

	// ErrInvalidConfidentialityLevel Invalid confidentiality level
	ErrInvalidConfidentialityLevel = errors.New("Invalid confidentiality level")

	// ErrInvalidConfidentialityProtocol Invalid confidentiality level
	ErrInvalidConfidentialityProtocol = errors.New("Invalid confidentiality protocol")

	// ErrInvalidTransactionType Invalid transaction type
	ErrInvalidTransactionType = errors.New("Invalid transaction type")

	// ErrInvalidProtocolVersion Invalid protocol version
	ErrInvalidProtocolVersion = errors.New("Invalid protocol version")
)

// ErrToString converts and error to a string. If the error is nil, it returns the string "<clean>"
func ErrToString(err error) string {
	if err != nil {
		return err.Error()
	}

	return "<clean>"
}
