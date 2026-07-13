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
	GitHub GitHubConfig `env-prefix:"GITHUB_"`
	Notion NotionConfig `env-prefix:"NOTION_"`
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
	Host       string `env:"HOST" env-default:"localhost"`
	Port       int    `env:"PORT" env-default:"5432"`
	User       string `env:"USER" env-default:"signal"`
	Password   string `env:"PASSWORD" env-default:"signal"`
	Name       string `env:"NAME" env-default:"signal"`
	SSLMode    string `env:"SSLMODE" env-default:"disable"`
	RailwayURL string // set from DATABASE_URL env var if present
}

func (d DBConfig) DSN() string {
	// Railway provides a full DATABASE_URL — use it directly if set
	if d.RailwayURL != "" {
		return d.RailwayURL
	}
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

// GitHubConfig holds GitHub integration configuration.
type GitHubConfig struct {
	Token string `env:"TOKEN" env-default:""`
	Org   string `env:"ORG" env-default:""`
}

// NotionConfig holds Notion integration configuration.
type NotionConfig struct {
	Token string `env:"TOKEN" env-default:""`
}

// Load reads configuration from environment variables and optional .env file.
func Load() (*Config, error) {
	var cfg Config

	// Read .env file — try current dir and parent dir (since binaries run from api/)
	// On Railway, .env won't exist — all vars come from environment
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
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Railway injects DATABASE_URL and REDIS_URL directly — handle both formats
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" && cfg.DB.Host == "localhost" {
		// Railway postgres URL — parse into DB config fields isn't needed
		// because pgx can accept a full DSN. We expose it via RailwayDSN().
		cfg.DB.RailwayURL = dbURL
	}
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" && cfg.Redis.Addr == "localhost:6379" {
		cfg.Redis.Addr = redisURL // go-redis accepts full redis:// URL as Addr too
	}
	// Railway injects PORT
	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.App.Port)
	}

	return &cfg, nil
}
