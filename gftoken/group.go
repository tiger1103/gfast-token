package gftoken

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"strings"
)

// Middleware 绑定group
func (m *GfToken) Middleware(group *ghttp.RouterGroup) error {
	group.Middleware(m.authMiddleware)
	return nil
}

func (m *GfToken) authMiddleware(r *ghttp.Request) {
	var token string
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			g.Log().Warning(r.GetCtx(), "authHeader:"+authHeader+" get token key fail")
			_ = r.Response.WriteJson(g.Map{
				"code": 401,
				"msg":  "get token key fail",
			})
			return
		} else if parts[1] == "" {
			g.Log().Warning(r.GetCtx(), "authHeader:"+authHeader+" get token fail")
			_ = r.Response.WriteJson(g.Map{
				"code": 401,
				"msg":  " get token fail",
			})
			return
		}

		token = parts[1]
	} else {
		authHeader = r.GetQuery("token").String()
		if authHeader == "" {
			_ = r.Response.WriteJson(g.Map{
				"code": 401,
				"msg":  "query token fail",
			})
			return
		}
		token = authHeader
	}
	if m.IsEffective(r.GetCtx(), token) == false {
		g.Log().Error(r.GetCtx(), "token error: token已失效!")
		_ = r.Response.WriteJson(g.Map{
			"code": 401,
			"msg":  "token已失效",
		})
		return
	}

	UserClaims, err := m.ParseToken(token)
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "ParseToken error: %s\n", err.Error())
		_ = r.Response.WriteJson(g.Map{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}
	r.SetCtxVar(MY_CLAIMS, UserClaims)
	r.SetCtxVar(MY_TOKEN, token)

	r.Middleware.Next()
}
