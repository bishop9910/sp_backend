package middleware

import (
	"log"
	"net/http"
	"strings"

	"sp_backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtConfig *utils.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// 提取 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtConfig.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// OptionalAuth 可选认证：有 Token 则解析，没有则放行
func OptionalAuth(jwtConfig *utils.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := jwtConfig.ParseToken(parts[1])
		if err != nil {
			log.Default().Printf("optional auth failed: %v", err)
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// RequiredAuth 强制认证：必须携带有效 Token
func RequiredAuth(jwtConfig *utils.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "未提供认证令牌",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "令牌格式错误，应为 'Bearer <token>'",
				"code":    "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		claims, err := jwtConfig.ParseToken(parts[1])
		if err != nil {
			log.Default().Printf("auth failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "令牌无效或已过期",
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
