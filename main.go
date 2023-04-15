package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RichterMaximilian/osttra-coding-assignment/api"
	"github.com/RichterMaximilian/osttra-coding-assignment/core"
	"github.com/RichterMaximilian/osttra-coding-assignment/migrate"
	"github.com/RichterMaximilian/osttra-coding-assignment/postgres"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type Config struct {
	DB struct {
		ConnStr       string `envconfig:"DB_CONN" required:"true"`
		MigrationsDir string `envconfig:"DB_MIGRATIONS_DIR" default:"file://migrations"`
	}
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		log.Fatalf("unable to parse config: %v", err)
	}

	err = migrate.Up(cfg.DB.MigrationsDir, cfg.DB.ConnStr)
	if err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, cfg.DB.ConnStr)
	if err != nil {
		log.Fatalf("connecting to DB: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("create logger: %v", err)
	}
	defer logger.Sync()

	repository := postgres.NewRepository(pool)

	service := core.NewService(repository)
	router := api.NewRouter(service, logger)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for SIGTERM
	<-signals
	// Shutdown server gracefully
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown server: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func parseConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("unable to process config: %w", err)
	}

	return cfg, nil
}
