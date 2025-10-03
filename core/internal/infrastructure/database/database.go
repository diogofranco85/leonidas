package database

import (
	"fmt"
	"log"

	"leonidas/core/internal/infrastructure/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database representa a conexão com o banco de dados
type Database struct {
	DB *gorm.DB
}

// NewDatabase cria uma nova conexão com o banco de dados
func NewDatabase(cfg *config.Config) (*Database, error) {
	// Construir DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	// Configurar logger do GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Conectar ao banco de dados
	var (
		err error
		db  *gorm.DB
	)
	switch cfg.Database.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "sqlserver":
		db, err = gorm.Open(sqlserver.Open(dsn), gormConfig)

	}
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	// Configurar pool de conexões
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter instância do banco: %w", err)
	}

	// Configurar pool de conexões
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Conexão com PostgreSQL estabelecida com sucesso")

	return &Database{DB: db}, nil
}

// Close fecha a conexão com o banco de dados
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping verifica se a conexão com o banco está ativa
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// AutoMigrate executa as migrações automáticas
func (d *Database) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}
