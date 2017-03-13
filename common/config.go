//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Config struct {
	conf *viper.Viper
}

//NewConfig return a new instance of Config by configPath
func NewConfig(configPath string) *Config {
	vp := viper.New()
	vp.SetConfigFile(configPath)
	err := vp.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Reading config file: %s error, %s !", configPath, err.Error()))
	}
	return &Config{
		conf: vp,
	}
}

func (cf *Config) Get(key string) interface{} {
	return cf.conf.Get(key)
}

func (cf *Config) GetString(key string) string {
	return cf.conf.GetString(key)
}

func (cf *Config) GetInt(key string) int {
	return cf.conf.GetInt(key)
}

func (cf *Config) GetInt64(key string) int64 {
	return cf.conf.GetInt64(key)
}

func (cf *Config) GetFloat64(key string) float64 {
	return cf.conf.GetFloat64(key)
}

func (cf *Config) GetBool(key string) bool {
	return cf.conf.GetBool(key)
}

func (cf *Config) GetDuration(key string) time.Duration {
	return cf.conf.GetDuration(key)
}

func (cf *Config) GetStringMap(key string) map[string]interface{} {
	return cf.conf.GetStringMap(key)
}

func (cf *Config) Set(key string, value interface{}) {
	cf.conf.Set(key, value)
}

// ContainsKey judge whether the key is set in the config
func (cf * Config) ContainsKey(key string) bool  {
	return cf.conf.IsSet(key)
}

// MergeConfig merge config by the config file path
func (cf *Config) MergeConfig(configPath string) (*Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		commonLogger.Errorf("open file: %s error, %v", configPath, err.Error())
		return cf, err
	}
	cf.conf.MergeConfig(f)
	return cf, nil
}
