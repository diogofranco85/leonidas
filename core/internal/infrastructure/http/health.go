package http

import (
	"leonidas/core/internal/infrastructure/cache"
	"leonidas/core/internal/infrastructure/config"
	"leonidas/core/internal/infrastructure/database"
	"leonidas/core/internal/infrastructure/plugin"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthResponse representa a resposta do health check
type HealthResponse struct {
	Status      string        `json:"status"`
	Message     string        `json:"message"`
	BasePath    string        `json:"basePath"`
	PluginsPath string        `json:"pluginsPath"`
	Port        string        `json:"port"`
	Timestamp   TimestampInfo `json:"timestamp"`
	Version     string        `json:"version"`
	Uptime      string        `json:"uptime"`
	System      *SystemInfo   `json:"system,omitempty"`
	Services    *ServicesInfo `json:"services,omitempty"`
}

// TimestampInfo contém informações de timestamp
type TimestampInfo struct {
	Unix int64  `json:"unix"`
	RFC  string `json:"rfc"`
}

// SystemInfo contém informações do sistema
type SystemInfo struct {
	GoVersion     string `json:"goVersion"`
	FiberVersion  string `json:"fiberVersion"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	NumCPU        int    `json:"numCpu"`
	NumGoroutines int    `json:"numGoroutines"`
}

// ServicesInfo contém status dos serviços
type ServicesInfo struct {
	Database string `json:"database"`
	Plugins  string `json:"plugins"`
}

// HealthHandler lida com as rotas de health check
type HealthHandler struct {
	config        *config.Config
	startTime     time.Time
	database      *database.Database
	redis         *cache.Redis
	pluginManager *plugin.Manager
}

// NewHealthHandler cria uma nova instância do HealthHandler
func NewHealthHandler(cfg *config.Config, database *database.Database, redis *cache.Redis, pluginManager *plugin.Manager) *HealthHandler {
	return &HealthHandler{
		config:        cfg,
		startTime:     time.Now(),
		database:      database,
		redis:         redis,
		pluginManager: pluginManager,
	}
}

// BasicHealth retorna um health check básico
func (h *HealthHandler) BasicHealth(c *fiber.Ctx) error {
	now := time.Now()
	uptime := now.Sub(h.startTime)

	response := HealthResponse{
		Status:   "healthy",
		Message:  h.config.ApiDescription,
		BasePath: h.config.BasePath,
		Port:     h.config.Port,
		Timestamp: TimestampInfo{
			Unix: now.Unix(),
			RFC:  now.Format("2006-01-02T15:04:05Z07:00"),
		},
		Version: "1.0.0",
		Uptime:  uptime.String(),
	}

	return c.JSON(response)
}

// DetailedHealth retorna um health check detalhado
func (h *HealthHandler) DetailedHealth(c *fiber.Ctx) error {
	now := time.Now()
	uptime := now.Sub(h.startTime)

	response := HealthResponse{
		Status:      "healthy",
		Message:     "Leonidas Core API is running",
		BasePath:    h.config.BasePath,
		PluginsPath: h.config.PluginsPath,
		Port:        h.config.Port,
		Timestamp: TimestampInfo{
			Unix: now.Unix(),
			RFC:  now.Format("2006-01-02T15:04:05Z07:00"),
		},
		Version: "1.0.0",
		Uptime:  uptime.String(),
		System: &SystemInfo{
			GoVersion:     runtime.Version(),
			FiberVersion:  "2.52.9",
			OS:            runtime.GOOS,
			Arch:          runtime.GOARCH,
			NumCPU:        runtime.NumCPU(),
			NumGoroutines: runtime.NumGoroutine(),
		},
		Services: &ServicesInfo{
			Database: h.getDatabaseStatus(),
			Plugins:  h.getPluginsStatus(),
		},
	}

	return c.JSON(response)
}

// ReadinessCheck verifica se a aplicação está pronta para receber tráfego
func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	// Aqui você pode adicionar verificações de dependências
	// como banco de dados, serviços externos, etc.

	response := HealthResponse{
		Status:   "ready",
		Message:  "Application is ready to receive traffic",
		BasePath: h.config.BasePath,
		Port:     h.config.Port,
		Timestamp: TimestampInfo{
			Unix: time.Now().Unix(),
			RFC:  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		},
		Version: "1.0.0",
	}

	return c.JSON(response)
}

// LivenessCheck verifica se a aplicação está viva
func (h *HealthHandler) LivenessCheck(c *fiber.Ctx) error {
	response := HealthResponse{
		Status:   "alive",
		Message:  "Application is alive",
		BasePath: h.config.BasePath,
		Port:     h.config.Port,
		Timestamp: TimestampInfo{
			Unix: time.Now().Unix(),
			RFC:  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		},
		Version: "1.0.0",
	}

	return c.JSON(response)
}
