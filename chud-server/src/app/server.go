package app

import (
	"context"
	"embed"
	"errors"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/sebnow/chud/platform/auth"
	appconfig "github.com/sebnow/chud/platform/config"
	"github.com/sebnow/chud/platform/httpx"
	"github.com/sebnow/chud/platform/log"
)

type ServerConfig struct {
	Host              string
	Port              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:              "0.0.0.0",
		Port:              os.Getenv(appconfig.APIPort),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       2 * time.Minute,
	}
}

func createServer(
	ctx context.Context,
	config ServerConfig,
	staticFiles embed.FS,
	handlers Handlers,
) (*http.Server, error) {
	publicRouter := http.NewServeMux()
	protectedRouter := http.NewServeMux()

	SetupRouters(publicRouter, protectedRouter, staticFiles, handlers)

	logger := log.FromContext(ctx).With().Str("component", "api_server").Logger()
	ctx = log.WithContext(ctx, &logger)

	protectedHandler := auth.JwtAuth(handlers.Session, httpx.JSONFallback(protectedRouter))
	publicAPIPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/health",
	}

	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if !strings.HasPrefix(path, "/api/") || slices.Contains(publicAPIPaths, path) {
			publicRouter.ServeHTTP(w, r)
			return
		}

		protectedHandler.ServeHTTP(w, r)
	})

	handler := httpx.SecurityHeaders(router)
	handler = httpx.Recoverer(handler)
	handler = httpx.Logger(ctx, handler)

	server := &http.Server{
		Addr:              net.JoinHostPort(config.Host, config.Port),
		Handler:           handler,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
	}
	return server, nil
}

func StartServer(
	ctx context.Context,
	config ServerConfig,
	staticFiles embed.FS,
	handlers Handlers,
) (*http.Server, error) {
	logger := log.FromContext(ctx).With().Str("component", "api_server").Logger()

	server, err := createServer(ctx, config, staticFiles, handlers)
	if err != nil {
		logger.Error().Err(err).Msg("Error creating API server")
		return nil, err
	}

	go func() {
		logger.Info().Msgf("API server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("Error starting API server")
		}
	}()

	return server, nil
}
