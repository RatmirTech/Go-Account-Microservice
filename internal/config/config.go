package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string
	}
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	JWT struct {
		Secret          string
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration
	}
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения конфига: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("ошибка разбора конфига: %w", err)
	}

	accessTTL, err := time.ParseDuration(viper.GetString("jwt.access_token_ttl"))
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга access_token_ttl: %w", err)
	}
	cfg.JWT.AccessTokenTTL = accessTTL

	refreshTTL, err := time.ParseDuration(viper.GetString("jwt.refresh_token_ttl"))
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга refresh_token_ttl: %w", err)
	}
	cfg.JWT.RefreshTokenTTL = refreshTTL

	return &cfg, nil
}
