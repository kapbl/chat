package handles

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JwtCustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func JWTTokenGenerate(secret string, userID string) string {
	claims := JwtCustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
			Issuer:    "my_app",
		},
	}
	// 创建Token，使用HS256算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名并获取完整Token字符串
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("生成Token失败:", err)
		return ""
	}
	// fmt.Println("JWT Token:", tokenString)
	return tokenString
}

func JWTValidator(secretKey []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(401, "缺少Authorization头")
			}
			tokenString := authHeader[len("Bearer "):]
			if tokenString == "" {
				return echo.NewHTTPError(401, "无效的Token格式")
			}

			// 解析并验证Token
			token, err := jwt.ParseWithClaims(
				tokenString,
				&JwtCustomClaims{},
				func(token *jwt.Token) (interface{}, error) {
					// 验证签名算法
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
					}
					return secretKey, nil
				},
			)
			// 处理解析错误
			if err != nil {
				return echo.NewHTTPError(401, "Token验证失败: "+err.Error())
			}
			// 验证Claims有效性
			if claims, ok := token.Claims.(*JwtCustomClaims); ok && token.Valid {
				// 将Claims存入Echo上下文，供后续使用
				c.Set("jwt_claims", claims)
				return next(c)
			} else {
				return echo.NewHTTPError(401, "无效的Token")
			}
		}
	}
}

func JWTUnencoder(secretKey []byte, tokenString string) *JwtCustomClaims {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JwtCustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
			}
			return secretKey, nil
		},
	)
	// 处理解析错误
	if err != nil {
		return nil
	}
	// 验证Claims有效性
	if claims, ok := token.Claims.(*JwtCustomClaims); ok && token.Valid {
		return claims
	}
	return nil
}
