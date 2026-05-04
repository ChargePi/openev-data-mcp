package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database        DatabaseConfig `mapstructure:"database"`
	RefreshInterval time.Duration  `mapstructure:"refresh_interval"`
	HealthPort      int            `mapstructure:"health_port"`
	Port            int            `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

func Load(cfgFile string) (*Config, error) {
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "openevdata")
	viper.SetDefault("database.password", "openevdata")
	viper.SetDefault("database.dbname", "openevdata")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("refresh_interval", "5m")
	viper.SetDefault("health_port", 9090)
	viper.SetDefault("port", 8080)

	viper.SetEnvPrefix("OPENEV_MCP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}
