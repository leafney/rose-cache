/**
 * @Author:      leafney
 * @GitHub:      https://github.com/leafney
 * @Project:     rose-cache
 * @Date:        2023-05-16 22:07
 * @Description:
 */

package rcache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"strconv"
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

// Cache 结构体封装了对 bigcache 的操作
type Cache struct {
	cache  *bigcache.BigCache
	mutex  sync.RWMutex // 读写锁，适用于读多写少的场景
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
// minute 参数指定默认的缓存生命周期（分钟）
// opts 可选参数用于自定义缓存配置
//
// Example:
//
//	// 创建一个默认10分钟过期的缓存
//	cache, err := NewCache(10)
//
//	// 创建一个自定义配置的缓存
//	cache, err := NewCache(10,
//	    WithLifeWindow(5*time.Minute),
//	    WithCleanWindow(1*time.Minute),
//	)
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

	return &Cache{cache: cache, cancel: cancel, mutex: sync.RWMutex{}}, nil
}

// Close 关闭缓存并释放资源
// 在程序结束时调用此方法以确保资源被正确释放
//
// Example:
//
//	cache, err := NewCache(10)
//	if err != nil {
//	    // 处理错误
//	}
//	defer cache.Close()
func (c *Cache) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cancel != nil {
		c.cancel()
	}
	if c.cache != nil {
		return c.cache.Close()
	}
	return ErrNilCache
}

// Del 删除指定键的缓存项
func (c *Cache) Del(key string) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if c.cache == nil {
		return ErrNilCache
	}
	return c.cache.Delete(key)
}

// Exists 检查键是否存在于缓存中
func (c *Cache) Exists(key string) bool {
	if key == "" || c.cache == nil {
		return false
	}
	_, err := c.cache.Get(key)
	return err == nil
}

// Get 获取指定键的缓存值
func (c *Cache) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, ErrKeyEmpty
	}
	if c.cache == nil {
		return nil, ErrNilCache
	}

	value, err := c.cache.Get(key)
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		return nil, ErrKeyNotFound
	}
	return value, err
}

// GetS 获取指定键的字符串类型缓存值
func (c *Cache) GetS(key string) (string, error) {
	value, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// Set 设置缓存键值对
func (c *Cache) Set(key string, value []byte) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if value == nil {
		return ErrValueEmpty
	}
	if c.cache == nil {
		return ErrNilCache
	}

	return c.cache.Set(key, value)
}

// SetS 设置字符串类型的缓存键值对
func (c *Cache) SetS(key string, value string) error {
	if value == "" {
		return ErrValueEmpty
	}
	return c.Set(key, []byte(value))
}

// *******************

// CacheType 定义缓存数据结构
type CacheType struct {
	Data   []byte
	Expire int64 // 过期时间戳（0表示永不过期）
}

// XSet 使用 gob 序列化存储数据，数据永不过期
//
// Example:
//
//	err := cache.XSet("key", []byte("value"))
func (c *Cache) XSet(key string, value []byte) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if value == nil {
		return ErrValueEmpty
	}
	if c.cache == nil {
		return ErrNilCache
	}

	// 创建 CacheType 实例
	ct := &CacheType{
		Data:   value,
		Expire: 0, // 默认不过期
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(ct); err != nil {
		return err
	}

	return c.cache.Set(key, buf.Bytes())
}

// XSetS 使用 gob 序列化存储字符串数据，数据永不过期
//
// Example:
//
//	err := cache.XSetS("key", "value")
func (c *Cache) XSetS(key string, value string) error {
	if value == "" {
		return ErrValueEmpty
	}
	return c.XSet(key, []byte(value))
}

// XGet 获取 gob 序列化存储的数据
// 如果数据已过期，会自动删除并返回 ErrKeyNotFound
//
// Example:
//
//	data, err := cache.XGet("key")
//	if err == ErrKeyNotFound {
//	    // 处理键不存在的情况
//	}
func (c *Cache) XGet(key string) ([]byte, error) {
	if key == "" {
		return nil, ErrKeyEmpty
	}
	if c.cache == nil {
		return nil, ErrNilCache
	}

	// 获取原始数据
	data, err := c.cache.Get(key)
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	// 解码 gob 数据
	var ct CacheType
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&ct); err != nil {
		return nil, err
	}

	// 检查是否过期
	if ct.Expire > 0 && ct.Expire <= time.Now().Unix() {
		c.Del(key) // 删除过期数据
		return nil, ErrKeyNotFound
	}

	return ct.Data, nil
}

