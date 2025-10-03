package plugin

import (
	"context"
	"time"
)

// PluginInfo contém informações básicas sobre um plugin
type PluginInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	BasePath    string    `json:"base_path"` // Caminho base para as rotas do plugin
	CreatedAt   time.Time `json:"created_at"`
}

// PluginConfig representa a configuração de um plugin
type PluginConfig struct {
	Enabled  bool                   `json:"enabled"`
	Settings map[string]interface{} `json:"settings"`
}

// Plugin é a interface base que todos os plugins devem implementar
type Plugin interface {
	// Info retorna informações sobre o plugin
	Info() PluginInfo

	// Initialize inicializa o plugin com configurações
	Initialize(ctx context.Context, config PluginConfig) error

	// Start inicia o plugin
	Start(ctx context.Context) error

	// Stop para o plugin
	Stop(ctx context.Context) error

	// IsRunning verifica se o plugin está rodando
	IsRunning() bool
}

// HTTPPlugin é um plugin que expõe rotas HTTP
type HTTPPlugin interface {
	Plugin

	// GetRoutes retorna as rotas HTTP do plugin
	GetRoutes() []Route

	// HandleRequest processa uma requisição HTTP
	HandleRequest(ctx context.Context, request HTTPRequest) (HTTPResponse, error)
}

// BackgroundPlugin é um plugin que executa tarefas em background
type BackgroundPlugin interface {
	Plugin

	// GetTasks retorna as tarefas do plugin
	GetTasks() []Task

	// ExecuteTask executa uma tarefa específica
	ExecuteTask(ctx context.Context, taskID string, params map[string]interface{}) error
}

// EventPlugin é um plugin que responde a eventos
type EventPlugin interface {
	Plugin

	// GetEvents retorna os eventos que o plugin escuta
	GetEvents() []string

	// HandleEvent processa um evento
	HandleEvent(ctx context.Context, event Event) error
}

// Route representa uma rota HTTP
type Route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
}

// HTTPRequest representa uma requisição HTTP
type HTTPRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
	Query   map[string]string `json:"query"`
}

// HTTPResponse representa uma resposta HTTP
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// Task representa uma tarefa de background
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schedule    string                 `json:"schedule"` // Cron expression
	Params      map[string]interface{} `json:"params"`
}

// Event representa um evento do sistema
type Event struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}
