package shared

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheClient struct {
	CacheConfig *CacheConfig
	Ctx         context.Context
	rdClient    *redis.Client
}

type CacheConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DB       int
}

// Return the default configuration for Redis
func RedisDefaultConfig() *CacheConfig {
	return &CacheConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
}

func NewCacheClient(config *CacheConfig) *CacheClient {
	return &CacheClient{
		CacheConfig: config,
		Ctx:         context.Background(),
	}
}

// Connect to the Redis server
// Return error if connection failed
func (c *CacheClient) Connect() error {
	c.rdClient = redis.NewClient(&redis.Options{
		Addr:     c.CacheConfig.Host + ":" + c.CacheConfig.Port,
		Password: c.CacheConfig.Password,
		DB:       c.CacheConfig.DB,
	})

	_, err := c.rdClient.Ping(c.Ctx).Result()
	if err != nil {
		return err
	}

	return nil
}

// Close the connection to the Redis server
func (c *CacheClient) Close() {
	c.rdClient.Close()
}

// Get the value of key. If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string, because GET only handles string values.
func (c *CacheClient) Get(key string) (string, error) {
	return c.rdClient.Get(c.Ctx, key).Result()
}

// Set key to hold the string value. If key already holds a value, it is overwritten, regardless of its type.
// Any previous time to live associated with the key is discarded on successful SET operation. Require TTL.
func (c *CacheClient) Set(key string, value interface{}, ttl time.Duration) error {
	return c.rdClient.Set(c.Ctx, key, value, ttl).Err()
}
