package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/binance"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/snowflake"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/db"
	"github.com/spf13/viper"
)

func NewInternalEnvConfig(
	cfg *viper.Viper,
) *internal.EnvConfig {
	c := internal.EnvConfig{
		IsDevMode: cfg.GetBool("ENV_ISDEVMODE"),
	}
	return &c
}

func NewSnowflakeConfig(
	cfg *viper.Viper,
) *snowflake.SnowflakeConfig {
	c := snowflake.SnowflakeConfig{
		NodeNumber: cfg.GetInt64("SNOWFLAKE_NODENUMBER"),
	}
	return &c
}

func NewBinanceConfig(
	cfg *viper.Viper,
) *binance.BinanceConfig {
	c := &binance.BinanceConfig{
		BaseEndpoint: cfg.GetString("BINANCE_BASEENDPOINT"),
	}
	if c.BaseEndpoint == "" {
		panic("binance base endpoint not provided")
	}
	return c
}

func NewDBConfig(
	cfg *viper.Viper,
) *db.DBConfigs {
	c := &db.DBConfigs{
		Host:     cfg.GetString("DB_HOST"),
		Port:     cfg.GetString("DB_PORT"),
		User:     cfg.GetString("DB_USER"),
		Password: cfg.GetString("DB_PASSWORD"),
		DBName:   cfg.GetString("DB_DBNAME"),
	}
	if c.Host == "" {
		panic("db host not provided")
	}
	if c.Port == "" {
		panic("db port not provided")
	}
	if c.User == "" {
		panic("db user not provided")
	}
	if c.Password == "" {
		panic("db password not provided")
	}
	if c.DBName == "" {
		panic("db name not provided")
	}
	return c
}

func InitializeConfig() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("Error loading .env file - %w", err)
	}

	viper.AutomaticEnv()
	return nil
}

func NewConfig() *viper.Viper {
	if err := InitializeConfig(); err != nil {
		panic(err)
	}

	return viper.GetViper()
}
