// Package config exposes the service configuration, by consuming the configs from the config_file.yml file.
package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

// Config exposes app initialization configuration.
type Config struct {
	Port            string
	Partitions      int
	PartitionBuffer int
}

func defaultConfig() {
	viper.SetConfigName("config_file")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("fatal error config file: %v,", err))
	}
}

// GetConfig reads the config file and instantiates the Config data.
func GetConfig() Config {
	defaultConfig()
	return Config{Port: viper.GetString("port"), Partitions: viper.GetInt("partitions"), PartitionBuffer: viper.GetInt("partitionBuffer")}
}
