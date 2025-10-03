package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"leonidas/core/pkg/plugin"

	"gorm.io/gorm"
)

// Variável global para armazenar a conexão do banco (será definida pelo core)
var globalDB *gorm.DB

// AuthPlugin é o plugin responsável pela autenticação e autorização
type AuthPlugin struct {
	info      plugin.PluginInfo
	running   bool
	config    plugin.PluginConfig
	startTime time.Time

	// Serviços
	authService       *AuthService
	userRepo          *UserRepository
	roleRepo          *RoleRepository
	permissionRepo    *PermissionRepository
	historyRepo       *UserHistoryRepository
	sessionRepo       *UserSessionRepository
	passwordResetRepo *PasswordResetRepository
	jwtService        *JWTService
	cacheService      *CacheService
}

// NewPlugin cria uma nova instância do AuthPlugin
func NewPlugin() plugin.Plugin {
	return &AuthPlugin{
		info: plugin.PluginInfo{
			Name:        "auth-plugin",
			Version:     "1.0.0",
			Description: "Plugin de autenticação e autorização com RBAC",
			Author:      "Leonidas Team",
			BasePath:    "/auth", // Caminho base para as rotas do plugin
		},
		running: false,
	}
}

// Info retorna informações sobre o plugin
func (p *AuthPlugin) Info() plugin.PluginInfo {
	return p.info
}

// Initialize inicializa o plugin com configurações
func (p *AuthPlugin) Initialize(ctx context.Context, config plugin.PluginConfig) error {
	p.config = config
	log.Printf("[%s] Plugin de autenticação inicializado", p.info.Name)
	return nil
}

// Start inicia o plugin e executa migrations
func (p *AuthPlugin) Start(ctx context.Context) error {
	if p.running {
		return fmt.Errorf("plugin %s já está rodando", p.info.Name)
	}

	log.Printf("[%s] Iniciando plugin de autenticação...", p.info.Name)

	// Inicializar serviços (será implementado quando integrarmos com o core)
	p.initializeServices()

	// Executar migrations
	if err := p.runMigrations(ctx); err != nil {
		return fmt.Errorf("falha ao executar migrations: %w", err)
	}

	p.running = true
	p.startTime = time.Now()

	log.Printf("[%s] Plugin de autenticação iniciado com sucesso!", p.info.Name)
	return nil
}

// initializeServices inicializa todos os serviços do plugin
func (p *AuthPlugin) initializeServices() {
	// Por enquanto, vamos simular a inicialização
	// Em uma implementação real, o core passaria a conexão do banco

	// Simular conexão com banco (será implementado quando integrarmos com o core)
	db := p.getDatabaseConnection()
	if db == nil {
		log.Printf("[%s] Aviso: Conexão com banco não disponível, usando modo simulado", p.info.Name)
		// Inicializar com valores nulos para evitar panics
		p.userRepo = nil
		p.roleRepo = nil
		p.permissionRepo = nil
		p.historyRepo = nil
		p.sessionRepo = nil
		p.passwordResetRepo = nil
	} else {
		// Inicializar repositórios com conexão real
		p.userRepo = NewUserRepository(db)
		p.roleRepo = NewRoleRepository(db)
		p.permissionRepo = NewPermissionRepository(db)
		p.historyRepo = NewUserHistoryRepository(db)
		p.sessionRepo = NewUserSessionRepository(db)
		p.passwordResetRepo = NewPasswordResetRepository(db)
	}

	// Inicializar JWT service
	p.jwtService = NewJWTService("leonidas-secret-key")

	// Inicializar cache service
	p.cacheService = NewCacheService()

	// Inicializar auth service
	p.authService = NewAuthService(
		p.userRepo,
		p.sessionRepo,
		p.passwordResetRepo,
		p.historyRepo,
		p.jwtService,
		p.cacheService,
	)

	log.Printf("[%s] Serviços inicializados", p.info.Name)
}

