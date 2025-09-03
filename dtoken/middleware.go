package gtoken

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"strings"
)

type Middleware struct {
	// Token
	Token Token
	// 自定义Token校验失败返回方法
	ResFun func(r *ghttp.Request)
}

func NewDefaultMiddleware(token Token, resFun ...func(r *ghttp.Request)) Middleware {
	if len(resFun) > 0 {
		return Middleware{
			Token:  token,
			ResFun: resFun[0],
		}
	}

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

// Auth 认证拦截
// 认证失败统一错误码：gcode.CodeBusinessValidationFailed
func (m Middleware) Auth(r *ghttp.Request) {
	if m.HasExcludePath(r) {
		// 如果不需要认证，继续
		r.Middleware.Next()
		return
	}

	// 获取请求token
	token, err := GetRequestToken(r)
	if err != nil {
		m.ResFun(r)
		return
	}

	// 校验token
	userCacheValue, err := m.Token.Validate(r.Context(), token)
	if err != nil {
		m.ResFun(r)
		return
	}

	r.SetCtxVar(KeyUserKey, userCacheValue)
	r.Middleware.Next()
}

// HasExcludePath 判断路径是否需要进行认证拦截过滤
// @return true 不需要认证
func (m Middleware) HasExcludePath(r *ghttp.Request) bool {
	var (
		urlPath      = r.URL.Path
		excludePaths = m.Token.GetOptions().AuthExcludePaths
	)
	if len(excludePaths) == 0 {
		return false
	}
	// 去除后斜杠
	if strings.HasSuffix(urlPath, "/") {
		urlPath = gstr.SubStr(urlPath, 0, len(urlPath)-1)
	}

	// 排除路径处理，到这里nextFlag为true
	for _, excludePath := range excludePaths {
		tmpPath := excludePath
		// 前缀匹配
		if strings.HasSuffix(tmpPath, "/*") {
			tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-2)
			if gstr.HasPrefix(urlPath, tmpPath) {
				// 前缀匹配不拦截
				return false
			}
		} else {
			// 全路径匹配
			if strings.HasSuffix(tmpPath, "/") {
				tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-1)
			}
			if urlPath == tmpPath {
				// 全路径匹配不拦截
				return true
			}
		}
	}

	return false
}

// GetRequestToken 返回请求Token
func GetRequestToken(r *ghttp.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param invalid")
		} else if parts[1] == "" {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "Bearer param empty")
		}

		return parts[1], nil
	}

	authHeader = r.Get(KeyToken).String()
	if authHeader == "" {
		return "", gerror.NewCode(gcode.CodeMissingParameter, "token empty")
	}
	return authHeader, nil
}
