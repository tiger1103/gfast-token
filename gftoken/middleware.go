package gftoken

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Middleware 绑定group
func (m *GfToken) Middleware(group *ghttp.RouterGroup) error {
	group.Middleware(m.authMiddleware)
	return nil
}

func (m *GfToken) authMiddleware(r *ghttp.Request) {
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
