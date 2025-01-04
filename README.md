# rose-cache

一个基于 [bigcache](https://github.com/allegro/bigcache) 的 Go 缓存库，提供了更丰富的缓存操作接口。

## 特性

- 支持设置缓存过期时间
- 支持字符串和字节数组存储
- 提供计数器操作（递增/递减）
- 线程安全
- 自动清理过期数据

## 安装

```bash
go get github.com/leafney/rose-cache
```

## 快速开始

```go
// 创建一个默认10分钟过期的缓存实例
cache, err := rcache.NewCache(10)
if err != nil {
    log.Fatal(err)
}
defer cache.Close()
// 存储数据
err = cache.XSetEx("key", []byte("value"), 5time.Minute)
// 获取数据
val, err := cache.XGet("key")
```

## 配置选项

### WithLifeWindow

设置缓存中条目的有效期。超过此时间后，条目将自动从缓存中删除。

### WithCleanWindow

设置缓存的清理间隔。每隔此时间段后，缓存会自动检查并删除过期的条目，以保持最佳性能和内存使用。

例如：设置为 10 分钟，则每隔 10 分钟会进行一次过期数据的清理操作。

### WithContext

允许为 BigCache 实例设置自定义上下文，用于控制缓存的生命周期。

## API 文档

### 基础操作

```go
// 创建新的缓存实例
func NewCache(minute int64, opts ...Option) (*Cache, error)

// 关闭缓存并释放资源
func (c *Cache) Close() error

// 检查键是否存在
func (c *Cache) Exists(key string) bool

// 删除指定键的缓存项
func (c *Cache) Del(key string) error
```

### 基本存取操作

```go
// 设置缓存键值对
func (c *Cache) Set(key string, value []byte) error
func (c *Cache) SetS(key string, value string) error

// 获取缓存值
func (c *Cache) Get(key string) ([]byte, error)
func (c *Cache) GetS(key string) (string, error)
```

### 扩展存取操作（支持过期时间）

```go
// 存储数据（支持过期时间）
func (c *Cache) XSet(key string, value []byte) error
func (c *Cache) XSetS(key string, value string) error
func (c *Cache) XSetEx(key string, value []byte, expires time.Duration) error
func (c *Cache) XSetExS(key string, value string, expires time.Duration) error
func (c *Cache) XSetExSec(key string, value []byte, seconds int64) error
func (c *Cache) XSetExSecS(key string, value string, seconds int64) error

// 获取数据
func (c *Cache) XGet(key string) ([]byte, error)
func (c *Cache) XGetS(key string) (string, error)
```

### 过期时间操作

```go
// 设置过期时间
func (c *Cache) XExpireAt(key string, tm time.Time) error
func (c *Cache) XExpire(key string, expires time.Duration) error
func (c *Cache) XExpireSec(key string, seconds int64) error

// 获取剩余生存时间（秒）
func (c *Cache) XTTL(key string) (int64, error)
```

### 计数器操作

```go
// 递增操作
func (c *Cache) XIncr(key string) (int64, error)
func (c *Cache) XIncrBy(key string, increment int64) (int64, error)

// 递减操作
func (c *Cache) XDecr(key string) (int64, error)
func (c *Cache) XDecrBy(key string, decrement int64) (int64, error)
```

## WithLifeWindow 和 WithCleanWindow 的区别

- `WithLifeWindow` -- 用来设置缓存的生命周期时长。
- `WithCleanWindow` -- 用来设置缓存过期后多久被清除的时间，超过该时间，所有过期的条目会被删除，仍在有效期内的条目不会被删除。

## 错误处理

库定义了以下错误类型：

```go
var (
    ErrKeyEmpty    = errors.New("key is empty")
    ErrKeyNotFound = errors.New("key not found")
    ErrValueEmpty  = errors.New("value is empty")
    ErrNilCache    = errors.New("cache is nil")
)
```
