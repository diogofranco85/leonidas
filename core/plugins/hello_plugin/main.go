package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"leonidas/core/pkg/plugin"
)

// HelloPlugin é um plugin de exemplo que implementa HTTPPlugin
type HelloPlugin struct {
	info    plugin.PluginInfo
	running bool
}

// NewPlugin é a função que será chamada pelo gerenciador de plugins
func NewPlugin() plugin.Plugin {
	return &HelloPlugin{
		info: plugin.PluginInfo{
			Name:        "hello-plugin",
			Version:     "1.0.0",
			Description: "Plugin de exemplo que retorna mensagens de saudação",
			Author:      "Leonidas Team",
			BasePath:    "/hello", // Caminho base para as rotas do plugin
			CreatedAt:   time.Now(),
		},
		running: false,
	}
}

// Info retorna informações sobre o plugin
func (p *HelloPlugin) Info() plugin.PluginInfo {
	return p.info
}

// Initialize inicializa o plugin
func (p *HelloPlugin) Initialize(ctx context.Context, config plugin.PluginConfig) error {
	fmt.Printf("Inicializando plugin %s v%s\n", p.info.Name, p.info.Version)
	return nil
}

// Start inicia o plugin
func (p *HelloPlugin) Start(ctx context.Context) error {
	p.running = true
	fmt.Printf("Plugin %s iniciado\n", p.info.Name)
	return nil
}

// Stop para o plugin
func (p *HelloPlugin) Stop(ctx context.Context) error {
	p.running = false
	fmt.Printf("Plugin %s parado\n", p.info.Name)
	return nil
}

// IsRunning verifica se o plugin está rodando
func (p *HelloPlugin) IsRunning() bool {
	return p.running
}

// GetRoutes retorna as rotas HTTP do plugin
func (p *HelloPlugin) GetRoutes() []plugin.Route {
	return []plugin.Route{
		{
			Method: "GET",
			Path:   "/hello",
			Name:   "hello",
		},
		{
			Method: "POST",
			Path:   "/hello",
			Name:   "hello-post",
		},
	}
}

// HandleRequest processa uma requisição HTTP
func (p *HelloPlugin) HandleRequest(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	switch request.Path {
	case "/hello":
		if request.Method == "GET" {
			return p.handleHelloGet(ctx, request)
		} else if request.Method == "POST" {
			return p.handleHelloPost(ctx, request)
		}
	}

	return plugin.HTTPResponse{
		StatusCode: 404,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"error": "Rota não encontrada"}`),
	}, nil
}

// handleHelloGet processa requisições GET /hello
func (p *HelloPlugin) handleHelloGet(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"message": "Hello World from HelloPlugin!",
		"method":  "GET",
		"path":    "/hello",
		"plugin":  p.info.Name,
		"version": p.info.Version,
		"time":    time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(response)
	if err != nil {
		return plugin.HTTPResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error": "Erro interno do servidor"}`),
		}, err
	}

	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

// handleHelloPost processa requisições POST /hello
func (p *HelloPlugin) handleHelloPost(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	var requestData map[string]interface{}
	if len(request.Body) > 0 {
		if err := json.Unmarshal(request.Body, &requestData); err != nil {
			return plugin.HTTPResponse{
				StatusCode: 400,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       []byte(`{"error": "JSON inválido"}`),
			}, nil
		}
	}

	name := "World"
	if n, ok := requestData["name"].(string); ok {
		name = n
	}

	response := map[string]interface{}{
		"message": fmt.Sprintf("Hello %s from HelloPlugin!", name),
		"method":  "POST",
		"path":    "/hello",
		"plugin":  p.info.Name,
		"version": p.info.Version,
		"time":    time.Now().Format(time.RFC3339),
		"input":   requestData,
	}

	body, err := json.Marshal(response)
	if err != nil {
		return plugin.HTTPResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error": "Erro interno do servidor"}`),
		}, err
	}

	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

// main é necessário para compilar como plugin
func main() {
	// Este main não será executado quando carregado como plugin
	fmt.Println("HelloPlugin compilado como plugin")
}
