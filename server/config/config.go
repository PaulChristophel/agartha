package config

import (
	"log"

	"github.com/spf13/viper"
)

func Config() {

	viper.SetConfigName("config")         // name of config file (without extension)
	viper.SetConfigType("yaml")           // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/agartha/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.agartha") // call multiple times to add many search paths
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		log.Fatal("fatal error config file: %w", err)
	}

}
