package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pedro/10db-launch/apps/server/internal/api"
	"github.com/pedro/10db-launch/apps/server/internal/auth"
	"github.com/pedro/10db-launch/apps/server/internal/config"
	"github.com/pedro/10db-launch/apps/server/internal/crypto"
	"github.com/pedro/10db-launch/apps/server/internal/db"
	"github.com/pedro/10db-launch/apps/server/internal/postgres"
	"github.com/pedro/10db-launch/apps/server/internal/projects"
	"github.com/pedro/10db-launch/apps/server/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	if err := config.LoadDotEnvIfPresent(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	sqliteDB, err := db.OpenSQLite(cfg.ControlDBPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer sqliteDB.Close()

	migrationsDir := filepath.Join("migrations")
	if err := db.RunMigrations(sqliteDB, migrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	ctx := context.Background()
	pgService, err := postgres.New(ctx, postgres.AdminConfig{
		Host:     cfg.PGAdminHost,
		Port:     cfg.PGAdminPort,
		DBName:   cfg.PGAdminDB,
		User:     cfg.PGAdminUser,
		Password: cfg.PGAdminPassword,
		SSLMode:  cfg.PGSSLMode,
	})
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pgService.Close()

	store := store.New(sqliteDB)
	cryptoService := crypto.New(cfg.MasterKey)
	projectService := projects.New(store, pgService, cryptoService, postgres.AdminConfig{
		Host:     cfg.PGAdminHost,
		Port:     cfg.PGAdminPort,
		DBName:   cfg.PGAdminDB,
		User:     cfg.PGAdminUser,
		Password: cfg.PGAdminPassword,
		SSLMode:  cfg.PGSSLMode,
	})
	authService := auth.New(cfg.AdminUsername, cfg.AdminPassword, cfg.MasterKey, cfg.AppBaseURL, cfg.SessionTTL)

	handler := api.New(authService, projectService, cfg.AllowedOrigins)
	server := &http.Server{
		Addr:    cfg.AppAddr,
		Handler: handler.Router(filepath.Join(".", "static")),
	}

	go func() {
		log.Printf("10DB Launch listening on %s", cfg.AppAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func runHealthcheck() int {
	resp, err := http.Get("http://127.0.0.1:8080/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
