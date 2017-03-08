package core

import (
	"github.com/op/go-logging"
)

var (
	log   *logging.Logger // package-level logger
)
func init() {
	log = logging.MustGetLogger("core")
}
