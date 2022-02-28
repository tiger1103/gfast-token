/*
* @desc:功能测试
* @company:云南奇讯科技有限公司
* @Author: yixiaohu
* @Date:   2022/2/28 10:09
 */

package test

import (
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/wilgx0/gftoken/gftoken"
	"testing"
)

func TestToken(t *testing.T) {
	t.Run("newToken", newToken)
}

func newToken(t *testing.T) {

}

func test(t *testing.T) {
	/**
	注意事项:
	1、token存活时间 = 超时时间 + 缓存刷新时间
	2、处理携带token的请求时当前时间大于超时时间并小于缓存刷新时间时token将自动刷新即重置token存活时间
	3、每创建一个gftoken实例时CacheKey必须不相同
	4、GenerateToken函数参数的User.UserKey为用户唯一标识，必须且唯一
	*/
	gft := gftoken.NewGfToken(
		gftoken.WithCacheKey("potato_"),
		gftoken.WithMultiLogin(2),
		gftoken.WithGRedis(&gredis.Config{
			Address: "127.0.0.1:6379",
			Db:      9,
		}))
	s := g.Server()
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.GET("/login", func(r *ghttp.Request) {
			userId := r.GetQuery("id").String()
			token, err := gft.GenerateToken(r.GetCtx(), gftoken.User{
				UserKey: userId,
				Data:    "myData",
			})

			if err != nil {
				g.Log().Error(r.GetCtx(), err)
			}

			r.Response.Write(token)
		})

		gft.Middleware(group)
		group.GET("/user", func(r *ghttp.Request) {
			user := gft.GetUserInfo(r)
			r.Response.Write(user)
		})
	})
	s.SetPort(8080)
	s.Run()
}
