package config

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type Server struct {
	Name   string `mapstructure:"name"`
	Host   string `mapstructure:"host"`
	Prefix string `mapstructure:"prefix"`
	Port   int    `mapstructure:"port"`
	ApiKey string `mapstructure:"api_key"`
}

type Config struct {
	Servers []Server `mapstructure:"servers"`
}

func Load(path string, l *log.Logger) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	if err := viper.ReadInConfig(); err != nil {
		l.Println("Error while reading config file")
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		l.Println("Error while unmarshaling config file")
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}
	return &cfg, nil
}
