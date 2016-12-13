package hyperstate

import (
	"hyperchain/crypto"
	"github.com/op/go-logging"
)

var (
	log *logging.Logger // package-level logger
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

func init() {
	log = logging.MustGetLogger("state")
}
