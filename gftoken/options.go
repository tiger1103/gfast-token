package gftoken

import (
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/os/gcache"
)

var (
	defaultGFToken = GfToken{
		ServerName: "defaultGFToken",
		CacheKey:   "defaultGFToken_",
		Timeout:    60 * 60 * 24 * 10,
		MaxRefresh: 60 * 60 * 24 * 5,
		cache:      gcache.New(),
		userJwt:    CreateMyJWT("defaultGFToken"),
		MultiLogin: 0,
	}
)

type OptionFunc func(*GfToken)

func NewGfToken(opts ...OptionFunc) *GfToken {
	g := defaultGFToken
	for _, o := range opts {
		o(&g)
	}
	return &g
}

func WithServerName(value string) OptionFunc {
	return func(g *GfToken) {
		g.ServerName = value
	}
}

func WithCacheKey(value string) OptionFunc {
	return func(g *GfToken) {
		g.CacheKey = value
	}
}

func WithTimeoutAndMaxRefresh(timeout, maxRefresh int64) OptionFunc {
	return func(g *GfToken) {
		g.Timeout = timeout
		g.MaxRefresh = maxRefresh
	}
}

func WithTimeout(value int64) OptionFunc {
	return func(g *GfToken) {
		g.Timeout = value
	}
}

func WithMaxRefresh(value int64) OptionFunc {
	return func(g *GfToken) {
		g.MaxRefresh = value
	}
}

func WithUserJwt(key string) OptionFunc {
	return func(g *GfToken) {
		g.userJwt = CreateMyJWT(key)
	}
}

func WithGCache() OptionFunc {
	return func(g *GfToken) {
		g.cache = gcache.New()
	}
}

func WithGRedis(redisConfig ...*gredis.Config) OptionFunc {
	return func(g *GfToken) {
		g.cache = gcache.New()
		redis, err := gredis.New(redisConfig...)
		if err != nil {
			panic(err)
		}
		g.cache.SetAdapter(gcache.NewAdapterRedis(redis))
	}
}

func WithMultiLogin(num int) OptionFunc {
	return func(g *GfToken) {
		g.MultiLogin = num
	}
}
