package redis

import (
	"context"
	"time"

	"github.com/wb-go/wbf/redis"
)

// Redis определяет подключение к Redis.
// Позволяет манипулировать данными в БД.
type Redis struct {
	Client redis.Client
}

// NewRedis создает новый Redis.
func NewRedis(addr string, password string, db int) (*Redis, error) {
	r := &Redis{Client: *redis.New(addr, password, db)}
	return r, r.Client.Ping(context.Background()).Err()
}

// Add добавляет новое значение по ключу.
func (r *Redis) Add(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	return r.Client.SetWithExpiration(ctx, key, value, exp)
}

// Get возвращает значение по ключу.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key)
}
