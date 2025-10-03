package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"leonidas/core/internal/infrastructure/config"

	"github.com/redis/go-redis/v9"
)

// Redis representa a conexão com o Redis
type Redis struct {
	Client *redis.Client
}

// NewRedis cria uma nova conexão com o Redis
func NewRedis(cfg *config.Config) (*Redis, error) {
	// Configurar opções do Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Testar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao Redis: %w", err)
	}

	log.Println("Conexão com Redis estabelecida com sucesso")

	return &Redis{Client: rdb}, nil
}

// Close fecha a conexão com o Redis
func (r *Redis) Close() error {
	return r.Client.Close()
}

// Ping verifica se a conexão com o Redis está ativa
func (r *Redis) Ping(ctx context.Context) error {
	_, err := r.Client.Ping(ctx).Result()
	return err
}

// Set define um valor no Redis
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get obtém um valor do Redis
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del remove uma chave do Redis
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists verifica se uma chave existe no Redis
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// Expire define o tempo de expiração de uma chave
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}
