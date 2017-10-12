package rbft

import (
	"github.com/facebookgo/ensure"
	"testing"
	"time"
)

func TestPbftImpl_func2(t *testing.T) {
	pbftList := CreatPBFT(t, 4, "./Testdatabase/", "../../configuration/namespaces/", "global", nil)
	time.Sleep(2 * time.Second)
	ensure.DeepEqual(t, 0, int(pbftList[1].status.inRecovery))
}
