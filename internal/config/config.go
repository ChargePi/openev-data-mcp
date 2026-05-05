package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database        DatabaseConfig  `mapstructure:"database"`
	Observability   ObservabilityConfig `mapstructure:"observability"`
	RefreshInterval time.Duration   `mapstructure:"refresh_interval"`
	HealthPort      int             `mapstructure:"health_port"`
	Port            int             `mapstructure:"port"`
}

type ObservabilityConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
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
	viper.SetDefault("observability.enabled", false)
	viper.SetDefault("observability.otlp_endpoint", "localhost:4317")
	viper.SetDefault("refresh_interval", "5m")
	viper.SetDefault("health_port", 9090)
	viper.SetDefault("port", 8080)

	viper.SetEnvPrefix("OPENEV_MCP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// AutomaticEnv does not resolve nested keys during Unmarshal; bind each explicitly.
	_ = viper.BindEnv("database.host", "OPENEV_MCP_DATABASE_HOST")
	_ = viper.BindEnv("database.port", "OPENEV_MCP_DATABASE_PORT")
	_ = viper.BindEnv("database.user", "OPENEV_MCP_DATABASE_USER")
	_ = viper.BindEnv("database.password", "OPENEV_MCP_DATABASE_PASSWORD")
	_ = viper.BindEnv("database.dbname", "OPENEV_MCP_DATABASE_DBNAME")
	_ = viper.BindEnv("database.ssl_mode", "OPENEV_MCP_DATABASE_SSL_MODE")
	_ = viper.BindEnv("observability.enabled", "OPENEV_MCP_OBSERVABILITY_ENABLED")
	_ = viper.BindEnv("observability.otlp_endpoint", "OPENEV_MCP_OBSERVABILITY_OTLP_ENDPOINT")
	_ = viper.BindEnv("refresh_interval", "OPENEV_MCP_REFRESH_INTERVAL")
	_ = viper.BindEnv("health_port", "OPENEV_MCP_HEALTH_PORT")
	_ = viper.BindEnv("port", "OPENEV_MCP_PORT")

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