// Stop para o plugin
func (p *AuthPlugin) Stop(ctx context.Context) error {
	if !p.running {
		return fmt.Errorf("plugin %s não está rodando", p.info.Name)
	}

	log.Printf("[%s] Parando plugin de autenticação...", p.info.Name)
	p.running = false
	log.Printf("[%s] Plugin de autenticação parado com sucesso!", p.info.Name)
	return nil
}

// IsRunning verifica se o plugin está rodando
func (p *AuthPlugin) IsRunning() bool {
	return p.running
}

// GetRoutes retorna as rotas HTTP do plugin
func (p *AuthPlugin) GetRoutes() []plugin.Route {
	return []plugin.Route{
		// Autenticação
		{Method: "POST", Path: "/auth/login"},
		{Method: "POST", Path: "/auth/refresh"},
		{Method: "POST", Path: "/auth/logout"},
		{Method: "POST", Path: "/auth/change-password"},
		{Method: "POST", Path: "/auth/forgot-password"},
		{Method: "POST", Path: "/auth/reset-password"},

		// Gerenciamento de usuários
		{Method: "GET", Path: "/users"},
		{Method: "GET", Path: "/users/:id"},
		{Method: "POST", Path: "/users"},
		{Method: "PUT", Path: "/users/:id"},
		{Method: "DELETE", Path: "/users/:id"},
		{Method: "POST", Path: "/users/:id/disable"},
		{Method: "POST", Path: "/users/:id/enable"},

		// Gerenciamento de roles
		{Method: "GET", Path: "/roles"},
		{Method: "GET", Path: "/roles/:id"},
		{Method: "POST", Path: "/roles"},
		{Method: "PUT", Path: "/roles/:id"},
		{Method: "DELETE", Path: "/roles/:id"},

		// Gerenciamento de permissões
		{Method: "GET", Path: "/permissions"},
		{Method: "GET", Path: "/permissions/:id"},
		{Method: "POST", Path: "/permissions"},
		{Method: "PUT", Path: "/permissions/:id"},
		{Method: "DELETE", Path: "/permissions/:id"},

		// Histórico
		{Method: "GET", Path: "/users/:id/history"},
		{Method: "GET", Path: "/history"},

		// Informações do plugin
		{Method: "GET", Path: "/auth/info"},
		{Method: "GET", Path: "/auth/routes"},
	}
}

// HandleRequest processa uma requisição HTTP
func (p *AuthPlugin) HandleRequest(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	if !p.running {
		return plugin.HTTPResponse{
			StatusCode: 503,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error": "Plugin não está rodando"}`),
		}, nil
	}

	// Roteamento das requisições
	switch {
	case request.Path == "/auth/login" && request.Method == "POST":
		return p.handleLogin(ctx, request)
	case request.Path == "/auth/refresh" && request.Method == "POST":
		return p.handleRefreshToken(ctx, request)
	case request.Path == "/auth/logout" && request.Method == "POST":
		return p.handleLogout(ctx, request)
	case request.Path == "/auth/change-password" && request.Method == "POST":
		return p.handleChangePassword(ctx, request)
	case request.Path == "/auth/forgot-password" && request.Method == "POST":
		return p.handleForgotPassword(ctx, request)
	case request.Path == "/auth/reset-password" && request.Method == "POST":
		return p.handleResetPassword(ctx, request)
	case request.Path == "/auth/info" && request.Method == "GET":
		return p.handleAuthInfo(ctx, request)
	case request.Path == "/auth/routes" && request.Method == "GET":
		return p.handleRoutesInfo(ctx, request)
	case request.Path == "/users" && request.Method == "GET":
		return p.handleGetUsers(ctx, request)
	case request.Path == "/users" && request.Method == "POST":
		return p.handleCreateUser(ctx, request)
	case request.Path == "/users/:id" && request.Method == "GET":
		return p.handleGetUser(ctx, request)
	case request.Path == "/users/:id" && request.Method == "PUT":
		return p.handleUpdateUser(ctx, request)
	case request.Path == "/users/:id" && request.Method == "DELETE":
		return p.handleDeleteUser(ctx, request)
	case request.Path == "/users/:id/disable" && request.Method == "POST":
		return p.handleDisableUser(ctx, request)
	case request.Path == "/users/:id/enable" && request.Method == "POST":
		return p.handleEnableUser(ctx, request)
	case request.Path == "/users/:id/history" && request.Method == "GET":
		return p.handleGetUserHistory(ctx, request)
	case request.Path == "/roles" && request.Method == "GET":
		return p.handleGetRoles(ctx, request)
	case request.Path == "/roles" && request.Method == "POST":
		return p.handleCreateRole(ctx, request)
	case request.Path == "/permissions" && request.Method == "GET":
		return p.handleGetPermissions(ctx, request)
	case request.Path == "/permissions" && request.Method == "POST":
		return p.handleCreatePermission(ctx, request)
	case request.Path == "/history" && request.Method == "GET":
		return p.handleGetHistory(ctx, request)
	default:
		return plugin.HTTPResponse{
			StatusCode: 404,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error": "Rota não encontrada"}`),
		}, nil
	}
}

