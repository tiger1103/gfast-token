package gftoken

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"strings"
)

const FailedAuthCode = 401

type AuthFailed struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *GfToken) getRequestToken(r *ghttp.Request) (token string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			g.Log().Error(r.GetCtx(), "authHeader:"+authHeader+" get Token key fail")
			err = gerror.New("get Token key fail")
			return
		} else if parts[1] == "" {
			g.Log().Error(r.GetCtx(), "authHeader:"+authHeader+" get Token fail")
			err = gerror.New(" get Token fail")
			return
		}
		token = parts[1]
	} else {
		authHeader = r.Get("token").String()
		if authHeader == "" {
			g.Log().Error(r.GetCtx(), "params:"+authHeader+" get Token key fail")
			err = gerror.New("query Token fail")
			return
		}
		token = authHeader
	}
	return
}

func (m *GfToken) GetToken(r *ghttp.Request) (tData *tokenData, err error) {
	var token string
	token, err = m.getRequestToken(r)
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "ParseToken error: %s\n", err.Error())
		return
	}
	tData, _, err = m.getTokenData(r.GetCtx(), token)
	return
}

func (m *GfToken) IsLogin(r *ghttp.Request) (b bool, failed *AuthFailed) {
	urlPath := r.URL.Path
	if !m.AuthPath(urlPath) {
		// 如果不需要认证，继续
		b = true
		return
	}
	token, err := m.getRequestToken(r)
	if err != nil {
		b = false
		failed = &AuthFailed{
			Code:    FailedAuthCode,
			Message: err.Error(),
		}
		return
	}
	if m.IsEffective(r.GetCtx(), token) == false {
		b = false
		failed = &AuthFailed{
			Code:    FailedAuthCode,
			Message: "token已失效",
		}
	}
	return
}
