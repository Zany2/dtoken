package dtoken

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/panjf2000/ants/v2"
)

// Token defines token interface | Token 接口定义
type Token interface {
	Generate(ctx context.Context, userKey string, data any) (token string, err error) // Generate token | 生成 Token
	Validate(ctx context.Context, token string) (data any, err error)                 // Validate token | 验证 Token
	Get(ctx context.Context, userKey string) (token string, data any, err error)      // Get token by userKey | 通过 userKey 获取 Token
	ParseToken(ctx context.Context, token string) (userKey string, data any, err error)
	Destroy(ctx context.Context, userKey string) error // Destroy token | 销毁 Token
	Renew(ctx context.Context, token string)           // Asynchronously renew token | 异步续期 Token
	GetOptions() Options                               // Get config options | 获取配置参数
}

// GTokenV2 main implementation | gToken 主体结构体
type GTokenV2 struct {
	Options          Options
	Codec            Codec
	Cache            Cache
	RenewPoolManager *RenewPoolManager
}

// NewDefaultTokenByConfig creates a token from global config | 从全局配置创建 Token
func NewDefaultTokenByConfig() Token {
	var options *Options
	err := g.Cfg().MustGet(gctx.New(), "gToken").Struct(&options)
	if err != nil {
		panic("gToken options init failed")
	}
	if options == nil {
		panic("gToken options config not configured")
	}
	return NewDefaultToken(*options)
}

// NewDefaultToken creates token instance with options | 使用配置创建 Token 实例
func NewDefaultToken(options Options) Token {
	// Apply defaults | 应用默认配置
	if options.CacheMode == 0 {
		options.CacheMode = CacheModeCache
	}
	if options.CachePreKey == "" {
		options.CachePreKey = DefaultCacheKey
	}
	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout
		options.MaxRefresh = DefaultTimeout / 2
	}
	if len(options.EncryptKey) == 0 {
		options.EncryptKey = []byte(DefaultEncryptKey)
	}
	if options.TokenDelimiter == "" {
		options.TokenDelimiter = DefaultTokenDelimiter
	}
	if options.PoolMinSize <= 0 {
		options.PoolMinSize = DefaultMinSize
	}
	if options.PoolMaxSize <= 0 {
		options.PoolMaxSize = DefaultMaxSize
	}
	if options.PoolScaleUpRate <= 0 {
		options.PoolScaleUpRate = DefaultScaleUpRate
	}
	if options.PoolScaleDownRate <= 0 {
		options.PoolScaleDownRate = DefaultScaleDownRate
	}
	if options.RenewInterval <= 0 {
		options.RenewInterval = DefaultRenewInterval.Milliseconds() // 默认续期间隔（毫秒）
	}

	// Initialize renew pool | 初始化续期协程池
	renewPoolManager, err := NewRenewPoolBuilder().
		MinSize(options.PoolMinSize).
		MaxSize(options.PoolMaxSize).
		ScaleUpRate(options.PoolScaleUpRate).
		ScaleDownRate(options.PoolScaleDownRate).
		Build()
	if err != nil {
		panic(err)
	}

	gfToken := &GTokenV2{
		Options:          options,
		Codec:            NewDefaultCodec(options.TokenDelimiter, options.EncryptKey),
		Cache:            NewDefaultCache(options.CacheMode, options.CachePreKey, options.Timeout),
		RenewPoolManager: renewPoolManager,
	}

	g.Log().Infof(gctx.New(), gfToken.Options.String())
	return gfToken
}

// Generate creates a new token for user | 生成 Token
func (m *GTokenV2) Generate(ctx context.Context, userKey string, data any) (token string, err error) {
	if userKey == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	// Support multi-login (reuse existing token) | 支持多端重复登录（重用旧 Token）
	if m.Options.MultiLogin {
		token, _, err = m.Get(ctx, userKey)
		if err == nil && token != "" {
			return token, nil
		}
	}

	token, err = m.Codec.Encode(ctx, userKey)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}

	userCache := g.Map{
		KeyUserKey:       userKey,
		KeyToken:         token,
		KeyData:          data,
		KeyRefreshNum:    0,
		KeyLastRenewTime: 0,
		KeyCreateTime:    gtime.Now().TimestampMilli(),
	}

	if err = m.Cache.Set(ctx, userKey, userCache); err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return token, nil
}

