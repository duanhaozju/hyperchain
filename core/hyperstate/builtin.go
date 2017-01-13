package hyperstate

import (
	"github.com/op/go-logging"
	"hyperchain/crypto"
)

var (
	log        *logging.Logger // package-level logger
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

func init() {
	log = logging.MustGetLogger("state")
}
