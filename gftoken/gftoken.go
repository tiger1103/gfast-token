package gftoken

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/golang-jwt/jwt"
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
	// 是否允许多点登录
	MultiLogin bool
	// Token加密key 32位
	EncryptKey [32]byte
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
func (m *GfToken) GenerateToken(ctx context.Context, key string, data interface{}) (keys string, err error) {
	// 支持多点登录
	if m.MultiLogin {

	}
	tokens, err := m.userJwt.CreateToken(CustomClaims{
		data,
		jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 10, // 生效开始时间
			ExpiresAt: m.diedLine().Unix(),    // 失效截止时间
		},
	})
	if err != nil {
		return
	}
	keys, err = m.EncryptToken(ctx, key)
	if err != nil {
		return
	}
	err = m.setCache(ctx, key, tokens)
	if err != nil {
		return
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
	_, code := m.IsNotExpired(cacheToken)
	if JwtTokenOK == code {
		// 刷新缓存
		if m.IsRefresh(cacheToken) {
			if newToken, err := m.RefreshToken(cacheToken); err == nil {
				err = m.setCache(ctx, m.CacheKey+token, newToken)
				if err != nil {
					g.Log().Error(ctx, err)
				}
			}
		}
		return true
	}
	return false
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

// EncryptToken token加密方法
func (m *GfToken) EncryptToken(ctx context.Context, key string) (encryptStr string, err error) {
	if key == "" {
		err = gerror.New("encrypt key empty")
		return
	}
	ek := m.EncryptKey[:]
	token, err := gaes.Encrypt([]byte(key), ek)
	if err != nil {
		g.Log().Error(ctx, "[GFToken]encrypt error token:", key, err)
		err = gerror.New("encrypt error")
		return
	}
	encryptStr = gbase64.EncodeToString(token)
	return
}

// DecryptToken token解密方法
func (m *GfToken) DecryptToken(ctx context.Context, token string) (DecryptStr string, err error) {
	if token == "" {
		err = gerror.New("decrypt token empty")
		return
	}
	token64, err := gbase64.Decode([]byte(token))
	if err != nil {
		g.Log().Error(ctx, "[GFToken]decode error token:", token, err)
		err = gerror.New("decode error")
		return
	}
	ek := m.EncryptKey[:]
	decryptToken, err := gaes.Decrypt(token64, ek)
	if err != nil {
		g.Log().Error(ctx, "[GFToken]decrypt error token:", token, err)
		err = gerror.New("decrypt error")
		return
	}
	DecryptStr = string(decryptToken)
	return
}
