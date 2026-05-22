package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Log        LogConfig        `mapstructure:"log"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	NATS       NATSConfig       `mapstructure:"nats"`
	Meilisearch MeilisearchConfig `mapstructure:"meilisearch"`
	LiteLLM    LiteLLMConfig    `mapstructure:"litellm"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Auth       AuthConfig       `mapstructure:"auth"`
	CORS       CORSConfig       `mapstructure:"cors"`
	OTEL       OTELConfig       `mapstructure:"otel"`
}

type ServerConfig struct {
	ListenAddr      string        `mapstructure:"listen_addr"`
	GracefulTimeout time.Duration `mapstructure:"graceful_timeout"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type DatabaseConfig struct {
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type SQLiteConfig struct {
	Path        string `mapstructure:"path"`
	WAL         bool   `mapstructure:"wal"`
	ForeignKeys bool   `mapstructure:"foreign_keys"`
}

type PostgresConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type NATSConfig struct {
	URL          string `mapstructure:"url"`
	ConsumerName string `mapstructure:"consumer_name"`
}

type MeilisearchConfig struct {
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"api_key"`
}

type LiteLLMConfig struct {
	URL     string        `mapstructure:"url"`
	APIKey  string        `mapstructure:"api_key"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

type AuthConfig struct {
	BcryptCost      int           `mapstructure:"bcrypt_cost"`
	RateLimit       int           `mapstructure:"rate_limit"`
	RateLimitWindow time.Duration `mapstructure:"rate_limit_window"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type OTELConfig struct {
	ServiceName string  `mapstructure:"service_name"`
	TraceRatio  float64 `mapstructure:"trace_ratio"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.nextm")
	v.AddConfigPath("/etc/nextm")

	v.SetEnvPrefix("NEXTM")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	setDefaults(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.listen_addr", ":8080")
	v.SetDefault("server.graceful_timeout", 30)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("database.sqlite.path", "./data/nextm.db")
	v.SetDefault("database.sqlite.wal", true)
	v.SetDefault("database.sqlite.foreign_keys", true)
	v.SetDefault("database.postgres.max_open_conns", 25)
	v.SetDefault("database.postgres.max_idle_conns", 10)
	v.SetDefault("database.postgres.conn_max_lifetime", "5m")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.consumer_name", "nextm-worker")
	v.SetDefault("meilisearch.host", "http://localhost:7700")
	v.SetDefault("litellm.url", "http://localhost:4000")
	v.SetDefault("litellm.timeout", "30s")
	v.SetDefault("jwt.access_token_ttl", "15m")
	v.SetDefault("jwt.refresh_token_ttl", "168h")
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.rate_limit", 5)
	v.SetDefault("auth.rate_limit_window", "15m")
	v.SetDefault("otel.service_name", "nextm-server")
	v.SetDefault("otel.trace_ratio", 0.1)
}
