package config

import (
	"bytes"
	"fmt"
	"github.com/terasum/viper"
	"io/ioutil"
)

func ReadToml(filepath string) {
	viper.SetConfigType("TOML")
	configByte, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	viper.ReadConfig(bytes.NewBuffer(configByte))
	eca := viper.Get("ecert")
	fmt.Println(eca)
}

func WriteConfig(filepath string) {
	viper := viper.New()
	viper.SetConfigFile(filepath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println(viper.Get("ecert"))
	viper.Set("name", "test")
	viper.WriteConfigAs("configagain.toml")
	viper.WriteConfigAs("configagain.yaml")
	viper.WriteConfigAs("configagain.json")

}
