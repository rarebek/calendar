// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"github.com/go-redis/redis/v8"
	"job_tasks/calendar/internal/usecase"
	"job_tasks/calendar/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	_ "job_tasks/calendar/docs"
)

// NewRouter -.
// Swagger spec:
// @title       Calendar by Nodirbek No'monov
// @description You can test methods
// @in header
// @name Authorization
// @version     1.0
// @host        localhost:8080
func NewRouter(handler *gin.Engine, l logger.Interface, u usecase.User, e usecase.Event, redis *redis.Client) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	handler.GET("/swagger/*any", swaggerHandler)

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newUserRoutes(h, u, l, redis)
		newEventRoutes(h, e, l)
	}
}
