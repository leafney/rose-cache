/**
 * @Author:      leafney
 * @GitHub:      https://github.com/leafney
 * @Project:     rose-cache
 * @Date:        2023-05-16 22:07
 * @Description:
 */

package rcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"sync"
	"time"
)

type Cache struct {
	cache  *bigcache.BigCache
	mutex  sync.RWMutex
	cancel context.CancelFunc
}

type Option func(ctx *context.Context, config *bigcache.Config)

// WithContext allows setting a custom context for the BigCache instance.
func WithContext(ctx context.Context) Option {
	return func(c *context.Context, cfg *bigcache.Config) {
		*c = ctx
	}
}

func WithLifeWindow(life time.Duration) Option {
	return func(c *context.Context, cfg *bigcache.Config) {
		cfg.LifeWindow = life
	}
}

func WithCleanWindow(clean time.Duration) Option {
	return func(c *context.Context, cfg *bigcache.Config) {
		cfg.CleanWindow = clean
	}
}

// NewCache returns a new instance of the Cache struct.
func NewCache(minute int64, opts ...Option) (*Cache, error) {
	ctx, cancel := context.WithCancel(context.Background())
	config := bigcache.DefaultConfig(time.Duration(minute) * time.Minute)

	for _, opt := range opts {
		opt(&ctx, &config)
	}

	cache, err := bigcache.New(ctx, config)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Cache{cache: cache, cancel: cancel}, nil
}

//// NewCache returns a new instance of the Cache struct.
//func NewCache(ctx context.Context, minute int64) (*Cache, error) {
//	ctx, cancel := context.WithCancel(ctx)
//	config := bigcache.DefaultConfig(time.Duration(minute) * time.Minute)
//	cache, err := bigcache.New(ctx, config)
//	if err != nil {
//		cancel()
//		return nil, err
//	}
//
//	return &Cache{cache: cache, cancel: cancel}, nil
//}

// Get retrieves a value from the cache using the provided key.
func (c *Cache) Get(key string) ([]byte, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, err := c.cache.Get(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// GetString retrieves a value from the cache using the provided key and returns it as a string.
func (c *Cache) GetString(key string) (string, error) {
	value, err := c.Get(key)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

// GetValue retrieves a value from the cache using the provided key.
func (c *Cache) GetValue(key string, value interface{}) error {
	data, err := c.cache.Get(key)
	if err != nil {
		return err
	}

	switch value := value.(type) {
	case *string:
		*value = string(data)
	default:
		if err := json.Unmarshal(data, value); err != nil {
			return fmt.Errorf("failed to unmarshal data: %v", err)
		}
	}

	return nil
}

// Set sets a value in the cache using the provided key and value.
func (c *Cache) Set(key string, value []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.cache.Set(key, value)
}

// SetString sets a value in the cache using the provided key and value.
func (c *Cache) SetString(key, value string) error {
	return c.cache.Set(key, []byte(value))
}

// SetValue sets a value in the cache using the provided key and value.
func (c *Cache) SetValue(key string, value interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	switch v := value.(type) {
	case string:
		return c.cache.Set(key, []byte(v))
	case []byte:
		return c.cache.Set(key, v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %v", err)
		}
		return c.cache.Set(key, data)
	}
}

// Delete removes a value from the cache using the provided key.
func (c *Cache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.cache.Delete(key)
}

func (c *Cache) Has(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, err := c.cache.Get(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return false
		}
		return false
	}
	return true
}

func (c *Cache) Close() {
	c.cancel()
}