// runMigrations executa as migrations do plugin
func (p *AuthPlugin) runMigrations(ctx context.Context) error {
	log.Printf("[%s] Executando migrations...", p.info.Name)

	// Obter conexão com o banco
	db := p.getDatabaseConnection()
	if db == nil {
		log.Printf("[%s] Aviso: Conexão com banco não disponível, pulando migrations", p.info.Name)
		return nil
	}

	// Executar migrations reais
	if err := p.runMigrationsFromFile(ctx); err != nil {
		return fmt.Errorf("falha ao executar migrations: %w", err)
	}

	log.Printf("[%s] Migrations executadas com sucesso", p.info.Name)
	return nil
}

// createDefaultAdmin cria o usuário administrador padrão
func (p *AuthPlugin) createDefaultAdmin(ctx context.Context) error {
	log.Printf("[%s] Verificando usuário admin padrão...", p.info.Name)

	// Aqui será implementada a criação do usuário admin
	// Por enquanto, apenas simular
	log.Printf("[%s] Usuário admin padrão verificado/criado", p.info.Name)
	return nil
}

// getDatabaseConnection obtém a conexão com o banco de dados
func (p *AuthPlugin) getDatabaseConnection() *gorm.DB {
	return globalDB
}

// SetDatabaseConnection define a conexão do banco (chamada pelo core)
func SetDatabaseConnection(db *gorm.DB) {
	globalDB = db
	log.Printf("[auth-plugin] Conexão com banco de dados definida")
}

