package gftoken

import "github.com/gogf/gf/v2/net/ghttp"

func (m *GfToken) GetUserInfo(r *ghttp.Request) *User {
	result, ok := r.GetCtxVar(MY_CLAIMS).Interface().(*CustomClaims)
	if ok {
		return &result.User
	}
	return nil
}

func (m *GfToken) GetToken(r *ghttp.Request) string {
	return r.GetCtxVar(MY_TOKEN).String()
}
