package gftoken

import "github.com/golang-jwt/jwt"

const (
	//token部分
	ErrorsParseTokenFail    string = "解析token失败"
	ErrorsTokenInvalid      string = "无效的token"
	ErrorsTokenNotActiveYet string = "token 尚未激活"
	ErrorsTokenMalFormed    string = "token 格式不正确"

	JwtTokenOK            int = 200100  //token有效
	JwtTokenInvalid       int = -400100 //无效的token
	JwtTokenExpired       int = -400101 //过期的token
	JwtTokenFormatErrCode int = -400102 //提交的 token 格式错误

	MY_TOKEN  = "my_token"
	MY_CLAIMS = "my_claims"
)

type CustomClaims struct {
	User User
	jwt.StandardClaims
}

type User struct {
	UserKey string      // 用户唯一标识，必须且唯一
	Data    interface{} // 其他需要携带的数据
}
