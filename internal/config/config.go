package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fileConfig struct {
	DefaultMode string                 `json:"default_mode"`
	Modes       map[string]Environment `json:"modes"`
}

type Environment struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Admin    AdminConfig    `json:"admin"`
	Paths    PathsConfig    `json:"paths"`
	Telegram TelegramConfig `json:"telegram"`
}

type ServerConfig struct {
	Address string `json:"address"`
}

type DatabaseConfig struct {
	URL string `json:"url"`
}

type AdminConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PathsConfig struct {
	TemplatesDir string `json:"templates_dir"`
	UploadsDir   string `json:"uploads_dir"`
	StaticDir    string `json:"static_dir"`
}

type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

type Settings struct {
	Mode string
	Environment
}

func Load(configPath string) (Settings, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Settings{}, fmt.Errorf("config read: %w", err)
	}

	var parsed fileConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		return Settings{}, fmt.Errorf("config parse: %w", err)
	}

	mode := strings.TrimSpace(os.Getenv("APP_ENV"))
	if mode == "" {
		mode = strings.TrimSpace(parsed.DefaultMode)
	}
	if mode == "" {
		mode = "development"
	}

	envConfig, ok := parsed.Modes[mode]
	if !ok {
		return Settings{}, fmt.Errorf("config mode not found: %s", mode)
	}

	applyEnvOverrides(&envConfig)

	return Settings{
		Mode:        mode,
		Environment: envConfig,
	}, nil
}

func applyEnvOverrides(cfg *Environment) {
	cfg.Server.Address = envOrDefault("APP_ADDR", cfg.Server.Address)
	cfg.Database.URL = envOrDefault("DATABASE_URL", cfg.Database.URL)
	cfg.Admin.Username = envOrDefault("GO_ADMIN_USERNAME", cfg.Admin.Username)
	cfg.Admin.Password = envOrDefault("GO_ADMIN_PASSWORD", cfg.Admin.Password)
	cfg.Telegram.BotToken = envOrDefault("TELEGRAM_BOT_TOKEN", cfg.Telegram.BotToken)
	cfg.Telegram.ChatID = envOrDefault("TELEGRAM_CHAT_ID", cfg.Telegram.ChatID)
	cfg.Paths.TemplatesDir = envOrDefault("TEMPLATES_DIR", cfg.Paths.TemplatesDir)
	cfg.Paths.UploadsDir = envOrDefault("UPLOADS_DIR", cfg.Paths.UploadsDir)
	cfg.Paths.StaticDir = envOrDefault("STATIC_DIR", cfg.Paths.StaticDir)
}

func envOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func ResolveExistingDir(candidates ...string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		return filepath.Clean(candidate)
	}
	return "."
}
