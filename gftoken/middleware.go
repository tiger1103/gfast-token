package gftoken

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"strings"
)

// Middleware 绑定group
func (m *GfToken) Middleware(group *ghttp.RouterGroup) error {
	group.Middleware(m.authMiddleware)
	return nil
}

func (m *GfToken) authMiddleware(r *ghttp.Request) {
	urlPath := r.URL.Path
	if !m.AuthPath(urlPath) {
		// 如果不需要认证，继续
		r.Middleware.Next()
		return
	}
	token, err := m.getRequestToken(r)
	if err != nil {
		_ = r.Response.WriteJson(g.Map{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}
	if m.IsEffective(r.GetCtx(), token) == false {
		_ = r.Response.WriteJson(g.Map{
			"code": 401,
			"msg":  "token已失效",
		})
		return
	}
	r.Middleware.Next()
}

// AuthPath 判断路径是否需要进行认证拦截
// return true 需要认证
func (m *GfToken) AuthPath(urlPath string) bool {
	// 去除后斜杠
	if strings.HasSuffix(urlPath, "/") {
		urlPath = gstr.SubStr(urlPath, 0, len(urlPath)-1)
	}
	// 排除路径处理，到这里nextFlag为true
	for _, excludePath := range m.ExcludePaths {
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
				return false
			}
		}
	}
	return true
}
