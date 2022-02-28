package gftoken

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"strings"
)

func (m *GfToken) getRequestToken(r *ghttp.Request) (token string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			g.Log().Warning(r.GetCtx(), "authHeader:"+authHeader+" get token key fail")
			err = gerror.New("get token key fail")
			return
		} else if parts[1] == "" {
			g.Log().Warning(r.GetCtx(), "authHeader:"+authHeader+" get token fail")
			err = gerror.New(" get token fail")
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
	return
}

func (m *GfToken) GetUserInfo(r *ghttp.Request) *User {
	UserClaims, err := m.ParseToken(m.GetToken(r))
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "ParseToken error: %s\n", err.Error())
		return nil
	}
	return &UserClaims.User
}

func (m *GfToken) GetToken(r *ghttp.Request) string {
	token, err := m.getRequestToken(r)
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "ParseToken error: %s\n", err.Error())
		return ""
	}
	return token
}
