package config

import (
	"fmt"
	viper "github.com/spf13/viper"
)

func GetString(config string) (s string) {
	viper.SetConfigName("lit-trader") // name of config file (without extension)
	viper.AddConfigPath("$GOPATH/src/github.com/navybluesilver/lit-trader/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	return viper.GetString(config)
}


func GetInt(config string) (i int) {
	viper.SetConfigName("lit-trader") // name of config file (without extension)
	viper.AddConfigPath("$GOPATH/src/github.com/navybluesilver/lit-trader/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	return viper.GetInt(config)
}
