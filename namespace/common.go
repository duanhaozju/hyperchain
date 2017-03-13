//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"hyperchain/common"
	"os"
)

const (
	NAMESPACE_CONFIG_DIR_ROOT = "gloable.nsConfigRootPath"
)

//ConstructConfigFromDir read all info needed by
func ConstructConfigFromDir(dir os.FileInfo) *common.Config {
	//TODO: read configuration for a single namespace
	return nil
}