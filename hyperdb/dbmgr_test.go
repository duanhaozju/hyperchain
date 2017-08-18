//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package hyperdb

import (
	"testing"
	"hyperchain/common"
)

func init()  {
	common.InitHyperLoggerManager(common.NewRawConfig())
}

func TestInitDBMgr(t *testing.T) {

	conf := common.NewRawConfig()

	InitDBMgr(conf)

}
