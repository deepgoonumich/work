package api

import (
	"gitlab.benzinga.io/benzinga/ftp-engine/sender"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.benzinga.io/benzinga/ftp-engine/config"
)

type H struct {
	logger *zap.Logger
	config *config.Config
	sender sender.Sender
}

func LoadRoutes(cfg *config.Config, logger *zap.Logger, s sender.Sender) *gin.Engine {

	h := H{
		logger: logger,
		config: cfg,
		sender: s,
	}

	// Use Gin Release Mode in Production Environment
	if cfg.AppEnv == config.ProductionEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	// Init Router
	g := gin.New()

	// Middlewares
	g.Use(gin.Recovery())

	// Prometheus Endpoint
	g.GET("/metrics", func(c *gin.Context) {
		prom := promhttp.Handler()
		prom.ServeHTTP(c.Writer, c.Request)
	}).Use()

	// API Routes
	g.GET("/healthz", h.getStatus)

	return g
}
