package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds all configuration for the Signal application.
type Config struct {
	App    AppConfig    `env-prefix:"APP_"`
	Slack  SlackConfig  `env-prefix:"SLACK_"`
	OpenAI OpenAIConfig `env-prefix:"OPENAI_"`
	DB     DBConfig     `env-prefix:"DB_"`
	Redis  RedisConfig  `env-prefix:"REDIS_"`
	MCP    MCPConfig    `env-prefix:"MCP_"`
}

type AppConfig struct {
	Env         string `env:"ENV" env-default:"development"`
	Port        int    `env:"PORT" env-default:"8080"`
	LogLevel    string `env:"LOG_LEVEL" env-default:"info"`
	FrontendURL string `env:"FRONTEND_URL" env-default:"http://localhost:3000"`
}

type SlackConfig struct {
	BotToken      string `env:"BOT_TOKEN" env-required:"true"`
	AppToken      string `env:"APP_TOKEN" env-required:"true"`
	SigningSecret string `env:"SIGNING_SECRET" env-required:"true"`
	ClientID      string `env:"CLIENT_ID" env-required:"true"`
	ClientSecret  string `env:"CLIENT_SECRET" env-required:"true"`
}

type OpenAIConfig struct {
	APIKey string `env:"API_KEY" env-required:"true"`
	Model  string `env:"MODEL" env-default:"gpt-4o-mini"`
}

type DBConfig struct {
	Host     string `env:"HOST" env-default:"localhost"`
	Port     int    `env:"PORT" env-default:"5432"`
	User     string `env:"USER" env-default:"signal"`
	Password string `env:"PASSWORD" env-default:"signal"`
	Name     string `env:"NAME" env-default:"signal"`
	SSLMode  string `env:"SSLMODE" env-default:"disable"`
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string `env:"ADDR" env-default:"localhost:6379"`
	Password string `env:"PASSWORD" env-default:""`
}

type MCPConfig struct {
	ServerURL string `env:"SERVER_URL" env-default:"http://localhost:3001/sse"`
	Timeout   string `env:"SERVER_TIMEOUT" env-default:"30s"`
}

// Load reads configuration from environment variables and optional .env file.
func Load() (*Config, error) {
	var cfg Config

	// Try loading .env file (ignore error if not present)
	_ = cleanenv.ReadEnv(&cfg)

	// Override with environment variables
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Also check .env file for development
	if cfg.App.Env == "development" {
		if _, err := os.Stat(".env"); err == nil {
			if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
				slog.Warn("could not read .env file", "error", err)
			}
		}
	}

	return &cfg, nil
}
