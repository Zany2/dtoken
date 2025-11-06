package dtoken

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
)

// Options defines all configuration for gToken | gToken 全局配置参数
type Options struct {
	CacheMode        int8       // Cache mode: 1-gcache 2-gredis 3-gfile | 缓存模式：1 gcache 2 gredis 3 gfile
	CachePreKey      string     // Cache key prefix | 缓存 key 前缀
	Timeout          int64      // Token expiration time (ms) | Token 超时时间（毫秒）
	MaxRefresh       int64      // Max auto-refresh interval (ms) | 最大自动刷新间隔（毫秒）
	MaxRefreshTimes  int        // Maximum number of refresh times (0 = unlimited) | 最大刷新次数（0 表示不限制）
	TokenDelimiter   string     // Token delimiter | Token 分隔符
	EncryptKey       []byte     // Token encryption key | Token 加密密钥
	MultiLogin       bool       // Allow multi-login | 是否允许多端登录
	AuthExcludePaths g.SliceStr // Paths excluded from authentication | 免认证路径列表

	PoolMinSize       int     // Minimum pool size | 最小协程数
	PoolMaxSize       int     // Maximum pool size | 最大协程数
	PoolScaleUpRate   float64 // Scale-up threshold (expand when usage exceeds this ratio) | 扩容阈值，当使用率超过此比例时扩容
	PoolScaleDownRate float64 // Scale-down threshold (shrink when usage below this ratio) | 缩容阈值，当使用率低于此比例时缩容
	RenewInterval     int64   // Minimum renewal interval (ms) | 最小续期间隔（毫秒）
}

// String returns a formatted configuration summary | 返回格式化的配置摘要
func (o *Options) String() string {
	// Helper: format key-value pairs into table-like output
	lines := []string{
		fmt.Sprintf("│ CacheMode        │ %-55v │", fmt.Sprintf("%d (1-gcache 2-gredis 3-gfile)", o.CacheMode)),
		fmt.Sprintf("│ CachePreKey      │ %-55v │", o.CachePreKey),
		fmt.Sprintf("│ Timeout          │ %-55v │", fmt.Sprintf("%d ms", o.Timeout)),
		fmt.Sprintf("│ MaxRefresh       │ %-55v │", fmt.Sprintf("%d ms", o.MaxRefresh)),
		fmt.Sprintf("│ MaxRefreshTimes  │ %-55v │", o.MaxRefreshTimes),
		fmt.Sprintf("│ TokenDelimiter   │ %-55v │", o.TokenDelimiter),
		fmt.Sprintf("│ EncryptKey       │ %-55v │", string(o.EncryptKey)),
		fmt.Sprintf("│ MultiLogin       │ %-55v │", o.MultiLogin),
		fmt.Sprintf("│ PoolMinSize      │ %-55v │", o.PoolMinSize),
		fmt.Sprintf("│ PoolMaxSize      │ %-55v │", o.PoolMaxSize),
		fmt.Sprintf("│ PoolScaleUpRate  │ %-55v │", fmt.Sprintf("%.2f", o.PoolScaleUpRate)),
		fmt.Sprintf("│ PoolScaleDownRate│ %-55v │", fmt.Sprintf("%.2f", o.PoolScaleDownRate)),
		fmt.Sprintf("│ RenewInterval    │ %-55v │", fmt.Sprintf("%d ms", o.RenewInterval)),
		fmt.Sprintf("│ AuthExcludePaths │ %-55v │", strings.Join(o.AuthExcludePaths, ", ")),
	}

	// Build visual box
	header := "------------------------------------------------------\n" +
		"│                gToken Configuration                 │\n" +
		"------------------------------------------------------"
	footer := "------------------------------------------------------"

	return fmt.Sprintf("\n%s\n%s\n%s\n", header, strings.Join(lines, "\n"), footer)
}
