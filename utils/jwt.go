package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config JWT 配置
type JWTConfig struct {
	SecretKey       string
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration
	Issuer          string
}

// CustomClaims 自定义 Claims，扩展用户信息
type CustomClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewJWTConfig 创建默认配置
func NewJWTConfig(secret string) *JWTConfig {
	return &JWTConfig{
		SecretKey:       secret,
		AccessTokenExp:  2 * time.Hour,      // 访问令牌 2 小时
		RefreshTokenExp: 7 * 24 * time.Hour, // 刷新令牌 7 天
		Issuer:          "community-app",
	}
}

// GenerateToken 生成访问令牌 + 刷新令牌
func (c *JWTConfig) GenerateToken(userID uint64, username string) (accessToken, refreshToken string, err error) {
	// 生成访问令牌
	accessToken, err = c.generateTokenWithExp(userID, username, c.AccessTokenExp)
	if err != nil {
		return "", "", err
	}

	// 生成刷新令牌
	refreshToken, err = c.generateTokenWithExp(userID, username, c.RefreshTokenExp)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateTokenWithExp 内部方法：生成指定过期时间的 Token
func (c *JWTConfig) generateTokenWithExp(userID uint64, username string, exp time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    c.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.SecretKey))
}

// ParseToken 解析并验证 Token
func (c *JWTConfig) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.SecretKey), nil
	})

	if err != nil {
		// 区分错误类型
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("token not active yet")
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("token malformed")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken 使用刷新令牌生成新的访问令牌
func (c *JWTConfig) RefreshToken(refreshToken string) (newAccessToken string, err error) {
	claims, err := c.ParseToken(refreshToken)
	if err != nil {
		return "", err
	}

	return c.generateTokenWithExp(claims.UserID, claims.Username, c.AccessTokenExp)
}
