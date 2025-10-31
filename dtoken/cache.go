package dtoken

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"time"
)

// Cache defines the cache interface | 缓存接口定义
type Cache interface {
	// Set sets the cache value | 设置缓存值
	Set(ctx context.Context, cacheKey string, cacheValue g.Map) error
	// Get retrieves the cache value | 获取缓存值
	Get(ctx context.Context, cacheKey string) (g.Map, error)
	// Remove removes the cache value | 删除缓存值
	Remove(ctx context.Context, cacheKey string) error
}

// DefaultCache implements the default cache | 默认缓存实现
type DefaultCache struct {
	Cache   *gcache.Cache // Cache instance | 缓存实例
	Mode    int8          // Cache mode: 1 for gcache, 2 for gredis, 3 for gfile | 缓存模式：1为gcache，2为gredis，3为gfile
	PreKey  string        // Cache key prefix | 缓存key前缀
	Timeout int64         // Timeout in milliseconds | 超时时间，单位毫秒
}

// NewDefaultCache creates a new DefaultCache instance | 创建新的默认缓存实例
func NewDefaultCache(mode int8, preKey string, timeout int64) *DefaultCache {
	c := &DefaultCache{
		Cache:   gcache.New(),
		Mode:    mode,
		PreKey:  preKey,
		Timeout: timeout,
	}

	// Initialize the cache based on mode | 根据模式初始化缓存
	if c.Mode == CacheModeFile {
		c.initFileCache(gctx.New()) // Initialize file cache | 初始化文件缓存
	} else if c.Mode == CacheModeRedis {
		c.Cache.SetAdapter(gcache.NewAdapterRedis(g.Redis())) // Initialize Redis cache | 初始化 Redis 缓存
	}

	return c
}

// Set sets a cache value | 设置缓存值
func (c *DefaultCache) Set(ctx context.Context, cacheKey string, cacheValue g.Map) error {
	if cacheValue == nil {
		return errors.New(MsgErrDataEmpty) // Error if cache value is empty | 如果缓存值为空，返回错误
	}
	value, err := gjson.Encode(cacheValue) // Encode cache value to JSON | 将缓存值编码为 JSON
	if err != nil {
		return err
	}
	err = c.Cache.Set(ctx, c.PreKey+cacheKey, string(value), gconv.Duration(c.Timeout)*time.Millisecond) // Set cache with timeout | 设置缓存并设置超时
	if err != nil {
		return err
	}
	if c.Mode == CacheModeFile {
		c.writeFileCache(ctx) // Write cache to file if file cache mode is used | 如果是文件缓存模式，则将缓存写入文件
	}
	return nil
}

// Get retrieves a cache value | 获取缓存值
func (c *DefaultCache) Get(ctx context.Context, cacheKey string) (g.Map, error) {
	dataVar, err := c.Cache.Get(ctx, c.PreKey+cacheKey) // Get the cache value | 获取缓存值
	if err != nil {
		return nil, err
	}
	if dataVar.IsNil() {
		return nil, nil // Return nil if cache value is empty | 如果缓存值为空，则返回 nil
	}
	return dataVar.Map(), nil
}

// Remove removes a cache value | 删除缓存值
func (c *DefaultCache) Remove(ctx context.Context, cacheKey string) error {
	_, err := c.Cache.Remove(ctx, c.PreKey+cacheKey) // Remove cache | 删除缓存
	if c.Mode == CacheModeFile {
		c.writeFileCache(ctx) // Write cache to file after removal | 删除后将缓存写入文件
	}
	return err
}

// writeFileCache writes the cache data to a file | 将缓存数据写入文件
func (c *DefaultCache) writeFileCache(ctx context.Context) {
	fileName := gstr.Replace(c.PreKey, ":", "_") + CacheModeFileDat // Generate file name | 生成文件名
	file := gfile.Temp(fileName)                                    // Create temporary file | 创建临时文件
	data, e := c.Cache.Data(ctx)                                    // Get cache data | 获取缓存数据
	if e != nil {
		g.Log().Error(ctx, "[GToken]cache writeFileCache data error", e) // Log error if data retrieval fails | 获取数据失败时记录错误
	}
	e = gfile.PutContents(file, gjson.New(data).MustToJsonString()) // Write data to file | 将数据写入文件
	if e != nil {
		g.Log().Error(ctx, "[GToken]cache writeFileCache put error", e) // Log error if writing to file fails | 写入文件失败时记录错误
	}
}

// initFileCache initializes the file cache | 初始化文件缓存
func (c *DefaultCache) initFileCache(ctx context.Context) {
	fileName := gstr.Replace(c.PreKey, ":", "_") + CacheModeFileDat // Generate file name | 生成文件名
	file := gfile.Temp(fileName)                                    // Create temporary file | 创建临时文件
	g.Log().Debug(ctx, "file cache init", file)                     // Log the file cache initialization | 记录文件缓存初始化
	if !gfile.Exists(file) {
		return // Return if the file does not exist | 如果文件不存在，则返回
	}
	data := gfile.GetContents(file) // Read data from file | 从文件读取数据
	maps := gconv.Map(data)         // Convert data to map | 将数据转换为 map
	if maps == nil || len(maps) <= 0 {
		return // Return if no data is found in the file | 如果文件中没有数据，则返回
	}
	// Load the cache data from file | 从文件加载缓存数据
	for k, v := range maps {
		_ = c.Cache.Set(ctx, k, v, gconv.Duration(c.Timeout)*time.Millisecond)
	}
}
