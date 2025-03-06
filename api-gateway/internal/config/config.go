package config

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	gatewayError "github.com/DarshDM/api-gateway/utils/error"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Server struct {
	Name      string   `mapstructure:"name"`
	Prefix    string   `mapstructure:"prefix"`
	Hosts     []string `mapstructure:"hosts"`
	Port      int      `mapstructure:"port"`
	ApiKey    string   `mapstructure:"api_key"`
	RateLimit int      `mapstructure:"rate_limit"`
}

type Config struct {
	Servers []Server `mapstructure:"servers"`
}

func Load(path string, l *log.Logger) (*Config, error) {
	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	// Attempt to read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found—try environment variables or defaults
			l.Warn("Config file not found, falling back to defaults or environment variables")
		} else {
			l.Errorf("Error reading config file: %v", err)
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
	}

	// Unmarshal into the Config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		l.Errorf("Error unmarshaling config file: %v", err)
		return nil, fmt.Errorf("cannot unmarshal config file: %v", err)
	}

	// Validate the configuration
	if len(cfg.Servers) == 0 {
		return nil, &gatewayError.GatewayError{Service: "config", Message: "No servers configured", Code: http.StatusInternalServerError}
	}

	for i, server := range cfg.Servers {
		if server.Name == "" {
			return nil, &gatewayError.GatewayError{Service: "config", Message: fmt.Sprintf("Server %d: name is required", i), Code: http.StatusInternalServerError}
		}
		if server.Prefix == "" {
			return nil, &gatewayError.GatewayError{Service: "config", Message: fmt.Sprintf("Server %d: prefix is required", i), Code: http.StatusInternalServerError}
		}
		if len(server.Hosts) == 0 {
			return nil, &gatewayError.GatewayError{Service: "config", Message: fmt.Sprintf("Server %d: at least one host is required", i), Code: http.StatusInternalServerError}
		}
		if server.Port <= 0 {
			return nil, &gatewayError.GatewayError{Service: "config", Message: fmt.Sprintf("Server %d: port must be positive", i), Code: http.StatusInternalServerError}
		}
		// if server.RateLimit <= 0 {
		// 	return nil, &gatewayError.GatewayError{Service: "config", Message: fmt.Sprintf("Server %d: rate_limit must be positive", i), Code: http.StatusInternalServerError}
		// }
		if !strings.HasPrefix(server.Prefix, "/") {
			server.Prefix = "/" + server.Prefix
			l.Warnf("Server %s: prefix adjusted to start with '/': %s", server.Name, server.Prefix)
		}

		for j, host := range server.Hosts {
			_, err := url.Parse(host)
			if err != nil || host == "" {
				return nil, &gatewayError.GatewayError{Service: "config", Message: fmt.Sprintf("Server %d: invalid host at index %d: %s", i, j, host), Code: http.StatusInternalServerError}
			}
		}
	}

	l.Info("Configuration loaded successfully")
	return &cfg, nil
}
