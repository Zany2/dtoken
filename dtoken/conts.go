package dtoken

const (
	GTokenCfgName = "gToken" // Global configuration node name for gToken | 全局配置文件中 gToken 节点名称

	CacheModeCache   = 1            // Cache mode using in-memory cache | 使用内存缓存的缓存模式
	CacheModeRedis   = 2            // Cache mode using Redis | 使用 Redis 的缓存模式
	CacheModeFile    = 3            // Cache mode using file | 使用文件存储的缓存模式
	CacheModeFileDat = "gtoken.dat" // Default file storage for cache | 缓存文件的默认存储名称

	DefaultTimeout        = 10 * 24 * 60 * 60 * 1000           // Default timeout (10 days in milliseconds) | 默认超时时间（10天，单位毫秒）
	DefaultCacheKey       = "GToken:"                          // Default prefix for cache keys | 默认缓存 key 前缀
	DefaultTokenDelimiter = "_"                                // Default delimiter for tokens | Token 的默认分隔符
	DefaultEncryptKey     = "12345678912345678912345678912345" // Default encryption key for token | 默认 Token 加密密钥

	// Cache key fields | 缓存 key 字段定义
	KeyUserKey       = "userKey"          // User identifier | 用户标识
	KeyCreateTime    = "createTime"       // Token creation time | 创建时间
	KeyRefreshNum    = "refreshNum"       // Token refresh count | 刷新次数
	KeyLastRenewTime = "keyLastRenewTime" // Last token renewal time | 上次续期时间
	KeyData          = "data"             // Custom data stored in cache | 缓存中的自定义数据
	KeyToken         = "token"            // The actual token value | 实际的 token 值
)

const (
	MsgErrUserKeyEmpty = "userKey empty"       // Error message when userKey is empty | 用户标识为空时的错误信息
	MsgErrTokenEmpty   = "token is empty"      // Error message when token is empty | Token 为空时的错误信息
	MsgErrTokenLen     = "token len error"     // Error message when token length is incorrect | Token 长度不正确时的错误信息
	MsgErrValidate     = "user validate error" // Error message for user validation failure | 用户验证失败时的错误信息
	MsgErrDataEmpty    = "cache value is nil"  // Error message when cache value is nil | 缓存值为空时的错误信息
)
