package gftoken

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	FailedAuthCode = 401
	BearerPrefix   = "Bearer "
)

type AuthFailed struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}


func (m *GfToken) GetRequestToken(r *ghttp.Request) (token string) {
	// 请求头获取
	n := len(BearerPrefix)
	auth := r.Header.Get("Authorization")
	if len(auth) >= n && auth[:n] == BearerPrefix {
		return auth[n:]
	}
	// 查询参数
	if q := r.Get("token"); !q.IsEmpty() {
		return q.String()
	}
	// Cookies
	if c := r.Cookie.Get("token"); !c.IsEmpty() {
		return c.String()
	}
	return
}

func (m *GfToken) GetToken(r *ghttp.Request) (tData *TokenData, err error) {
	token := m.GetRequestToken(r)
	tData, _, err = m.GetTokenData(r.GetCtx(), token)
	return
}

func (m *GfToken) IsLogin(r *ghttp.Request) (b bool, failed *AuthFailed) {
	b = true
	urlPath := r.URL.Path
	if !m.AuthPath(urlPath) {
		// 如果不需要认证，继续
		return
	}
	token := m.GetRequestToken(r)
	if m.IsEffective(r.GetCtx(), token) == false {
		b = false
		failed = &AuthFailed{
			Code:    FailedAuthCode,
			Message: "token已失效",
		}
	}
	return
}
