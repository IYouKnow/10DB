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

	"github.com/pedro/10db-launch/apps/server/internal/platform/auth"
	"github.com/pedro/10db-launch/apps/server/internal/platform/config"
	"github.com/pedro/10db-launch/apps/server/internal/platform/crypto"
	"github.com/pedro/10db-launch/apps/server/internal/platform/db"
	"github.com/pedro/10db-launch/apps/server/internal/platform/postgres"
	"github.com/pedro/10db-launch/apps/server/internal/project"
	"github.com/pedro/10db-launch/apps/server/internal/user"
	"github.com/pedro/10db-launch/apps/server/internal/web"
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

	store := project.NewStore(sqliteDB)
	if err := store.EnsureSchema(ctx); err != nil {
		log.Fatalf("ensure project schema: %v", err)
	}
	userStore := user.NewStore(sqliteDB)
	if err := userStore.EnsureSchema(ctx); err != nil {
		log.Fatalf("ensure user schema: %v", err)
	}
	userService := user.New(userStore)
	cryptoService := crypto.New(cfg.MasterKey)
	projectService := project.New(store, pgService, cryptoService, postgres.AdminConfig{
		Host:     cfg.PGAdminHost,
		Port:     cfg.PGAdminPort,
		DBName:   cfg.PGAdminDB,
		User:     cfg.PGAdminUser,
		Password: cfg.PGAdminPassword,
		SSLMode:  cfg.PGSSLMode,
	})
	authService := auth.New(cfg.MasterKey, cfg.AppBaseURL, cfg.SessionTTL)

	handler := web.New(authService, userService, projectService, cfg.AllowedOrigins)
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
