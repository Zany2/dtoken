package dtoken

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"strings"
)

// Middleware defines the authentication middleware | 认证中间件结构体
type Middleware struct {
	Token  Token                  // Token instance | Token 实例
	ResFun func(r *ghttp.Request) // Custom response for validation failure | 自定义 Token 校验失败响应方法
}

// NewDefaultMiddleware creates a middleware instance | 创建默认中间件实例
// If resFun is provided, it will be used as the custom response handler | 如果传入 resFun，将使用自定义响应函数
func NewDefaultMiddleware(token Token, resFun ...func(r *ghttp.Request)) Middleware {
	if len(resFun) > 0 {
		return Middleware{
			Token:  token,
			ResFun: resFun[0],
		}
	}

	// Default error response when validation fails | 默认 Token 校验失败响应
	return Middleware{
		Token: token,
		ResFun: func(r *ghttp.Request) {
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    gcode.CodeInternalError.Code(),
				Message: gcode.CodeInternalError.Message(),
				Data:    []interface{}{},
			})
		},
	}
}

// Auth performs token authentication for requests | 执行请求认证拦截
// If validation fails, returns a unified error response | 校验失败返回统一错误响应
// Error Code: gcode.CodeBusinessValidationFailed | 错误码：gcode.CodeBusinessValidationFailed
func (m Middleware) Auth(r *ghttp.Request) {
	// Skip authentication if path is excluded | 路径在排除列表中则跳过认证
	if m.HasExcludePath(r) {
		r.Middleware.Next()
		return
	}

	// Extract token from request | 从请求中获取 Token
	token, err := GetRequestToken(r)
	if err != nil {
		m.ResFun(r)
		return
	}

	// Validate token | 校验 Token 合法性
	userCacheValue, err := m.Token.Validate(r.Context(), token)
	if err != nil {
		m.ResFun(r)
		return
	}

	// Store user info in request context | 将用户数据存入请求上下文
	r.SetCtxVar(KeyUserKey, userCacheValue)

	// Continue request | 执行后续中间件链
	r.Middleware.Next()
}

// HasExcludePath determines if the current request path should bypass authentication | 判断路径是否应跳过认证
// @return true: skip authentication | true 表示不需要认证
func (m Middleware) HasExcludePath(r *ghttp.Request) bool {
	var (
		urlPath      = r.URL.Path
		excludePaths = m.Token.GetOptions().AuthExcludePaths
	)

	// No exclusion rules configured | 未配置排除路径
	if len(excludePaths) == 0 {
		return false
	}

	// Remove trailing slash | 去除路径末尾斜杠
	if strings.HasSuffix(urlPath, "/") {
		urlPath = gstr.SubStr(urlPath, 0, len(urlPath)-1)
	}

	// Iterate through exclude paths | 遍历排除路径规则
	for _, excludePath := range excludePaths {
		tmpPath := excludePath

		// Prefix match: e.g., "/api/*" | 前缀匹配（如 /api/*）
		if strings.HasSuffix(tmpPath, "/*") {
			tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-2)
			if gstr.HasPrefix(urlPath, tmpPath) {
				// Path matches prefix -> skip authentication | 匹配前缀路径则跳过认证
				return true
			}
		} else {
			// Full path match | 全路径匹配
			if strings.HasSuffix(tmpPath, "/") {
				tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-1)
			}
			if urlPath == tmpPath {
				// Exact match -> skip authentication | 精确匹配则跳过认证
				return true
			}
		}
	}

	// No exclusion match -> require authentication | 未匹配排除规则则需认证
	return false
}

// GetRequestToken extracts token from HTTP request | 从 HTTP 请求中提取 Token
// Supported methods: Header("Authorization: Bearer <token>") or param "token" | 支持 Header 方式和参数方式
func GetRequestToken(r *ghttp.Request) (string, error) {
	// 1. Try Authorization header | 优先从 Authorization 头中获取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)

		// Validate Bearer token format | 校验 Bearer 格式是否正确
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param invalid | Bearer 参数格式错误")
		} else if parts[1] == "" {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param empty | Bearer 参数为空")
		}

		return parts[1], nil
	}

	// 2. Fallback to token parameter | 尝试从请求参数中读取 Token
	authHeader = r.Get(KeyToken).String()
	if authHeader == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, "token empty | 缺少 token 参数")
	}

	return authHeader, nil
}
