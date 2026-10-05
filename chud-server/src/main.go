package main

import (
	"context"
	"embed"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/sebnow/chud/app"
	"github.com/sebnow/chud/features/activities"
	"github.com/sebnow/chud/features/entries"
	"github.com/sebnow/chud/features/plans"
	"github.com/sebnow/chud/features/stats"
	"github.com/sebnow/chud/features/summaries"
	"github.com/sebnow/chud/features/users"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/config"
	"github.com/sebnow/chud/platform/db"
	"github.com/sebnow/chud/platform/llm"
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

	if err := clock.Init(os.Getenv(config.AppTimezone)); err != nil {
		logger.Fatal().Err(err).Msg("Invalid APP_TIMEZONE")
	}

	database, err := db.NewFromEnv(appCtx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize PostgreSQL connection")
	}
	defer database.Close()
	logger.Info().Msg("PostgreSQL connection established")

	userDAO := users.NewUserDAO(database.Querier())
	activityDAO := activities.NewActivityDAO(database.Querier())
	planDAO := plans.NewPlanDAO(database.Querier())
	statsDAO := stats.NewStatsDAO(database.Querier())

	userService := users.NewUserService(users.UserServiceDeps{
		UserDAO: userDAO,
	})
	activityService := activities.NewActivityService(activities.ActivityServiceDeps{
		ActivityDAO: activityDAO,
	})
	planService := plans.NewPlanService(plans.PlanServiceDeps{
		PlanDAO:     planDAO,
		ActivityDAO: activityDAO,
	})
	entryService := entries.NewEntryService(entries.EntryServiceDeps{
		DB:          database,
		ActivityDAO: activityDAO,
		PlanDAO:     planDAO,
	})
	statsService := stats.NewStatsService(stats.StatsServiceDeps{
		StatsDAO: statsDAO,
		UserDAO:  userDAO,
	})

	llmConfig, err := llm.ConfigFromEnv()
	if err != nil {
		logger.Warn().Err(err).Msg("LLM configuration")
	}
	llmClient := llm.New(llmConfig)
	if llmClient.Enabled() {
		logger.Info().Str("url", llmConfig.URL).Str("model", llmConfig.Model).Dur("timeout", llmConfig.Timeout).Int("max_tokens", llmConfig.MaxTokens).Msg("LLM summaries enabled")
	} else {
		logger.Info().Msg("LLM summaries disabled, LLM_URL is not set")
	}
	summaryService := summaries.NewSummaryService(summaries.SummaryServiceDeps{
		Ctx:          appCtx,
		SummaryDAO:   summaries.NewSummaryDAO(database.Querier()),
		StatsService: statsService,
		UserDAO:      userDAO,
		ActivityDAO:  activityDAO,
		LLM:          llmClient,
	})

	err = userService.EnsureAdminUserExists(appCtx, os.Getenv(config.AdminUser), os.Getenv(config.AdminPassword))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to ensure admin user exists")
	}

	handlers := app.Handlers{
		User:     users.NewUserAPIController(userService),
		Activity: activities.NewActivityAPIController(activityService),
		Plan:     plans.NewPlanAPIController(planService),
		Entry:    entries.NewEntryAPIController(entryService),
		Stats:    stats.NewStatsAPIController(statsService),
		Summary:  summaries.NewSummaryAPIController(summaryService),
	}

	serverConfig := app.DefaultServerConfig()
	logger.Info().Msg("Starting API server...")
	server, err := app.StartServer(appCtx, serverConfig, staticFiles, handlers)
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

	logger.Info().Msg("All processes terminated successfully")
}
