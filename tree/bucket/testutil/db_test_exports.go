package testutil

import (
	"github.com/spf13/viper"
	"os"
	"testing"
)

// TestDBWrapper wraps the db. Can be used by other modules for testing
type TestDBWrapper struct {
	performCleanup bool
}

// NewTestDBWrapper constructs a new TestDBWrapper
func NewTestDBWrapper() *TestDBWrapper {
	return &TestDBWrapper{}
}

///////////////////////////
// Test db creation and cleanup functions

// CleanDB This method closes existing db, remove the db dir.
// Can be called before starting a test so that data from other tests does not interfere
func (testDB *TestDBWrapper) CleanDB(t testing.TB) {
	// cleaning up test db here so that each test does not have to call it explicitly
	// at the end of the test
	// TODO close the db
	// testDB.cleanup()
	testDB.removeDBPath()
	t.Logf("Creating testDB")

	// TODO open the db
	//Start()
	testDB.performCleanup = true
}

func (testDB *TestDBWrapper) cleanup() {

}

func (testDB *TestDBWrapper) removeDBPath() {
	dbPath := viper.GetString("peer.fileSystemPath")
	if dbPath == "" || len(dbPath) == 0 {
		dbPath = "/tmp/hyperchain/cachedb"
	}
	os.RemoveAll(dbPath)
}
