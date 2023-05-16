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
	"fmt"
	"github.com/allegro/bigcache/v3"
	"time"
)

type Cache struct {
	cache *bigcache.BigCache
}

// NewCache returns a new instance of the Cache struct.
func NewCache(ctx context.Context, min int64) (*Cache, error) {
	config := bigcache.DefaultConfig(time.Duration(min) * time.Minute)
	cache, err := bigcache.New(ctx, config)
	if err != nil {
		return nil, err
	}

	return &Cache{cache: cache}, nil
}

// Get retrieves a value from the cache using the provided key.
func (c *Cache) Get(key string) ([]byte, error) {
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
	return c.cache.Set(key, value)
}

// SetString sets a value in the cache using the provided key and value.
func (c *Cache) SetString(key, value string) error {
	return c.cache.Set(key, []byte(value))
}

// SetValue sets a value in the cache using the provided key and value.
func (c *Cache) SetValue(key string, value interface{}) error {
	var data []byte

	switch value := value.(type) {
	case string:
		data = []byte(value)
	case []byte:
		data = value
	default:
		var err error
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %v", err)
		}
	}

	return c.cache.Set(key, data)
}

// Delete removes a value from the cache using the provided key.
func (c *Cache) Delete(key string) error {
	return c.cache.Delete(key)
}
