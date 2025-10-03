package plugin

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	goplugin "plugin"
	"strings"
	"sync"

	"leonidas/core/internal/infrastructure/config"
	"leonidas/core/pkg/plugin"

	"gorm.io/gorm"
)

// Manager gerencia o carregamento e execução de plugins
type Manager struct {
	config      *config.Config
	plugins     map[string]plugin.Plugin
	httpPlugins map[string]plugin.HTTPPlugin
	mu          sync.RWMutex
	logger      *log.Logger
	database    *gorm.DB
}

// NewManager cria uma nova instância do gerenciador de plugins
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:      cfg,
		plugins:     make(map[string]plugin.Plugin),
		httpPlugins: make(map[string]plugin.HTTPPlugin),
		logger:      log.New(os.Stdout, "[PluginManager] ", log.LstdFlags),
		database:    nil,
	}
}

// SetDatabase define a conexão do banco de dados para os plugins
func (m *Manager) SetDatabase(db *gorm.DB) {
	m.database = db
}

// LoadPlugins carrega todos os plugins da pasta configurada
func (m *Manager) LoadPlugins(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pluginsPath := m.config.PluginsPath
	m.logger.Printf("Verificando pasta de plugins: %s", pluginsPath)

	// Verificar se a pasta existe
	if _, err := os.Stat(pluginsPath); os.IsNotExist(err) {
		m.logger.Printf("Pasta de plugins não encontrada: %s", pluginsPath)
		return fmt.Errorf("pasta de plugins não encontrada: %s", pluginsPath)
	}

	// Buscar arquivos .so (shared objects) na pasta
	soFiles, err := m.findPluginFiles(pluginsPath)
	if err != nil {
		return fmt.Errorf("erro ao buscar arquivos de plugin: %w", err)
	}

	if len(soFiles) == 0 {
		m.logger.Printf("Nenhum plugin encontrado na pasta: %s", pluginsPath)
		return nil
	}

	m.logger.Printf("Encontrados %d plugins para carregar", len(soFiles))

	// Carregar cada plugin
	for _, soFile := range soFiles {
		if err := m.loadPlugin(ctx, soFile); err != nil {
			m.logger.Printf("Erro ao carregar plugin %s: %v", soFile, err)
			continue
		}
	}

	m.logger.Printf("Plugins carregados com sucesso: %d", len(m.plugins))
	return nil
}

// findPluginFiles busca arquivos .so na pasta de plugins
func (m *Manager) findPluginFiles(pluginsPath string) ([]string, error) {
	var soFiles []string

	err := filepath.Walk(pluginsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Verificar se é um arquivo .so
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".so") {
			soFiles = append(soFiles, path)
		}

		return nil
	})

	return soFiles, err
}

// loadPlugin carrega um plugin específico
func (m *Manager) loadPlugin(ctx context.Context, pluginPath string) error {
	m.logger.Printf("Carregando plugin: %s", pluginPath)

	// Carregar o plugin
	p, err := goplugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("falha ao abrir plugin %s: %w", pluginPath, err)
	}

	// Buscar o símbolo "NewPlugin" ou "Plugin"
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		// Tentar buscar diretamente por "Plugin"
		symbol, err = p.Lookup("Plugin")
		if err != nil {
			return fmt.Errorf("símbolo 'NewPlugin' ou 'Plugin' não encontrado em %s", pluginPath)
		}
	}

	// Verificar se o símbolo é uma função
	newPluginFunc, ok := symbol.(func() plugin.Plugin)
	if !ok {
		return fmt.Errorf("símbolo não é uma função que retorna plugin.Plugin em %s", pluginPath)
	}

	// Criar instância do plugin
	pluginInstance := newPluginFunc()
	if pluginInstance == nil {
		return fmt.Errorf("plugin retornou nil em %s", pluginPath)
	}

	// Obter informações do plugin
	info := pluginInstance.Info()
	pluginKey := fmt.Sprintf("%s-%s", info.Name, info.Version)

	m.logger.Printf("Plugin carregado: %s v%s", info.Name, info.Version)

	// Inicializar o plugin
	pluginConfig := plugin.PluginConfig{
		Enabled:  true,
		Settings: make(map[string]interface{}),
	}

	if err := pluginInstance.Initialize(ctx, pluginConfig); err != nil {
		return fmt.Errorf("falha ao inicializar plugin %s: %w", pluginKey, err)
	}

	// Armazenar o plugin
	m.plugins[pluginKey] = pluginInstance

	// Configurar conexão do banco para o auth-plugin
	if info.Name == "auth-plugin" && m.database != nil {
		// Tentar chamar a função SetDatabaseConnection se existir
		if setDBFunc, err := p.Lookup("SetDatabaseConnection"); err == nil {
			if setDB, ok := setDBFunc.(func(*gorm.DB)); ok {
				setDB(m.database)
				m.logger.Printf("Conexão do banco configurada para o plugin %s", info.Name)
			}
		}
	}

	// Iniciar automaticamente o auth-plugin
	if info.Name == "auth-plugin" {
		m.logger.Printf("Iniciando automaticamente o plugin %s", info.Name)
		if err := pluginInstance.Start(ctx); err != nil {
			m.logger.Printf("Aviso: Falha ao iniciar automaticamente o plugin %s: %v", info.Name, err)
		} else {
			m.logger.Printf("Plugin %s iniciado automaticamente com sucesso", info.Name)
		}
	}

	// Verificar se é um plugin HTTP
	if httpPlugin, ok := pluginInstance.(plugin.HTTPPlugin); ok {
		m.httpPlugins[pluginKey] = httpPlugin
		m.logger.Printf("Plugin HTTP registrado: %s", pluginKey)
	}

	return nil
}

