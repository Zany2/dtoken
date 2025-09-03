package gtoken

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
)

type Options struct {
	CacheMode        int8       // 缓存模式 1 gcache 2 gredis 3 gfile 默认1
	CachePreKey      string     // 缓存key前缀
	Timeout          int64      // 超时时间 默认10天（毫秒）
	MaxRefresh       int64      // 缓存刷新时间 默认为超时时间的一半（毫秒）
	MaxRefreshTimes  int        // 最大刷新次数 默认0 不限制
	TokenDelimiter   string     // Token分隔符
	EncryptKey       []byte     // Token加密key
	MultiLogin       bool       // 是否支持多端登录，默认false
	AuthExcludePaths g.SliceStr // 排除拦截地址
}

func (o *Options) String() string {
	return "gToken Options: " + fmt.Sprintf(
		"缓存模式: %d 缓存key前缀: %s 超时时间: %d 缓存刷新时间: %d 最大刷新次数: %d Token分隔符: %s 是否支持多端登录: %t 排除拦截地址: %v",
		o.CacheMode, o.CachePreKey, o.Timeout, o.MaxRefresh, o.MaxRefreshTimes, o.TokenDelimiter, o.MultiLogin, o.AuthExcludePaths)
}
