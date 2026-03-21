package config

import (
	"cmp"
	"os"
	"time"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	RedisURL    string
	S3          S3Config
	JWT         JWTConfig
	Composio    ComposioConfig
	OAuth       OAuthConfig
}

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type ComposioConfig struct {
	APIKey            string
	GmailAccountID    string
	TelegramAccountID string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	RedirectBaseURL    string
}

func Load() Config {
	return Config{
		ServerPort:  cmp.Or(os.Getenv("SERVER_PORT"), "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    cmp.Or(os.Getenv("REDIS_URL"), "redis://localhost:6379"),
		S3: S3Config{
			Endpoint:  cmp.Or(os.Getenv("S3_ENDPOINT"), "localhost:9000"),
			AccessKey: cmp.Or(os.Getenv("S3_ACCESS_KEY"), "minioadmin"),
			SecretKey: cmp.Or(os.Getenv("S3_SECRET_KEY"), "minioadmin"),
			Bucket:    cmp.Or(os.Getenv("S3_BUCKET"), "estimatepro"),
			UseSSL:    os.Getenv("S3_USE_SSL") == "true",
		},
		JWT: JWTConfig{
			Secret:     os.Getenv("JWT_SECRET"),
			AccessTTL:  parseDuration(os.Getenv("JWT_ACCESS_TTL"), 15*time.Minute),
			RefreshTTL: parseDuration(os.Getenv("JWT_REFRESH_TTL"), 30*24*time.Hour),
		},
		Composio: ComposioConfig{
			APIKey:            os.Getenv("COMPOSIO_API_KEY"),
			GmailAccountID:    os.Getenv("COMPOSIO_GMAIL_ACCOUNT_ID"),
			TelegramAccountID: os.Getenv("COMPOSIO_TELEGRAM_ACCOUNT_ID"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectBaseURL:    cmp.Or(os.Getenv("OAUTH_REDIRECT_BASE_URL"), "http://localhost:3000"),
		},
	}
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
