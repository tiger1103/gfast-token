package gftoken

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"strings"
)

const FailedAuthCode = 401

type AuthFailed struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *GfToken) getRequestToken(r *ghttp.Request) (token string) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return
		} else if parts[1] == "" {
			return
		}
		token = parts[1]
	} else {
		authHeader = r.Get("token").String()
		if authHeader == "" {
			return
		}
		token = authHeader
	}
	return
}

func (m *GfToken) GetToken(r *ghttp.Request) (tData *tokenData, err error) {
	token := m.getRequestToken(r)
	tData, _, err = m.getTokenData(r.GetCtx(), token)
	return
}

func (m *GfToken) IsLogin(r *ghttp.Request) (b bool, failed *AuthFailed) {
	b = true
	urlPath := r.URL.Path
	if !m.AuthPath(urlPath) {
		// 如果不需要认证，继续
		return
	}
	token := m.getRequestToken(r)
	if m.IsEffective(r.GetCtx(), token) == false {
		b = false
		failed = &AuthFailed{
			Code:    FailedAuthCode,
			Message: "token已失效",
		}
	}
	return
}
