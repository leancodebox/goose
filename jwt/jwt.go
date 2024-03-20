package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserId uint64
	jwt.RegisteredClaims
}

var (
	TokenExpired     = errors.New("Token is expired ")
	TokenNotValidYet = errors.New("Token not active yet ")
	TokenMalformed   = errors.New("That's not even a token ")
	TokenInvalid     = errors.New("Couldn't handle this token: ")
)

type JWT struct {
	SigningKey []byte
}

func NewJWT(signingKey []byte) *JWT {
	return &JWT{
		signingKey,
	}
}

func CreateNewToken(userId uint64, expireTime time.Duration) (string, error) {
	cc := CustomClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now().Add(-10)),        // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)), // 过期时间 7天  配置文件
			Issuer:    "thh",                                          // 签名的发行者
		},
	}
	return Std().CreateToken(cc)
}

// CreateToken 创建一个token
func (itself *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(itself.SigningKey)
}

// ParseToken 解析 token
func (itself *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (i any, e error) {
			return itself.SigningKey, nil
		},
	)
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, TokenMalformed
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return nil, TokenExpired
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				return nil, TokenNotValidYet
			default:
				return nil, TokenInvalid
			}
		} else {
			return nil, err
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, TokenInvalid
}
