package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/leancodebox/goose/preferences"
	"github.com/spf13/cast"
	"sync"
	"time"
)

var (
	once       sync.Once
	std        *JWT
	signingKey = preferences.Get("jwt.signingKey", "mq+ZeGafL+b1xdC0u9vSVg==")
	validTime  = cast.ToDuration(preferences.GetInt64("jwt.validTime", 86400*7)) * time.Second
)

func Std() *JWT {
	once.Do(func() {
		std = NewJWT([]byte(signingKey))
	})
	return std
}

// VerifyTokenWithFresh 验证token 并刷新， 如果token还有1天就过期则生成新的token，否则还是用原来的
func VerifyTokenWithFresh(tokenStr string) (userId uint64, newToken string, err error) {
	claims, err := Std().ParseToken(tokenStr)
	if err != nil {
		return 0, "", err
	}

	if !claims.VerifyExpiresAt(time.Now().Add(time.Second*86400*1), true) {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(validTime))
		tokenStr, _ = Std().CreateToken(*claims)
	}
	return claims.UserId, tokenStr, err
}

func VerifyToken(tokenStr string) (userId uint64, err error) {
	claims, err := Std().ParseToken(tokenStr)
	if err != nil {
		return 0, err
	}
	return claims.UserId, err
}
