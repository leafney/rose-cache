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
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
)

var (
	ErrKeyEmpty    = errors.New("key is empty")
	ErrKeyNotFound = errors.New("key not found")
	ErrValueEmpty  = errors.New("value is empty")
	ErrNilCache    = errors.New("cache is nil")
)

type Cache struct {
	cache  *bigcache.BigCache
	mutex  sync.RWMutex
	cancel context.CancelFunc
}

type Option func(ctx *context.Context, config *bigcache.Config)

// WithContext 允许为 BigCache 实例设置自定义上下文。
// 该上下文可用于控制缓存的生命周期，允许进行取消和超时管理。
func WithContext(ctx context.Context) Option {
	return func(c *context.Context, cfg *bigcache.Config) {
		*c = ctx
	}
}

// WithLifeWindow 设置缓存中条目的有效期。
// 超过此时间后，条目将自动从缓存中删除。
// 这有助于管理内存使用，并确保不会提供过时的数据。
func WithLifeWindow(life time.Duration) Option {
	return func(c *context.Context, cfg *bigcache.Config) {
		cfg.LifeWindow = life
	}
}

// WithCleanWindow 设置缓存的清理频率。
// 在此时间段内，过期的条目将从缓存中删除。
// 这有助于保持最佳性能和内存使用，确保过期条目不会滞留在缓存中。
func WithCleanWindow(clean time.Duration) Option {
	return func(c *context.Context, cfg *bigcache.Config) {
		cfg.CleanWindow = clean
	}
}

// NewCache 返回一个新的 Cache 实例。
// 它使用提供的配置选项初始化一个新的 BigCache 实例。
// 缓存将根据提供的分钟参数具有默认的生命周期。
// 如果指定了任何选项，将应用于缓存配置。
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

// Get 根据提供的键从缓存中检索值。
// 返回值为字节切片，如果键不存在或检索过程中出现其他问题，则返回错误。
// 此方法是线程安全的。
func (c *Cache) Get(key string) ([]byte, error) {
	if c.cache == nil {
		return nil, ErrNilCache
	}
	if key == "" {
		return nil, ErrKeyEmpty
	}

	value, err := c.cache.Get(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	// 尝试将值反序列化为包装结构（用于 SetEX 值）
	wrapper := struct {
		Value     []byte    `json:"value"`
		ExpiresAt time.Time `json:"expires_at"`
	}{}

	if err := json.Unmarshal(value, &wrapper); err != nil {
		// 如果反序列化失败，则返回常规值
		return value, nil
	}

	// 检查值是否已过期
	if time.Now().After(wrapper.ExpiresAt) {
		c.cache.Delete(key)
		return nil, ErrKeyNotFound
	}

	return wrapper.Value, nil
}

// GetString 根据提供的键从缓存中检索值并返回字符串。
func (c *Cache) GetString(key string) (string, error) {
	value, err := c.Get(key)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

// GetValue 根据提供的键从缓存中检索值。
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

// Set 使用提供的键和值在缓存中设置一个值。
// 如果键为空或在设置操作中出现任何问题，则返回错误。
// 此方法是线程安全的。
func (c *Cache) Set(key string, value []byte) error {
	if c.cache == nil {
		return ErrNilCache
	}
	if key == "" {
		return ErrKeyEmpty
	}
	if len(value) == 0 {
		return ErrValueEmpty
	}

	return c.cache.Set(key, value)
}

// SetString 使用提供的键和值在缓存中设置一个字符串值。
func (c *Cache) SetString(key, value string) error {
	return c.cache.Set(key, []byte(value))
}

// SetValue 使用提供的键和值在缓存中设置一个值。
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

// SetEX 使用过期时间在缓存中设置一个值。
// 值将在指定的持续时间后从缓存中删除。
// 如果键为空或在设置操作中出现任何问题，则返回错误。
func (c *Cache) SetEX(key string, value []byte, expiration time.Duration) error {
	if c.cache == nil {
		return ErrNilCache
	}
	if key == "" {
		return ErrKeyEmpty
	}
	if len(value) == 0 {
		return ErrValueEmpty
	}

	// 创建带时间戳的包装结构
	wrapper := struct {
		Value     []byte    `json:"value"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.cache.Set(key, data)
}

// SetEXString 使用过期时间在缓存中设置一个字符串值。
// 这是一个便捷方法，先将字符串转换为字节切片，然后调用 SetEX。
func (c *Cache) SetEXString(key, value string, expiration time.Duration) error {
	return c.SetEX(key, []byte(value), expiration)
}

// Delete 根据提供的键从缓存中删除一个值。
// 如果键为空或在删除过程中出现任何问题，则返回错误。
func (c *Cache) Delete(key string) error {
	if c.cache == nil {
		return ErrNilCache
	}
	if key == "" {
		return ErrKeyEmpty
	}

	return c.cache.Delete(key)
}

// Has 检查缓存中是否存在某个键。
// 如果键存在，则返回 true，否则返回 false。
// 此方法是线程安全的。
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

// Close 关闭缓存并释放与之相关的任何资源。
// 当缓存不再需要时，应调用此方法以确保正确清理。
func (c *Cache) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.cache != nil {
		c.cache.Close()
	}
}
