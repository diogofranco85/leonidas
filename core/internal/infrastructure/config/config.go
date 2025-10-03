package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// DatabaseConfig representa as configurações do banco de dados
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig representa as configurações do Redis
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Config representa as configurações da aplicação
type Config struct {
	Port           string
	BasePath       string
	PluginsPath    string
	ApiDescription string
	Database       DatabaseConfig
	Redis          RedisConfig
}

// Load carrega as configurações do arquivo .env e variáveis de ambiente
func Load() *Config {
	// Carregar arquivo .env se existir
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Definir valores padrão
	port := getEnv("PORT", "3000")
	basePath := getEnv("BASE_PATH", "/core")
	pluginsPath := getEnv("PLUGINS_PATH", "./plugins")
	apiDescription := getEnv("API_DESCRIPTION", "API Rest do Leonidas Core")

	// Configurações do banco de dados
	database := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "leonidas_core"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Configurações do Redis
	redis := RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0, // DB padrão
	}

	return &Config{
		Port:           port,
		BasePath:       basePath,
		PluginsPath:    pluginsPath,
		ApiDescription: apiDescription,
		Database:       database,
		Redis:          redis,
	}
}

// getEnv obtém uma variável de ambiente com valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
