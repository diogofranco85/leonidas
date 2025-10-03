package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"leonidas/core/pkg/plugin"
)

// BackgroundPlugin é um plugin que executa tarefas em background
type BackgroundPlugin struct {
	info       plugin.PluginInfo
	running    bool
	stopChan   chan struct{}
	ticker     *time.Ticker
	workerDone chan struct{}
}

// NewPlugin é a função que será chamada pelo gerenciador de plugins
func NewPlugin() plugin.Plugin {
	return &BackgroundPlugin{
		info: plugin.PluginInfo{
			Name:        "background-plugin",
			Version:     "1.0.0",
			Description: "Plugin que executa tarefas em background",
			Author:      "Leonidas Team",
			CreatedAt:   time.Now(),
		},
		running:    false,
		stopChan:   make(chan struct{}),
		workerDone: make(chan struct{}),
	}
}

// Info retorna informações sobre o plugin
func (p *BackgroundPlugin) Info() plugin.PluginInfo {
	return p.info
}

// Initialize inicializa o plugin (configuração inicial)
func (p *BackgroundPlugin) Initialize(ctx context.Context, config plugin.PluginConfig) error {
	log.Printf("Inicializando plugin %s v%s", p.info.Name, p.info.Version)

	// Aqui você faria configurações iniciais como:
	// - Validar configurações
	// - Preparar estruturas de dados
	// - Conectar a serviços externos (sem iniciar processamento)

	log.Printf("Plugin %s inicializado com sucesso", p.info.Name)
	return nil
}

// Start inicia o plugin (ativa funcionalidades)
func (p *BackgroundPlugin) Start(ctx context.Context) error {
	if p.running {
		return fmt.Errorf("plugin %s já está rodando", p.info.Name)
	}

	log.Printf("Iniciando plugin %s", p.info.Name)

	// 1. Iniciar timer para tarefas periódicas
	p.ticker = time.NewTicker(30 * time.Second)

	// 2. Iniciar worker em background
	go p.backgroundWorker()

	// 3. Iniciar processamento de tarefas periódicas
	go p.periodicTasks()

	p.running = true
	log.Printf("Plugin %s iniciado com sucesso", p.info.Name)

	return nil
}

// Stop para o plugin
func (p *BackgroundPlugin) Stop(ctx context.Context) error {
	if !p.running {
		return fmt.Errorf("plugin %s não está rodando", p.info.Name)
	}

	log.Printf("Parando plugin %s", p.info.Name)

	// 1. Parar timer
	if p.ticker != nil {
		p.ticker.Stop()
	}

	// 2. Sinalizar para parar worker
	close(p.stopChan)

	// 3. Aguardar worker terminar (com timeout)
	select {
	case <-p.workerDone:
		log.Printf("Worker do plugin %s parou", p.info.Name)
	case <-time.After(5 * time.Second):
		log.Printf("Timeout ao parar worker do plugin %s", p.info.Name)
	}

	p.running = false
	log.Printf("Plugin %s parado com sucesso", p.info.Name)

	return nil
}

// IsRunning verifica se o plugin está rodando
func (p *BackgroundPlugin) IsRunning() bool {
	return p.running
}

// backgroundWorker executa tarefas em background
func (p *BackgroundPlugin) backgroundWorker() {
	defer close(p.workerDone)

	log.Printf("Worker do plugin %s iniciado", p.info.Name)

	for {
		select {
		case <-p.stopChan:
			log.Printf("Worker do plugin %s recebeu sinal de parada", p.info.Name)
			return
		default:
			// Executar tarefa de background
			p.processBackgroundTask()
			time.Sleep(10 * time.Second)
		}
	}
}

// periodicTasks executa tarefas periódicas
func (p *BackgroundPlugin) periodicTasks() {
	for {
		select {
		case <-p.stopChan:
			return
		case <-p.ticker.C:
			p.processPeriodicTask()
		}
	}
}

// processBackgroundTask simula uma tarefa de background
func (p *BackgroundPlugin) processBackgroundTask() {
	// Simular processamento de dados
	log.Printf("Plugin %s: Processando tarefa de background...", p.info.Name)

	// Aqui você faria:
	// - Processar filas de mensagens
	// - Limpar dados antigos
	// - Sincronizar com APIs externas
	// - Gerar relatórios
}

// processPeriodicTask simula uma tarefa periódica
func (p *BackgroundPlugin) processPeriodicTask() {
	// Simular tarefa periódica
	log.Printf("Plugin %s: Executando tarefa periódica...", p.info.Name)

	// Aqui você faria:
	// - Backup de dados
	// - Verificação de saúde
	// - Atualização de métricas
	// - Limpeza de logs antigos
}

// GetRoutes retorna as rotas HTTP do plugin
func (p *BackgroundPlugin) GetRoutes() []plugin.Route {
	return []plugin.Route{
		{
			Method: "GET",
			Path:   "/background/status",
			Name:   "background-status",
		},
		{
			Method: "POST",
			Path:   "/background/task",
			Name:   "background-task",
		},
	}
}

// HandleRequest processa uma requisição HTTP
func (p *BackgroundPlugin) HandleRequest(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	switch request.Path {
	case "/background/status":
		return p.handleStatus(ctx, request)
	case "/background/task":
		return p.handleTask(ctx, request)
	}

	return plugin.HTTPResponse{
		Code:    404,
		Headers: map[string]string{"Content-Type": "application/json"},
		Content: []byte(`{"error": "Rota não encontrada"}`),
	}, nil
}

// handleStatus retorna o status do plugin
func (p *BackgroundPlugin) handleStatus(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	status := map[string]interface{}{
		"plugin":    p.info.Name,
		"version":   p.info.Version,
		"running":   p.running,
		"uptime":    "calculado baseado no start time",
		"worker":    "ativo",
		"periodic":  "ativo",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(status)
	if err != nil {
		return plugin.HTTPResponse{
			Code:    500,
			Headers: map[string]string{"Content-Type": "application/json"},
			Content: []byte(`{"error": "Erro interno"}`),
		}, err
	}

	return plugin.HTTPResponse{
		Code:    200,
		Headers: map[string]string{"Content-Type": "application/json"},
		Content: body,
	}, nil
}

// handleTask executa uma tarefa específica
func (p *BackgroundPlugin) handleTask(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	if !p.running {
		return plugin.HTTPResponse{
			Code:    400,
			Headers: map[string]string{"Content-Type": "application/json"},
			Content: []byte(`{"error": "Plugin não está rodando"}`),
		}, nil
	}

	// Simular execução de tarefa
	log.Printf("Plugin %s: Executando tarefa via HTTP", p.info.Name)

	response := map[string]interface{}{
		"message":   "Tarefa executada com sucesso",
		"plugin":    p.info.Name,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(response)
	if err != nil {
		return plugin.HTTPResponse{
			Code:    500,
			Headers: map[string]string{"Content-Type": "application/json"},
			Content: []byte(`{"error": "Erro interno"}`),
		}, err
	}

	return plugin.HTTPResponse{
		Code:    200,
		Headers: map[string]string{"Content-Type": "application/json"},
		Content: body,
	}, nil
}

// main é necessário para compilar como plugin
func main() {
	fmt.Println("BackgroundPlugin compilado como plugin")
}
