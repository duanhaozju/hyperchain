// author: Lizhong kuang
// date: 2016-09-29

package crypto

import (
	"testing"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func TestCryptoInitInheritsLoggingLevel(t *testing.T) {
	logging.SetLevel(logging.WARNING, "crypto")

	Init()

	assertCryptoLoggingLevel(t, logging.WARNING)
}

func TestCryptoInitDoesntOverrideLoggingLevel(t *testing.T) {
	logging.SetLevel(logging.WARNING, "crypto")
	viper.Set("logging.crypto", "info")

	Init()

	assertCryptoLoggingLevel(t, logging.WARNING)
}

func assertCryptoLoggingLevel(t *testing.T, expected logging.Level) {
	actual := logging.GetLevel("crypto")

	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
