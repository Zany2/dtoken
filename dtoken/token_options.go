package dtoken

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"runtime"
)

// Version version number | 版本号
const Version = "v1.0.0"

const Banner = `
          ____  ______      __            
         / __ \/_  __/___  / /_____  ____ 
        / / / / / / / __ \/ //_/ _ \/ __ \
       / /_/ / / / / /_/ / ,< /  __/ / / /
      /_____/ /_/  \____/_/|_|\___/_/ /_/ 

:: DToken-Go ::                               %s
`

const (
	boxWidth   = 63
	labelWidth = 22
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

// PrintBanner prints startup banner only | 打印启动横幅
func PrintBanner() {
	fmt.Printf(Banner, Version)
	fmt.Printf(":: Go Version ::                              %s\n", runtime.Version())
	fmt.Printf(":: GOOS/GOARCH ::                             %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
}

// formatLine formats configuration rows with alignment | 格式化配置行
func formatLine(label string, value any) string {
	valueStr := fmt.Sprintf("%v", value)
	valueWidth := boxWidth - labelWidth - 5
	return fmt.Sprintf("│ %-*s: %-*s │\n", labelWidth, label, valueWidth, valueStr)
}

// PrintWithOptions prints banner with configuration summary | 打印启动横幅和配置信息
func PrintWithOptions(opt *Options) {
	PrintBanner()

	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│                       GToken Configuration                   │")
	fmt.Println("├──────────────────────────────────────────────────────────────┤")

	// Cache and storage | 缓存与存储配置
	fmt.Print(formatLine("Cache Mode", fmt.Sprintf("%d (1-gcache 2-gredis 3-gfile)", opt.CacheMode)))
	fmt.Print(formatLine("Cache PreKey", opt.CachePreKey))
	fmt.Print(formatLine("Timeout", fmt.Sprintf("%d ms", opt.Timeout)))
	fmt.Print(formatLine("Max Refresh", fmt.Sprintf("%d ms", opt.MaxRefresh)))
	fmt.Print(formatLine("Max Refresh Times", fmt.Sprintf("%d", opt.MaxRefreshTimes)))
	fmt.Print(formatLine("Renew Interval", fmt.Sprintf("%d ms", opt.RenewInterval)))

	// Token settings | Token 配置
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Print(formatLine("Token Delimiter", opt.TokenDelimiter))
	fmt.Print(formatLine("Multi Login", fmt.Sprintf("%t", opt.MultiLogin)))
	fmt.Print(formatLine("Encrypt Key", maskKey(string(opt.EncryptKey))))

	// Pool settings | 协程池配置
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Print(formatLine("Pool Min Size", opt.PoolMinSize))
	fmt.Print(formatLine("Pool Max Size", opt.PoolMaxSize))
	fmt.Print(formatLine("Scale Up Rate", fmt.Sprintf("%.2f", opt.PoolScaleUpRate)))
	fmt.Print(formatLine("Scale Down Rate", fmt.Sprintf("%.2f", opt.PoolScaleDownRate)))

	// Auth excluded paths | 免认证路径
	if len(opt.AuthExcludePaths) > 0 {
		fmt.Println("├──────────────────────────────────────────────────────────────┤")
		for _, path := range opt.AuthExcludePaths {
			fmt.Print(formatLine("Auth Exclude Path", path))
		}
	}

	fmt.Println("└──────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

// maskKey hides encryption key for display | 对加密密钥进行脱敏展示
func maskKey(key string) string {
	if len(key) <= 6 {
		return "***"
	}
	return fmt.Sprintf("%s***%s", key[:3], key[len(key)-3:])
}
