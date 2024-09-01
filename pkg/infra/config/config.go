package config

import (
	"log"
	"strings"

	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/binance"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/db"
	"github.com/spf13/viper"
)

func NewBinanceConfig(
	cfg *viper.Viper,
) *binance.BinanceConfig {
	c := &binance.BinanceConfig{
		BaseEndpoint: cfg.GetString("BinanceConfig.BaseEndpoint"),
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
		Host:     cfg.GetString("DBConfig.Host"),
		Port:     cfg.GetString("DBConfig.Port"),
		User:     cfg.GetString("DBConfig.User"),
		Password: cfg.GetString("DBConfig.Password"),
		DBName:   cfg.GetString("DBConfig.DBName"),
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

func InitializeConfig(
	envName string,
	envPath string,
	envPrefix string,
) error {
	viper.SetConfigName(envName)
	viper.KeyDelimiter("__")
	viper.AddConfigPath(envPath)
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()
	viper.BindEnv("PORT", "PORT")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Print("No config file found")
		} else {
			log.Print("Error: Failed loading config")
			return err
		}
	}
	return nil
}

func NewConfig() *viper.Viper {
	if err := InitializeConfig(
		"env",
		"./env",
		internal.SERVICE_NAME,
	); err != nil {
		panic(err)
	}

	return viper.GetViper()
}
