package gftoken

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"time"
)

type GfToken struct {
	//  server name
	ServerName string
	// 缓存key (每创建一个实例CacheKey必须不相同)
	CacheKey string
	// 超时时间 默认10天（秒）
	Timeout int64
	// 缓存刷新时间 默认5天（秒）
	// 处理携带token的请求时当前时间大于超时时间并小于缓存刷新时间时token将自动刷新即重置token存活时间
	// MaxRefresh值为0时,token将不会自动刷新
	MaxRefresh int64
	// 多点登录时总共允许登录的token个数，默认0即不限制
	MultiLogin int
	// 缓存 (缓存模式:gcache 或 gredis)
	cache *gcache.Cache
	// jwt
	userJwt *JwtSign
}

// 存活时间 (存活时间 = 超时时间 + 缓存刷新时间)
func (m *GfToken) diedLine() time.Time {
	return time.Now().Add(time.Second * time.Duration(m.Timeout+m.MaxRefresh))
}

// 生成token
func (m *GfToken) GenerateToken(ctx context.Context, user User) (tokens string, err error) {
	tokens, err = m.userJwt.CreateToken(CustomClaims{
		user,
		jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 10, // 生效开始时间
			ExpiresAt: m.diedLine().Unix(),    // 失效截止时间
		},
	})
	if err != nil {
		return
	}
	err = m.setCache(ctx, m.CacheKey+tokens, tokens)
	if err != nil {
		return
	}

	if m.MultiLogin > 0 {
		err = m.addUserKeyCache(ctx, m.CacheKey+user.UserKey, m.CacheKey+tokens)
		if err != nil {
			return
		}
	}
	return
}

// 解析token (只验证格式并不验证过期)
func (m *GfToken) ParseToken(tokenStr string) (*CustomClaims, error) {
	if customClaims, err := m.userJwt.ParseToken(tokenStr); err == nil {
		return customClaims, nil
	} else {
		return &CustomClaims{}, errors.New(ErrorsParseTokenFail)
	}
}

// 检查缓存的token是否有效且自动刷新缓存token
func (m *GfToken) IsEffective(ctx context.Context, token string) bool {
	cacheToken, err := m.getCache(ctx, m.CacheKey+token)
	if err != nil {
		g.Log().Error(ctx, err)
		return false
	}
	if cacheToken == "" {
		return false
	}

	if m.MultiLogin > 0 && m.checkMultiLogin(ctx, token) == false {
		return false
	}

	claims, code := m.IsNotExpired(cacheToken)
	if JwtTokenOK == code {
		// 刷新缓存
		if m.IsRefresh(cacheToken) {
			if newToken, err := m.RefreshToken(cacheToken); err == nil {
				err = m.setCache(ctx, m.CacheKey+token, newToken)
				if err != nil {
					g.Log().Error(ctx, err)
				}
				err = m.UpdateExpireUserKeyCache(ctx, m.CacheKey+claims.User.UserKey)
				if err != nil {
					g.Log().Error(ctx, err)
				}
				// g.Log().Print(ctx, "token 已刷新!")
			}
			if err != nil {
				g.Log().Error(ctx, err)
			}
		}
		return true
	}
	return false
}

// 检查多点登录
func (m *GfToken) checkMultiLogin(ctx context.Context, token string) bool {

	claims, err := m.ParseToken(token)
	if err != nil {
		g.Log().Error(ctx, err)
		return false
	}
	tokenList, err := m.getUserKeyCache(ctx, m.CacheKey+claims.User.UserKey)
	if err != nil {
		g.Log().Error(ctx, err)
		return false
	}
	j := m.MultiLogin
	flag := false
	for i := len(tokenList) - 1; i >= 0; i-- {
		if j <= 0 {
			break
		}
		if tokenList[i] == m.CacheKey+token {
			flag = true
			break
		}
		j--
	}

	return flag
}

// 检查token是否过期 (过期时间 = 超时时间 + 缓存刷新时间)
func (m *GfToken) IsNotExpired(token string) (*CustomClaims, int) {
	if customClaims, err := m.userJwt.ParseToken(token); err == nil {
		if time.Now().Unix()-customClaims.ExpiresAt < 0 {
			// token有效
			return customClaims, JwtTokenOK
		} else {
			// 过期的token
			return customClaims, JwtTokenExpired
		}
	} else {
		// 无效的token
		return customClaims, JwtTokenInvalid
	}
}

// 刷新token的缓存有效期
func (m *GfToken) RefreshToken(oldToken string) (newToken string, err error) {
	if newToken, err = m.userJwt.RefreshToken(oldToken, m.diedLine().Unix()); err != nil {
		return
	}
	return
}

// token是否处于刷新期
func (m *GfToken) IsRefresh(token string) bool {
	if m.MaxRefresh == 0 {
		return false
	}
	if customClaims, err := m.userJwt.ParseToken(token); err == nil {
		now := time.Now().Unix()
		if now < customClaims.ExpiresAt && now > (customClaims.ExpiresAt-m.MaxRefresh) {
			return true
		}
	}
	return false
}
