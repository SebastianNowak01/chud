package main

import (
	"context"
	"embed"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sebnow/chud/app"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/config"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/log"
)

//go:embed all:static
var staticFiles embed.FS

func main() {
	logger := log.New()

	appCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	appCtx = log.WithContext(appCtx, &logger)

	err := config.Validate()
	if err != nil {
		logger.Fatal().Err(err).Msg("Environment variable validation failed")
	}
	logger.Info().Msg("Environment variable validation OK")

	database, err := db.NewFromEnv(appCtx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize PostgreSQL connection")
	}
	defer database.Close()
	logger.Info().Msg("PostgreSQL connection established")

	userService := users.NewUserService(users.UserServiceDeps{
		UserDAO: users.NewUserDAO(database.Querier()),
	})

	err = userService.EnsureAdminUserExists(appCtx, os.Getenv(config.AdminUser), os.Getenv(config.AdminPassword))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to ensure admin user exists")
	}

	handlers := app.Handlers{
		User: users.NewUserAPIController(userService),
	}

	// Start the server
	serverConfig := app.DefaultServerConfig()
	logger.Info().Msg("Starting API server...")
	server, done, err := app.StartServer(appCtx, serverConfig, staticFiles, handlers)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start API server")
		os.Exit(1)
	}
	logger.Info().Msg("API server started successfully")

	<-appCtx.Done()
	logger.Info().Msg("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	logger.Info().Msg("Shutting down API server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("API server shutdown error")
	} else {
		logger.Info().Msg("API server stopped gracefully")
	}
	close(done)

	logger.Info().Msg("All processes terminated successfully")
}
