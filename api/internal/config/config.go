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
	UserToken     string `env:"USER_TOKEN"` // For RTS search (optional)
	AppToken      string `env:"APP_TOKEN" env-required:"true"`
	SigningSecret string `env:"SIGNING_SECRET" env-required:"true"`
	ClientID      string `env:"CLIENT_ID" env-required:"true"`
	ClientSecret  string `env:"CLIENT_SECRET" env-required:"true"`
}

type OpenAIConfig struct {
	APIKey  string `env:"API_KEY" env-required:"true"`
	Model   string `env:"MODEL" env-default:"gpt-4o-mini"`
	BaseURL string `env:"BASE_URL" env-default:"https://api.openai.com/v1"`
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
	ServerURL           string `env:"SERVER_URL" env-default:"http://localhost:3001/sse"`
	Timeout             string `env:"SERVER_TIMEOUT" env-default:"30s"`
	CalendarCredsPath   string `env:"CALENDAR_CREDENTIALS_PATH" env-default:""`
	CalendarID          string `env:"CALENDAR_ID" env-default:"primary"`
}

// Load reads configuration from environment variables and optional .env file.
func Load() (*Config, error) {
	var cfg Config

	// Read .env file — try current dir and parent dir (since binaries run from api/)
	envPaths := []string{".env", "../.env"}
	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			if err := cleanenv.ReadConfig(p, &cfg); err != nil {
				slog.Warn("could not read .env file", "path", p, "error", err)
			}
			break
		}
	}

	// Environment variables override .env file values
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}