// Validate checks token validity and optionally triggers renewal | 验证 Token 并触发续期
func (m *GTokenV2) Validate(ctx context.Context, token string) (data any, err error) {
	if token == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrTokenEmpty)
	}

	userKey, err := m.Codec.Decrypt(ctx, token)
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err)
	}

	userCache, err := m.Cache.Get(ctx, userKey)
	if err != nil {
		return nil, err
	}
	if userCache == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	if token != userCache[KeyToken] {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, MsgErrValidate)
	}

	// Renewal check | 检查是否需要续期
	now := gtime.Now().TimestampMilli()
	create := gconv.Int64(userCache[KeyCreateTime])
	refreshNum := gconv.Int(userCache[KeyRefreshNum])
	lastRenew := gconv.Int64(userCache[KeyLastRenewTime])

	// Prevent renewal spam | 防止重复续期
	if lastRenew > 0 && now-lastRenew < m.Options.RenewInterval {
		return userCache[KeyData], nil
	}

	if m.Options.MaxRefresh > 0 &&
		now > create+m.Options.MaxRefresh &&
		(m.Options.MaxRefreshTimes == 0 || refreshNum < m.Options.MaxRefreshTimes) {
		m.Renew(ctx, token)
	}

	return userCache[KeyData], nil
}

// Get retrieves token and data by userKey | 通过 userKey 获取 Token
func (m *GTokenV2) Get(ctx context.Context, userKey string) (token string, data any, err error) {
	if userKey == "" {
		return "", nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	userCache, err := m.Cache.Get(ctx, userKey)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if userCache == nil {
		return "", nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return gconv.String(userCache[KeyToken]), userCache[KeyData], nil
}

// ParseToken parses token to retrieve userKey and data | 解析 Token 获取 userKey 和数据
func (m *GTokenV2) ParseToken(ctx context.Context, token string) (userKey string, data any, err error) {
	if token == "" {
		return "", nil, gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	userKey, err = m.Codec.Decrypt(ctx, token)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInvalidParameter, err)
	}

	userCache, err := m.Cache.Get(ctx, userKey)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if userCache == nil {
		return "", nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return userKey, userCache[KeyData], nil
}

// Destroy removes user token from cache | 销毁 Token
func (m *GTokenV2) Destroy(ctx context.Context, userKey string) error {
	if userKey == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}
	if err := m.Cache.Remove(ctx, userKey); err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return nil
}

// Renew asynchronously renews a token | 异步续期 Token
func (m *GTokenV2) Renew(ctx context.Context, token string) {
	if err := m.RenewPoolManager.Submit(func() {

		// 1. Decode token to extract userKey | 解析 Token 获取用户标识
		userKey, err := m.Codec.Decrypt(ctx, token)
		if err != nil {
			g.Log().Errorf(ctx, "gToken: Token renewal decode failed token:%s err:%s", token, err.Error())
			return
		}

		// 2. Retrieve user cache from storage | 获取用户缓存数据
		userCache, err := m.Cache.Get(ctx, userKey)
		if err != nil || userCache == nil {
			return
		}

		// 3. Read current time and renewal-related metadata | 读取当前时间与续期相关信息
		nowTime := gtime.Now().TimestampMilli()
		createTime := gconv.Int64(userCache[KeyCreateTime])       // Token creation time | 创建时间
		refreshNum := gconv.Int(userCache[KeyRefreshNum])         // Number of times token has been refreshed | 已续期次数
		lastRenewTime := gconv.Int64(userCache[KeyLastRenewTime]) // Last token renewal time | 上次续期时间

		// 4. Limit renewal frequency by configuration | 按配置限制续期间隔
		if lastRenewTime > 0 && nowTime-lastRenewTime < m.Options.RenewInterval {
			g.Log().Debugf(ctx, "gToken: User %s renewal too frequent (interval < %dms)", userKey, m.Options.RenewInterval)
			return
		}

		// 5. Check renewal policy | 检查续期策略
		if m.Options.MaxRefresh == 0 ||
			(m.Options.MaxRefreshTimes > 0 && refreshNum >= m.Options.MaxRefreshTimes) {
			return
		}

		// 6. Check if token is due for renewal | 检查是否到达续期时间
		if nowTime > createTime+m.Options.MaxRefresh {

			// Update renewal info | 更新续期信息
			userCache[KeyRefreshNum] = refreshNum + 1
			userCache[KeyLastRenewTime] = nowTime

			// Write back to cache | 写回缓存
			if err = m.Cache.Set(ctx, userKey, userCache); err != nil {
				g.Log().Errorf(ctx, "gToken: Token renewal cache write failed userKey:%s err:%s", userKey, err.Error())
			}
		}

	}); err != nil {

		// 7. Handle pool overload or submission errors | 协程池已满或任务提交失败
		if gerror.Is(err, ants.ErrPoolOverload) {
			g.Log().Warningf(ctx, "gToken: Token renewal task dropped (pool full) token:%s", token)
			return
		}

		g.Log().Errorf(ctx, "gToken: Token renewal task submission failed token:%s err:%s", token, err.Error())
	}
}

// GetOptions 获取Options配置
func (m *GTokenV2) GetOptions() Options {
	return m.Options
}

// Shutdown gracefully stops renew pool | 优雅关闭续期协程池
func (m *GTokenV2) Shutdown() {
	if m.RenewPoolManager != nil {
		m.RenewPoolManager.Stop()
	}
}
