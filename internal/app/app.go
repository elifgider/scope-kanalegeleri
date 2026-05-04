package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"kanalegeleri/go-app/internal/config"
	"kanalegeleri/go-app/internal/handler"
	"kanalegeleri/go-app/internal/repository"
	"kanalegeleri/go-app/internal/service"
)

func Run() error {
	// .env dosyasını yükle (eğer varsa)
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ .env dosyası bulunamadı, sistem değişkenleri kullanılacak.")
	}

	cfg := config.Load()

	// GORM log seviyesi: production'da sadece hataları göster
	logLevel := gormlogger.Error
	if cfg.Mode == "development" {
		logLevel = gormlogger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("database open: %w", err)
	}

	// Connection pool ayarları
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	store := repository.NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	// Otomatik Admin kontrolü/oluşturma
	if err := store.EnsureAdminExists(context.Background(), cfg.Admin.Username, cfg.Admin.Password); err != nil {
		log.Printf("⚠️  Admin oluşturulamadı: %v", err)
	}

	uploadService := service.NewUploadService(cfg.Paths.UploadsDir)
	if err := uploadService.EnsureUploadsDir(); err != nil {
		return fmt.Errorf("create uploads dir: %w", err)
	}

	catalogService := service.NewCatalogService(store)
	orderService := service.NewOrderService(store, service.TelegramNotifier{
		BotToken: cfg.Telegram.BotToken,
		ChatID:   cfg.Telegram.ChatID,
	})
	authService := service.NewAuthService(store, cfg.SessionSecret, cfg.Mode == "production")

	engine, err := handler.Load(cfg, catalogService, orderService, authService, uploadService)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("app starting in %s mode on %s", cfg.Mode, cfg.Server.Address)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-quit:
		log.Printf("received signal %s, shutting down...", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		log.Println("server stopped gracefully")
	}

	return nil
}
