package http

import (
	"context"
	"fmt"
	"time"
)

// getDatabaseStatus verifica o status da conexão com o banco de dados
func (h *HealthHandler) getDatabaseStatus() string {
	if h.database == nil {
		return "disconnected"
	}

	if err := h.database.Ping(); err != nil {
		return "error"
	}

	return "connected"
}

// getRedisStatus verifica o status da conexão com o Redis
func (h *HealthHandler) getRedisStatus() string {
	if h.redis == nil {
		return "disconnected"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.redis.Ping(ctx); err != nil {
		return "error"
	}

	return "connected"
}

// getPluginsStatus verifica o status dos plugins
func (h *HealthHandler) getPluginsStatus() string {
	if h.pluginManager == nil {
		return "disabled"
	}

	count := h.pluginManager.GetPluginCount()
	if count == 0 {
		return "no_plugins"
	}

	return fmt.Sprintf("loaded_%d", count)
}
