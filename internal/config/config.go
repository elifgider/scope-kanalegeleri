package config

import (
	"os"
	"strings"
)

type Settings struct {
	Mode          string
	Server        ServerConfig
	Database      DatabaseConfig
	Admin         AdminConfig
	Paths         PathsConfig
	Telegram      TelegramConfig
	SessionSecret string
	ContactName   string
	ContactPhone  string
	ContactEmail  string
}

type ServerConfig struct {
	Address string
}

type DatabaseConfig struct {
	URL string
}

type AdminConfig struct {
	Username string
	Password string
}

type PathsConfig struct {
	TemplatesDir string
	UploadsDir   string
	StaticDir    string
}

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

// Load configurations from environment variables only
func Load() Settings {
	mode := envOrDefault("APP_ENV", "development")

	return Settings{
		Mode: mode,
		Server: ServerConfig{
			Address: envOrDefault("APP_ADDR", ":8080"),
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		Admin: AdminConfig{
			Username: envOrDefault("GO_ADMIN_USERNAME", "admin"),
			Password: os.Getenv("GO_ADMIN_PASSWORD"),
		},
		Telegram: TelegramConfig{
			BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
			ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		},
		Paths: PathsConfig{
			TemplatesDir: envOrDefault("TEMPLATES_DIR", "templates"),
			UploadsDir:   envOrDefault("UPLOADS_DIR", "uploads"),
			StaticDir:    envOrDefault("STATIC_DIR", "public/static"),
		},
		SessionSecret: envOrDefault("SESSION_SECRET", "super-secret-key-change-me"),
		ContactName:   envOrDefault("CONTACT_NAME", ""),
		ContactPhone:  envOrDefault("CONTACT_PHONE", ""),
		ContactEmail:  envOrDefault("CONTACT_EMAIL", ""),
	}
}

func envOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}
