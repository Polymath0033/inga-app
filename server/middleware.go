package server

import (
	"fmt"
	"net/http"
	"strings"

	"example.com/inga/internal/api"
	"example.com/inga/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims for authenticated users
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token from Authorization header
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Abort()
			ctx.JSON(http.StatusUnauthorized, api.ResponseEnvelope{
				Code:    http.StatusUnauthorized,
				Message: "missing authorization header",
				Data:    nil,
				Error:   "missing authorization header",
			})
			return
		}

		// Extract bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.Abort()
			ctx.JSON(http.StatusUnauthorized, api.ResponseEnvelope{
				Code:    http.StatusUnauthorized,
				Message: "invalid authorization header format",
				Data:    nil,
				Error:   "invalid authorization header format",
			})
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			ctx.Abort()
			ctx.JSON(http.StatusUnauthorized, api.ResponseEnvelope{
				Code:    http.StatusUnauthorized,
				Message: "invalid or expired token",
				Data:    nil,
				Error:   "invalid or expired token",
			})
			return
		}

		// Store claims in context for use in handlers
		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)
		ctx.Set("claims", claims)

		ctx.Next()
	}
}

// RecoveryMiddleware handles panics and converts them to proper error responses
func RecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.JSON(http.StatusInternalServerError, api.ResponseEnvelope{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
					Data:    nil,
					Error:   "unexpected error occurred",
				})
				ctx.Abort()
			}
		}()
		ctx.Next()
	}
}

// LoggingMiddleware logs request details
func LoggingMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// You can add structured logging here using your logger
		// For now, this is a placeholder that you can expand with your logger
		_ = cfg // placeholder for future use
		ctx.Next()
	}
}

// CORSMiddleware handles cross-origin requests
func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