// Handlers das rotas (implementação básica por enquanto)
func (p *AuthPlugin) handleLogin(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	var loginReq LoginRequest
	if err := json.Unmarshal(request.Body, &loginReq); err != nil {
		return plugin.HTTPResponse{
			StatusCode: 400,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error": "Payload inválido"}`),
		}, nil
	}

	// Validar campos obrigatórios
	if loginReq.Username == "" || loginReq.Password == "" {
		return plugin.HTTPResponse{
			StatusCode: 400,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error": "Username e password são obrigatórios"}`),
		}, nil
	}

	// Verificar se o auth service está disponível
	if p.authService == nil {
		return plugin.HTTPResponse{
			StatusCode: 503,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error": "Serviço de autenticação não disponível"}`),
		}, nil
	}

	// Obter informações do cliente
	ipAddress, userAgent := getClientInfo(request)

	// Tentar fazer login
	loginResp, err := p.authService.Login(ctx, loginReq, ipAddress, userAgent)
	if err != nil {
		return plugin.HTTPResponse{
			StatusCode: 401,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
		}, nil
	}

	body, _ := json.Marshal(loginResp)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) handleRefreshToken(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"message": "Refresh token endpoint - implementação em desenvolvimento",
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) handleLogout(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"message": "Logout endpoint - implementação em desenvolvimento",
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) handleChangePassword(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"message": "Change password endpoint - implementação em desenvolvimento",
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) handleForgotPassword(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"message": "Forgot password endpoint - implementação em desenvolvimento",
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) handleResetPassword(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"message": "Reset password endpoint - implementação em desenvolvimento",
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) handleAuthInfo(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"plugin":      p.info.Name,
		"version":     p.info.Version,
		"description": p.info.Description,
		"author":      p.info.Author,
		"running":     p.running,
		"uptime":      time.Since(p.startTime).String(),
		"features": []string{
			"JWT Authentication",
			"RBAC Authorization",
			"User Management",
			"Password Management",
			"Audit History",
		},
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) handleRoutesInfo(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	routes := p.GetRoutes()
	routeInfo := make([]map[string]interface{}, len(routes))

	for i, route := range routes {
		routeInfo[i] = map[string]interface{}{
			"method":      route.Method,
			"path":        route.Path,
			"description": p.getRouteDescription(route.Path, route.Method),
		}
	}

	response := map[string]interface{}{
		"plugin": p.info.Name,
		"routes": routeInfo,
		"count":  len(routes),
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

func (p *AuthPlugin) getRouteDescription(path, method string) string {
	descriptions := map[string]string{
		"POST /auth/login":           "Autentica usuário e retorna JWT token",
		"POST /auth/refresh":         "Renova JWT token",
		"POST /auth/logout":          "Invalida JWT token",
		"POST /auth/change-password": "Altera senha do usuário autenticado",
		"POST /auth/forgot-password": "Solicita recuperação de senha",
		"POST /auth/reset-password":  "Redefine senha com token",
		"GET /auth/info":             "Informações do plugin de autenticação",
		"GET /auth/routes":           "Lista todas as rotas disponíveis",
		"GET /users":                 "Lista usuários (com paginação)",
		"POST /users":                "Cria novo usuário",
		"GET /users/:id":             "Obtém dados de um usuário",
		"PUT /users/:id":             "Atualiza dados de um usuário",
		"DELETE /users/:id":          "Soft delete de um usuário",
		"POST /users/:id/disable":    "Desabilita um usuário",
		"POST /users/:id/enable":     "Habilita um usuário",
		"GET /users/:id/history":     "Histórico de alterações de um usuário",
		"GET /roles":                 "Lista roles disponíveis",
		"POST /roles":                "Cria nova role",
		"GET /permissions":           "Lista permissões disponíveis",
		"POST /permissions":          "Cria nova permissão",
		"GET /history":               "Histórico geral do sistema",
	}

	key := fmt.Sprintf("%s %s", method, path)
	if desc, exists := descriptions[key]; exists {
		return desc
	}
	return "Endpoint de autenticação e autorização"
}

// Handlers básicos para outras rotas (implementação será expandida)
func (p *AuthPlugin) handleGetUsers(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Lista de usuários - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleCreateUser(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Criar usuário - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleGetUser(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Obter usuário - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleUpdateUser(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Atualizar usuário - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleDeleteUser(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Deletar usuário - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleDisableUser(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Desabilitar usuário - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleEnableUser(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Habilitar usuário - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleGetUserHistory(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Histórico do usuário - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleGetRoles(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Lista de roles - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleCreateRole(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Criar role - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleGetPermissions(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Lista de permissões - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleCreatePermission(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Criar permissão - implementação em desenvolvimento")
}

func (p *AuthPlugin) handleGetHistory(ctx context.Context, request plugin.HTTPRequest) (plugin.HTTPResponse, error) {
	return p.createResponse("Histórico geral - implementação em desenvolvimento")
}

// createResponse cria uma resposta HTTP padrão
func (p *AuthPlugin) createResponse(message string) (plugin.HTTPResponse, error) {
	response := map[string]interface{}{
		"message": message,
		"plugin":  p.info.Name,
		"version": p.info.Version,
	}
	body, _ := json.Marshal(response)
	return plugin.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       body,
	}, nil
}

// main é necessário para compilar como plugin
func main() {
	fmt.Println("AuthPlugin compilado como plugin")
}
