package config

import (
	"fmt"
	viper "github.com/spf13/viper"
)

func GetString(config string) (s string) {
	viper.SetConfigName("dlcexchange") // name of config file (without extension)
	viper.AddConfigPath("/home/user/dlcexchange/alice")
	viper.AddConfigPath("$HOME/dlcexchange/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	return viper.GetString(config)
}


func GetInt(config string) (i int) {
	viper.SetConfigName("dlcexchange") // name of config file (without extension)
	viper.AddConfigPath("/home/user/dlcexchange/alice")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	return viper.GetInt(config)
}
