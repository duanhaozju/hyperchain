package test_util

import "hyperchain/common"

const (
	DB_CONFIG_PATH = "global.dbConfig"
)

// initial a config handler for testing.
func InitConfig(configPath string, dbConfigPath string) *common.Config {
	conf := common.NewConfig(configPath)
	conf.MergeConfig(dbConfigPath)
	return conf
}


