package common

import (
	"github.com/op/go-logging"
	"math/rand"
	"time"
)

var (
	logger *logging.Logger // package-level logger
)

func init() {
	logger = logging.MustGetLogger("utils.common")
	rand.Seed(time.Now().UnixNano())
}
