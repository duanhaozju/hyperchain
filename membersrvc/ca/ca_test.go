// author: Lizhong kuang
// date: 2016-09-29
package ca

import (
	"io/ioutil"
	"os"
	"testing"

	"database/sql"

	"github.com/spf13/viper"
	"hyperchain/core/crypto"
	"hyperchain/core/crypto/primitives"
)

const (
	name = "TestCA"
)

var (
	ca      *CA
	caFiles = [4]string{name + ".cert", name + ".db", name + ".priv", name + ".pub"}
)

func TestNewCA(t *testing.T) {

	//init the crypto layer
	if err := crypto.Init(); err != nil {
		t.Errorf("Failed initializing the crypto layer [%s]", err)
	}

	//initialize logging to avoid panics in the current code
	LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr, os.Stdout)
	CacheConfiguration() // Cache configuration
	//Create new CA
	ca := NewCA(name, initializeTables)
	if ca == nil {
		t.Error("could not create new CA")
	}

	missing := 0
	//check to see that the expected files were created
	for _, file := range caFiles {
		if _, err := os.Stat(ca.path + "/" + file); err != nil {
			missing++
			t.Logf("failed to find file [%s]", file)
		}
	}

	if missing > 0 {
		t.FailNow()
	}

	//check CA certificate for correct properties
	pem, err := ioutil.ReadFile(ca.path + "/" + name + ".cert")
	if err != nil {
		t.Fatalf("could not read CA X509 certificate [%s]", name+".cert")
	}

	cacert, err := primitives.PEMtoCertificate(pem)
	if err != nil {
		t.Fatalf("could not parse CA X509 certificate [%s]", name+".cert")
	}

	//check that commonname, organization and country match config
	org := viper.GetString("pki.ca.subject.organization")
	if cacert.Subject.Organization[0] != org {
		t.Fatalf("ca cert subject organization [%s] did not match configuration [%s]",
			cacert.Subject.Organization, org)
	}

	country := viper.GetString("pki.ca.subject.country")
	if cacert.Subject.Country[0] != country {
		t.Fatalf("ca cert subject country [%s] did not match configuration [%s]",
			cacert.Subject.Country, country)
	}

}

// Empty initializer for CA
func initializeTables(db *sql.DB) error {
	return nil
}