// GetPlugin retorna um plugin específico
func (m *Manager) GetPlugin(name, version string) (plugin.Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pluginKey := fmt.Sprintf("%s-%s", name, version)
	plugin, exists := m.plugins[pluginKey]
	if !exists {
		return nil, fmt.Errorf("plugin %s v%s não encontrado", name, version)
	}

	return plugin, nil
}

// GetAllPlugins retorna todos os plugins carregados
func (m *Manager) GetAllPlugins() []plugin.Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]plugin.Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// GetHTTPPlugins retorna todos os plugins HTTP
func (m *Manager) GetHTTPPlugins() []plugin.HTTPPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]plugin.HTTPPlugin, 0, len(m.httpPlugins))
	for _, p := range m.httpPlugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// StartPlugin inicia um plugin específico
func (m *Manager) StartPlugin(ctx context.Context, name, version string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[fmt.Sprintf("%s-%s", name, version)]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s v%s não encontrado", name, version)
	}

	return plugin.Start(ctx)
}

// StopPlugin para um plugin específico
func (m *Manager) StopPlugin(ctx context.Context, name, version string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[fmt.Sprintf("%s-%s", name, version)]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s v%s não encontrado", name, version)
	}

	return plugin.Stop(ctx)
}

// StartAllPlugins inicia todos os plugins
func (m *Manager) StartAllPlugins(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for key, plugin := range m.plugins {
		if err := plugin.Start(ctx); err != nil {
			m.logger.Printf("Erro ao iniciar plugin %s: %v", key, err)
			continue
		}
		m.logger.Printf("Plugin iniciado: %s", key)
	}

	return nil
}

// StopAllPlugins para todos os plugins
func (m *Manager) StopAllPlugins(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for key, plugin := range m.plugins {
		if err := plugin.Stop(ctx); err != nil {
			m.logger.Printf("Erro ao parar plugin %s: %v", key, err)
			continue
		}
		m.logger.Printf("Plugin parado: %s", key)
	}

	return nil
}

// GetPluginCount retorna o número de plugins carregados
func (m *Manager) GetPluginCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.plugins)
}

// GetPluginInfo retorna informações de todos os plugins
func (m *Manager) GetPluginInfo() []plugin.PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]plugin.PluginInfo, 0, len(m.plugins))
	for _, p := range m.plugins {
		infos = append(infos, p.Info())
	}

	return infos
}

// GetPluginByName retorna um plugin pelo nome (pega a primeira versão encontrada)
func (m *Manager) GetPluginByName(name string) (plugin.Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.plugins {
		info := p.Info()
		if info.Name == name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("plugin %s não encontrado", name)
}

// StartPluginByName inicia um plugin pelo nome
func (m *Manager) StartPluginByName(ctx context.Context, name string) error {
	plugin, err := m.GetPluginByName(name)
	if err != nil {
		return err
	}

	return plugin.Start(ctx)
}

// StopPluginByName para um plugin pelo nome
func (m *Manager) StopPluginByName(ctx context.Context, name string) error {
	plugin, err := m.GetPluginByName(name)
	if err != nil {
		return err
	}

	return plugin.Stop(ctx)
}

// GetHTTPPlugin obtém um plugin HTTP pelo nome
func (m *Manager) GetHTTPPlugin(name string) (plugin.HTTPPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Buscar plugin pelo nome
	for _, p := range m.plugins {
		info := p.Info()
		if info.Name == name {
			// Verificar se é um HTTPPlugin
			if httpPlugin, ok := p.(plugin.HTTPPlugin); ok {
				return httpPlugin, nil
			}
			return nil, fmt.Errorf("plugin %s não é um plugin HTTP", name)
		}
	}

	return nil, fmt.Errorf("plugin %s não encontrado", name)
}

// GetHTTPPluginByBasePath obtém um plugin HTTP pelo base_path
func (m *Manager) GetHTTPPluginByBasePath(requestPath string) (plugin.HTTPPlugin, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Buscar plugin pelo base_path
	for _, p := range m.plugins {
		info := p.Info()
		if info.BasePath != "" && strings.HasPrefix(requestPath, info.BasePath) {
			// Verificar se é um HTTPPlugin
			if httpPlugin, ok := p.(plugin.HTTPPlugin); ok {
				return httpPlugin, info.BasePath, nil
			}
			return nil, "", fmt.Errorf("plugin %s não é um plugin HTTP", info.Name)
		}
	}

	return nil, "", fmt.Errorf("nenhum plugin encontrado para o caminho %s", requestPath)
}
