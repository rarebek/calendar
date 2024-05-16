// Package app configures and runs application.
package app

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/go-redis/redis/v8"
	"job_tasks/calendar/config"
	v1 "job_tasks/calendar/internal/controller/http/v1"
	"job_tasks/calendar/internal/usecase"
	"job_tasks/calendar/internal/usecase/repo"
	"job_tasks/calendar/pkg/httpserver"
	"job_tasks/calendar/pkg/logger"
	"job_tasks/calendar/pkg/postgres"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
	})

	// Use case
	userUseCase := usecase.NewUserUseCase(
		repo.NewUserRepo(pg),
	)

	eventUseCase := usecase.NewEventUseCase(
		repo.NewEventRepo(pg))

	// HTTP Server
	handler := gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.AllowBrowserExtensions = true
	corsConfig.AllowMethods = []string{"*"}
	handler.Use(cors.New(corsConfig))

	v1.NewRouter(handler, l, userUseCase, eventUseCase, redisClient)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
