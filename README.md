# rose-cache

## WithLifeWindow 和 WithCleanWindow 的区别

- `WithLifeWindow` 是配置生命周期窗口的时间，这个时间是指缓存的生命周期从写入缓存到缓存过期的时间内，缓存在这个时间内是可用的。
- `WithCleanWindow` 是配置缓存过期后多久被清除的时间，缓存过期后，在这个时间内缓存依然可以被读取到，但是这个时间内缓存不能被写入。
