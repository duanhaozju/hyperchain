package network

import (
	"github.com/op/go-logging"
	"hyperchain/common"
)

var logger *logging.Logger

func init(){
	logger = common.GetLogger(common.DEFAULT_LOG,"hypernet")
}
