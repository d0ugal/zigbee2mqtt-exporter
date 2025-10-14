package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/version"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	cfg     *config.Config
	metrics *metrics.Registry
	server  *http.Server
	router  *gin.Engine
}

// New creates a new server instance
func New(cfg *config.Config, metrics *metrics.Registry) *Server {
	// Set Gin to release mode unless debug logging is enabled
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	s := &Server{
		cfg:     cfg,
		metrics: metrics,
		router:  router,
	}

	s.setupRoutes()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) setupRoutes() {
	// Root endpoint with HTML dashboard
	s.router.GET("/", s.handleRoot)

	// Metrics endpoint
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health endpoint
	s.router.GET("/health", s.handleHealth)
}

// handleRoot serves the web UI interface
func (s *Server) handleRoot(c *gin.Context) {
	versionInfo := version.Get()
	metricsInfo := s.metrics.GetMetricsInfo()

	// Convert metrics to template data
	metrics := make([]MetricData, 0, len(metricsInfo))
	for _, metric := range metricsInfo {
		metrics = append(metrics, MetricData{
			Name:         metric.Name,
			Help:         metric.Help,
			Labels:       metric.Labels,
			ExampleValue: metric.ExampleValue,
		})
	}

	data := TemplateData{
		Version:   versionInfo.Version,
		Commit:    versionInfo.Commit,
		BuildDate: versionInfo.BuildDate,
		Status:    "Connected",
		Metrics:   metrics,
		Config: ConfigData{
			WebSocketURL: s.cfg.WebSocket.URL,
			DeviceCount:  0, // Hardcoded for now - would need Zigbee2MQTT client reference to get actual count
		},
	}

	c.Header("Content-Type", "text/html")

	if err := mainTemplate.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
	}
}

func (s *Server) handleHealth(c *gin.Context) {
	versionInfo := version.Get()
	c.JSON(http.StatusOK, gin.H{
		"status":     "healthy",
		"timestamp":  time.Now().Unix(),
		"service":    "zigbee2mqtt-exporter",
		"version":    versionInfo.Version,
		"commit":     versionInfo.Commit,
		"build_date": versionInfo.BuildDate,
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	slog.Info("Starting server", "address", addr)

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	slog.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
