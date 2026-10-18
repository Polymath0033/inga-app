package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"example.com/inga/internal/api"
	"example.com/inga/internal/config"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

// NewAuth0Middleware creates an Auth0 JWT validator middleware for Gin
// TEACHING GUIDE:
//
// What is Auth0 JWT middleware?
// Auth0 provides a reusable JWT validation library that:
// 1. Parses and validates tokens
// 2. Verifies signatures using keys from KeyFunc
// 3. Checks expiry and standard claims
// 4. Extracts claims into your custom struct
//
// Why use it over hand-rolled validation?
// - Better error handling and edge case coverage
// - Standard implementation used by Auth0 itself
// - Supports key rotation (for RS256/asymmetric keys)
// - Cleaner API
//
// How does it work?
// 1. validator.New() creates a validator with your config
// 2. You provide a KeyFunc that returns the signing key
// 3. The validator returns a *jwt.Token with your claims unmarshalled
// 4. You extract claims and store in context
func NewAuth0Middleware(cfg *config.Config) gin.HandlerFunc {
	// KeyFunc is called during validation to get the signing key
	// For HS256 (symmetric): always return your secret
	// For RS256 (asymmetric): fetch from a key endpoint
	keyFunc := func(ctx context.Context) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	}

	// Create the validator
	// Parameters:
	// - keyFunc: function to retrieve signing keys
	// - algorithm: which signing algorithm to accept (HS256, RS256, etc)
	// - audience: expected "aud" claim (can be empty string to skip validation)
	// - issuer: expected "iss" claim as a slice (can be nil to skip validation)
	issuer := []string{}
	if cfg.JWT.Issuer != "" {
		issuer = append(issuer, cfg.JWT.Issuer)
	}
	jwtValidator, err := validator.New(
		keyFunc,
		validator.HS256,
		cfg.JWT.Audience,
		issuer,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to setup JWT validator: %v", err))
	}

	// Return the middleware function
	return func(ginCtx *gin.Context) {
		// Step 1: Get the Authorization header
		authHeader := ginCtx.GetHeader("Authorization")
		if authHeader == "" {
			ginCtx.JSON(http.StatusUnauthorized, api.ResponseEnvelope{
				Code:    http.StatusUnauthorized,
				Message: "missing authorization header",
				Data:    nil,
				Error:   "missing authorization header",
			})
			ginCtx.Abort()
			return
		}

		// Step 2: Extract the token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ginCtx.JSON(http.StatusUnauthorized, api.ResponseEnvelope{
				Code:    http.StatusUnauthorized,
				Message: "invalid authorization header format",
				Data:    nil,
				Error:   "invalid authorization header format",
			})
			ginCtx.Abort()
			return
		}

		tokenString := parts[1]

		// Step 3: Validate the token using Auth0's validator
		// This:
		// - Decodes the JWT
		// - Verifies the signature using keyFunc
		// - Checks expiry
		// - Returns claims as map[string]interface{}
		claims, err := jwtValidator.ValidateToken(ginCtx.Request.Context(), tokenString)
		if err != nil {
			ginCtx.JSON(http.StatusUnauthorized, api.ResponseEnvelope{
				Code:    http.StatusUnauthorized,
				Message: "invalid or expired token",
				Data:    nil,
				Error:   fmt.Sprintf("token validation failed: %v", err),
			})
			ginCtx.Abort()
			return
		}

		// Step 4: Extract the claims from the result
		// ValidateToken returns jwt.MapClaims (which is map[string]interface{})
		claimsMap, ok := claims.(map[string]interface{})
		if !ok {
			ginCtx.JSON(http.StatusUnauthorized, api.ResponseEnvelope{
				Code:    http.StatusUnauthorized,
				Message: "invalid token claims",
				Data:    nil,
				Error:   "unexpected claims type",
			})
			ginCtx.Abort()
			return
		}

		// Extract the fields we care about from the claims map
		userID, _ := claimsMap["user_id"].(string)
		email, _ := claimsMap["email"].(string)

		// Store in Gin context for use in handlers
		ginCtx.Set("user_id", userID)
		ginCtx.Set("email", email)
		ginCtx.Set("claims", claimsMap)

		ginCtx.Next()
	}
}
