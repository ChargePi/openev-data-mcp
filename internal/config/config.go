package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database        DatabaseConfig      `mapstructure:"database"`
	Observability   ObservabilityConfig `mapstructure:"observability"`
	Qdrant          QdrantConfig        `mapstructure:"qdrant"`
	Embedding       EmbeddingConfig     `mapstructure:"embedding"`
	RefreshInterval time.Duration       `mapstructure:"refresh_interval"`
	HealthPort      int                 `mapstructure:"health_port"`
	Port            int                 `mapstructure:"port"`
}

type ObservabilityConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
}

type QdrantConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	APIKey      string `mapstructure:"api_key"`
	UseTLS      bool   `mapstructure:"use_tls"`
	Collection  string `mapstructure:"collection"`
	SearchLimit int    `mapstructure:"search_limit"`
	// MinScore discards search results with a cosine similarity below this value.
	MinScore float32 `mapstructure:"min_score"`
}

// EmbeddingConfig configures the Gemini embedding model used for semantic search.
type EmbeddingConfig struct {
	APIKey     string `mapstructure:"api_key"`
	Model      string `mapstructure:"model"`
	Dimensions int    `mapstructure:"dimensions"`
	BatchSize  int    `mapstructure:"batch_size"`
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
	viper.SetDefault("qdrant.host", "localhost")
	viper.SetDefault("qdrant.port", 6334)
	viper.SetDefault("qdrant.api_key", "")
	viper.SetDefault("qdrant.use_tls", false)
	viper.SetDefault("qdrant.collection", "ev_vehicles")
	viper.SetDefault("qdrant.search_limit", 10)
	viper.SetDefault("qdrant.min_score", 0.6)
	viper.SetDefault("embedding.api_key", "")
	viper.SetDefault("embedding.model", "gemini-embedding-001")
	viper.SetDefault("embedding.dimensions", 768)
	viper.SetDefault("embedding.batch_size", 100)
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
	_ = viper.BindEnv("qdrant.host", "OPENEV_MCP_QDRANT_HOST")
	_ = viper.BindEnv("qdrant.port", "OPENEV_MCP_QDRANT_PORT")
	_ = viper.BindEnv("qdrant.api_key", "OPENEV_MCP_QDRANT_API_KEY")
	_ = viper.BindEnv("qdrant.use_tls", "OPENEV_MCP_QDRANT_USE_TLS")
	_ = viper.BindEnv("qdrant.collection", "OPENEV_MCP_QDRANT_COLLECTION")
	_ = viper.BindEnv("qdrant.search_limit", "OPENEV_MCP_QDRANT_SEARCH_LIMIT")
	_ = viper.BindEnv("qdrant.min_score", "OPENEV_MCP_QDRANT_MIN_SCORE")
	_ = viper.BindEnv("embedding.api_key", "OPENEV_MCP_EMBEDDING_API_KEY")
	_ = viper.BindEnv("embedding.model", "OPENEV_MCP_EMBEDDING_MODEL")
	_ = viper.BindEnv("embedding.dimensions", "OPENEV_MCP_EMBEDDING_DIMENSIONS")
	_ = viper.BindEnv("embedding.batch_size", "OPENEV_MCP_EMBEDDING_BATCH_SIZE")
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
	if cfg.Embedding.APIKey == "" {
		return nil, fmt.Errorf("embedding.api_key (OPENEV_MCP_EMBEDDING_API_KEY) is required")
	}
	return &cfg, nil
}
