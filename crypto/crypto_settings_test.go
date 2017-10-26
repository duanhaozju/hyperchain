//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"testing"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func TestCryptoInitInheritsLoggingLevel(t *testing.T) {
	logging.SetLevel(logging.WARNING, "crypto")

	assertCryptoLoggingLevel(t, logging.WARNING)
}

func TestCryptoInitDoesntOverrideLoggingLevel(t *testing.T) {
	logging.SetLevel(logging.WARNING, "crypto")
	viper.Set("logging.crypto", "info")

	assertCryptoLoggingLevel(t, logging.WARNING)
}

func assertCryptoLoggingLevel(t *testing.T, expected logging.Level) {
	actual := logging.GetLevel("crypto")

	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
