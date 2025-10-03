package http

import (
	"context"
	"fmt"
	"log"

	"leonidas/core/internal/infrastructure/cache"
	"leonidas/core/internal/infrastructure/config"
	"leonidas/core/internal/infrastructure/database"
	"leonidas/core/internal/infrastructure/plugin"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/helmet/v2"
)

// Server representa o servidor HTTP
type Server struct {
	app           *fiber.App
	config        *config.Config
	healthHandler *HealthHandler
	database      *database.Database
	redis         *cache.Redis
	pluginManager *plugin.Manager
}

// NewServer cria uma nova instância do servidor
func NewServer() *Server {
	cfg := config.Load()

	// Conectar ao banco de dados
	database, err := database.NewDatabase(cfg)
	if err != nil {
		log.Printf("Aviso: Falha ao conectar ao banco de dados: %v", err)
		database = nil
	}

	// Conectar ao Redis
	redis, err := cache.NewRedis(cfg)
	if err != nil {
		log.Printf("Aviso: Falha ao conectar ao Redis: %v", err)
		redis = nil
	}

	// Criar gerenciador de plugins
	pluginManager := plugin.NewManager(cfg)

	// Carregar plugins
	if err := pluginManager.LoadPlugins(context.Background()); err != nil {
		log.Printf("Aviso: Falha ao carregar plugins: %v", err)
	}
	// Criar health handler com as conexões
	healthHandler := NewHealthHandler(cfg, database, redis, pluginManager)

	app := fiber.New(fiber.Config{
		AppName: "Leonidas Core API",
	})

	// Middlewares
	app.Use(helmet.New())
	app.Use(logger.New())

	// Configurar rotas
	setupRoutes(app, cfg, healthHandler, database, redis, pluginManager)

	return &Server{
		app:           app,
		config:        cfg,
		healthHandler: healthHandler,
		database:      database,
		redis:         redis,
		pluginManager: pluginManager,
	}
}

// setupRoutes configura as rotas da aplicação
func setupRoutes(app *fiber.App, cfg *config.Config, healthHandler *HealthHandler, database *database.Database, redis *cache.Redis, pluginManager *plugin.Manager) {
	// Rotas de health check
	app.Get(cfg.BasePath+"/health", healthHandler.BasicHealth)
	app.Get(cfg.BasePath+"/health/detailed", healthHandler.DetailedHealth)
	app.Get(cfg.BasePath+"/health/ready", healthHandler.ReadinessCheck)
	app.Get(cfg.BasePath+"/health/live", healthHandler.LivenessCheck)

	// Grupo de rotas com base path
	api := app.Group(cfg.BasePath)
	{
		// Exemplo de rota da API
		api.Get("/", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"message": cfg.ApiDescription,
				"version": "1.0.0",
			})
		})

		// Rotas de plugins
		api.Get("/plugins", func(c *fiber.Ctx) error {
			plugins := pluginManager.GetAllPlugins()
			pluginInfos := make([]fiber.Map, 0, len(plugins))

			for _, p := range plugins {
				info := p.Info()
				pluginInfos = append(pluginInfos, fiber.Map{
					"name":        info.Name,
					"version":     info.Version,
					"description": info.Description,
					"author":      info.Author,
					"running":     p.IsRunning(),
				})
			}

			return c.JSON(fiber.Map{
				"plugins": pluginInfos,
				"count":   len(pluginInfos),
			})
		})

		// Rota para obter informações de um plugin específico
		api.Get("/plugins/:name/:version", func(c *fiber.Ctx) error {
			name := c.Params("name")
			version := c.Params("version")

			plugin, err := pluginManager.GetPlugin(name, version)
			if err != nil {
				return c.Status(404).JSON(fiber.Map{
					"error": "Plugin não encontrado",
				})
			}

			info := plugin.Info()
			return c.JSON(fiber.Map{
				"name":        info.Name,
				"version":     info.Version,
				"description": info.Description,
				"author":      info.Author,
				"running":     plugin.IsRunning(),
			})
		})

		// Rota para iniciar um plugin
		api.Post("/plugins/:name/:version/start", func(c *fiber.Ctx) error {
			name := c.Params("name")
			version := c.Params("version")

			if err := pluginManager.StartPlugin(context.Background(), name, version); err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"message": "Plugin iniciado com sucesso",
			})
		})

		// Rota para parar um plugin
		api.Post("/plugins/:name/:version/stop", func(c *fiber.Ctx) error {
			name := c.Params("name")
			version := c.Params("version")

			if err := pluginManager.StopPlugin(context.Background(), name, version); err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"message": "Plugin parado com sucesso",
			})
		})
	}
}

// Start inicia o servidor
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Port)
	return s.app.Listen(addr)
}

// Shutdown para o servidor graciosamente
func (s *Server) Shutdown() error {
	// Fechar conexões com banco e Redis
	if s.database != nil {
		if err := s.database.Close(); err != nil {
			log.Printf("Erro ao fechar conexão com banco de dados: %v", err)
		}
	}

	if s.redis != nil {
		if err := s.redis.Close(); err != nil {
			log.Printf("Erro ao fechar conexão com Redis: %v", err)
		}
	}

	return s.app.Shutdown()
}
