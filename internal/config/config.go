package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Platforms PlatformsConfig `mapstructure:"platforms"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name,
	)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type RabbitMQConfig struct {
	URL           string `mapstructure:"url"`
	PrefetchCount int    `mapstructure:"prefetch_count"`
}

type RateLimitConfig struct {
	Enabled bool                       `mapstructure:"enabled"`
	Tiers   map[string]RateLimitTier   `mapstructure:"tiers"`
}

type RateLimitTier struct {
	RequestsPerMin int `mapstructure:"requests_per_min"`
}

type PlatformsConfig struct {
	SMS      PlatformConfig `mapstructure:"sms"`
	WhatsApp PlatformConfig `mapstructure:"whatsapp"`
	Telegram PlatformConfig `mapstructure:"telegram"`
	Email    PlatformConfig `mapstructure:"email"`
}

type PlatformConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Provider  string `mapstructure:"provider"`
	RateLimit int    `mapstructure:"rate_limit"`
}

// Platform credential configs loaded from environment variables.
type TwilioConfig struct {
	AccountSID  string
	AuthToken   string
	PhoneNumber string
}

type SendGridConfig struct {
	APIKey    string
	FromEmail string
}

type WhatsAppConfig struct {
	APIKey  string
	PhoneID string
}

type TelegramConfig struct {
	BotToken string
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from file and environment variables.
func Load(path string) (*Config, error) {
	v := viper.New()

	// Config file
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific env vars to config keys
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.name", "DB_NAME")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("rabbitmq.url", "RABBITMQ_URL")
	v.BindEnv("logging.level", "LOG_LEVEL")

	// Defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("rabbitmq.prefetch_count", 10)
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Read config file (optional — env vars can work alone)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// LoadPlatformCredentials loads platform provider credentials from environment variables.
func LoadPlatformCredentials() (twilio TwilioConfig, sendgrid SendGridConfig, whatsapp WhatsAppConfig, telegram TelegramConfig) {
	v := viper.New()
	v.AutomaticEnv()

	twilio = TwilioConfig{
		AccountSID:  v.GetString("TWILIO_ACCOUNT_SID"),
		AuthToken:   v.GetString("TWILIO_AUTH_TOKEN"),
		PhoneNumber: v.GetString("TWILIO_PHONE_NUMBER"),
	}
	sendgrid = SendGridConfig{
		APIKey:    v.GetString("SENDGRID_API_KEY"),
		FromEmail: v.GetString("SENDGRID_FROM_EMAIL"),
	}
	whatsapp = WhatsAppConfig{
		APIKey:  v.GetString("WHATSAPP_API_KEY"),
		PhoneID: v.GetString("WHATSAPP_PHONE_ID"),
	}
	telegram = TelegramConfig{
		BotToken: v.GetString("TELEGRAM_BOT_TOKEN"),
	}

	return
}
