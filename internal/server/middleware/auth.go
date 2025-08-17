package middleware

import (
	"context"
	"net/http"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/server/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	jwtSecret        string
	dashboardService services.DashboardService
}

func NewAuthMiddleware(jwtSecret string, dashboardService services.DashboardService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:        jwtSecret,
		dashboardService: dashboardService,
	}
}

// JWTAuth validates JWT tokens for dashboard endpoints
func (a *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(a.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("email", claims["email"])
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyAuth validates API keys for external API access with rate limiting
func (a *AuthMiddleware) APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Check query parameter as fallback
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		// Validate API key
		keyData, err := a.dashboardService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Check subscription-based rate limits and usage limits
		rateLimited, usageLimitExceeded, err := a.dashboardService.CheckAPILimits(c.Request.Context(), keyData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check API limits"})
			c.Abort()
			return
		}

		if rateLimited {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please wait before making more requests.",
			})
			c.Abort()
			return
		}

		if usageLimitExceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Monthly API usage limit exceeded. Please upgrade your subscription.",
			})
			c.Abort()
			return
		}

		// Record API usage
		go func() {
			if err := a.dashboardService.RecordAPIUsage(
				context.Background(), // Use background context to avoid cancellation
				keyData,
				c.Request.URL.Path,
				c.Request.Method,
				c.ClientIP(),
				c.Request.UserAgent(),
			); err != nil {
				logging.Logger.Error("Failed to record API usage in goroutine", zap.Error(err))
			}
		}()

		c.Set("api_key_id", keyData.ID)
		c.Set("profile_id", keyData.ProfileID)
		c.Next()
	}
}

// OptionalAPIKeyAuth validates API keys but doesn't reject requests without them
func (a *AuthMiddleware) OptionalAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		logging.Logger.Info("Optional API key provided", zap.String("api_key", apiKey))

		if apiKey != "" {
			// Validate API key if provided
			keyData, err := a.dashboardService.ValidateAPIKey(c.Request.Context(), apiKey)
			if err == nil {
				// Record API usage
				go func() {
					if err := a.dashboardService.RecordAPIUsage(
						context.Background(), // Use background context to avoid cancellation
						keyData,
						c.Request.URL.Path,
						c.Request.Method,
						c.ClientIP(),
						c.Request.UserAgent(),
					); err != nil {
						logging.Logger.Error("Failed to record API usage in optional auth goroutine", zap.Error(err))
					}
				}()

				c.Set("api_key_id", keyData.ID)
			}
		}

		c.Next()
	}
}

// FrontendAuth allows both JWT authentication (for official frontend) and API key authentication
func (a *AuthMiddleware) FrontendAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for JWT token first (for official frontend)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			
			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(a.jwtSecret), nil
			})

			if err == nil && token.Valid {
				// Valid JWT - this is the official frontend
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					c.Set("user_id", claims["user_id"])
					c.Set("email", claims["email"])
					c.Set("auth_type", "jwt")
					c.Next()
					return
				}
			}
		}

		// If no valid JWT, check for API key (for third-party integrations)
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required: provide either Bearer token or API key"})
			c.Abort()
			return
		}

		// Validate API key
		keyData, err := a.dashboardService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Check subscription-based rate limits and usage limits
		rateLimited, usageLimitExceeded, err := a.dashboardService.CheckAPILimits(c.Request.Context(), keyData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check API limits"})
			c.Abort()
			return
		}

		if rateLimited {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please wait before making more requests.",
			})
			c.Abort()
			return
		}

		if usageLimitExceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Monthly API usage limit exceeded. Please upgrade your subscription.",
			})
			c.Abort()
			return
		}

		// Record API usage
		go func() {
			if err := a.dashboardService.RecordAPIUsage(
				context.Background(),
				keyData,
				c.Request.URL.Path,
				c.Request.Method,
				c.ClientIP(),
				c.Request.UserAgent(),
			); err != nil {
				logging.Logger.Error("Failed to record API usage in frontend auth goroutine", zap.Error(err))
			}
		}()

		c.Set("api_key_id", keyData.ID)
		c.Set("profile_id", keyData.ProfileID)
		c.Set("auth_type", "api_key")
		c.Next()
	}
}

// ValidateProfileOwnership ensures the user can only access their own resources
func (a *AuthMiddleware) ValidateProfileOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		profileID := c.Param("id")
		if profileID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Profile ID required"})
			c.Abort()
			return
		}

		if userID != profileID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: can only access own resources"})
			c.Abort()
			return
		}

		c.Next()
	}
}