// XGetS 获取 gob 序列化存储的字符串数据
//
// Example:
//
//	str, err := cache.XGetS("key")
//	if err != nil {
//	    // 处理错误
//	}
func (c *Cache) XGetS(key string) (string, error) {
	data, err := c.XGet(key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// XSetEx 使用 Duration 类型设置过期时间存储数据
//
// Example:
//
//	// 存储10分钟后过期的数据
//	err := cache.XSetEx("key", []byte("value"), 10*time.Minute)
func (c *Cache) XSetEx(key string, value []byte, expires time.Duration) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if value == nil {
		return ErrValueEmpty
	}
	if c.cache == nil {
		return ErrNilCache
	}

	ct := &CacheType{
		Data:   value,
		Expire: time.Now().Add(expires).Unix(),
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(ct); err != nil {
		return err
	}

	return c.cache.Set(key, buf.Bytes())
}

// XSetExS 使用 Duration 类型设置过期时间存储字符串数据
//
// Example:
//
//	// 存储1小时后过期的字符串
//	err := cache.XSetExS("key", "value", time.Hour)
func (c *Cache) XSetExS(key string, value string, expires time.Duration) error {
	if value == "" {
		return ErrValueEmpty
	}
	return c.XSetEx(key, []byte(value), expires)
}

// XSetExSec 使用秒数设置过期时间存储数据
func (c *Cache) XSetExSec(key string, value []byte, seconds int64) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if value == nil {
		return ErrValueEmpty
	}
	if c.cache == nil {
		return ErrNilCache
	}

	ct := &CacheType{
		Data:   value,
		Expire: time.Now().Add(time.Duration(seconds) * time.Second).Unix(),
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(ct); err != nil {
		return err
	}

	return c.cache.Set(key, buf.Bytes())
}

// XSetExSecS 使用秒数设置过期时间存储字符串数据
func (c *Cache) XSetExSecS(key string, value string, seconds int64) error {
	if value == "" {
		return ErrValueEmpty
	}
	return c.XSetExSec(key, []byte(value), seconds)
}

// XExpireAt 设置键在指定时间点过期
//
// Example:
//
//	// 设置在1小时后过期
//	err := cache.XExpireAt("key", time.Now().Add(time.Hour))
//
//	// 设置在明天零点过期
//	tomorrow := time.Now().Add(24*time.Hour)
//	expireTime := time.Date(
//	    tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
//	    0, 0, 0, 0, tomorrow.Location(),
//	)
//	err := cache.XExpireAt("key", expireTime)
func (c *Cache) XExpireAt(key string, tm time.Time) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if c.cache == nil {
		return ErrNilCache
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	expireAt := tm.Unix()
	if expireAt <= time.Now().Unix() {
		return c.Del(key)
	}

	// 获取原始数据
	data, err := c.cache.Get(key)
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		return ErrKeyNotFound
	}
	if err != nil {
		return err
	}

	// 解码现有数据
	var ct CacheType
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&ct); err != nil {
		return err
	}

	// 更新过期时间
	ct.Expire = expireAt

	// 重新编码并存储
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&ct); err != nil {
		return err
	}

	return c.cache.Set(key, buf.Bytes())
}

// XExpire 设置键的过期时间（Duration类型）
//
// Example:
//
//	// 设置5分钟后过期
//	err := cache.XExpire("key", 5*time.Minute)
func (c *Cache) XExpire(key string, expires time.Duration) error {
	return c.XExpireAt(key, time.Now().Add(expires))
}

// XExpireSec 设置键的过期时间（秒数）
func (c *Cache) XExpireSec(key string, seconds int64) error {
	return c.XExpireAt(key, time.Now().Add(time.Duration(seconds)*time.Second))
}

// XTTL 获取键的剩余生存时间（秒）
// 返回值说明：
//   - -2: 键不存在
//   - -1: 键存在但没有设置过期时间
//   - >= 0: 剩余生存时间（秒）
//
// Example:
//
//	ttl, err := cache.XTTL("key")
//	switch ttl {
//	case -2:
//	    fmt.Println("键不存在")
//	case -1:
//	    fmt.Println("键永不过期")
//	default:
//	    fmt.Printf("剩余 %d 秒\n", ttl)
//	}
func (c *Cache) XTTL(key string) (int64, error) {
	if key == "" {
		return -2, ErrKeyEmpty
	}
	if c.cache == nil {
		return -2, ErrNilCache
	}

	// 获取原始数据
	data, err := c.cache.Get(key)
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		return -2, nil // 键不存在返回 -2
	}
	if err != nil {
		return -2, err
	}

	// 解码数据
	var ct CacheType
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&ct); err != nil {
		return -2, err
	}

	// 如果没有设置过期时间
	if ct.Expire == 0 {
		return -1, nil // 永不过期返回 -1
	}

	// 计算剩余时间
	remaining := ct.Expire - time.Now().Unix()
	if remaining <= 0 {
		c.Del(key) // 已过期，删除键
		return -2, nil
	}

	return remaining, nil
}

// XIncr 将键存储的数字值加1
// 如果键不存在，会创建并设置值为1
//
// Example:
//
//	newVal, err := cache.XIncr("counter")
//	// newVal 是增加后的新值
func (c *Cache) XIncr(key string) (int64, error) {
	return c.XIncrBy(key, 1)
}

// XIncrBy 将键存储的数字值增加指定的增量
//
// Example:
//
//	// 增加5
//	newVal, err := cache.XIncrBy("counter", 5)
//
//	// 减少3（通过负数实现）
//	newVal, err := cache.XIncrBy("counter", -3)
func (c *Cache) XIncrBy(key string, increment int64) (int64, error) {
	if key == "" {
		return 0, ErrKeyEmpty
	}
	if c.cache == nil {
		return 0, ErrNilCache
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 尝试获取现有值
	var currentVal int64
	data, err := c.XGet(key)
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		return 0, err
	}

	if err == nil {
		// 如果键存在，尝试转换为int64
		currentVal, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return 0, errors.New("value is not an integer")
		}
	}

	// 执行增加操作
	newVal := currentVal + increment

	// 存储新值
	err = c.XSet(key, []byte(strconv.FormatInt(newVal, 10)))
	if err != nil {
		return 0, err
	}

	return newVal, nil
}

// XDecr 将键存储的数字值减1
//
// Example:
//
//	newVal, err := cache.XDecr("counter")
func (c *Cache) XDecr(key string) (int64, error) {
	return c.XDecrBy(key, 1)
}

// XDecrBy 将键存储的数字值减少指定的减量
//
// Example:
//
//	// 减少5
//	newVal, err := cache.XDecrBy("counter", 5)
func (c *Cache) XDecrBy(key string, decrement int64) (int64, error) {
	return c.XIncrBy(key, -decrement)
}
