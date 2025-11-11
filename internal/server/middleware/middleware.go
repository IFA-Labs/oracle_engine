package middleware

import (
	"net/http"
	"oracle_engine/internal/logging"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CORS middleware for standard http
// Allows all origins for GET requests only
// Only specific origins (dashboard.ifalabs.com and localhost:3000) can use all HTTP methods
func CORS(next http.Handler) http.Handler {
	allowedOriginsForAllMethods := []string{
		"https://dashboard.ifalabs.com",
		"http://localhost:3000",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		method := r.Method

		// Always allow GET requests from any origin
		if method == "GET" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		} else {
			// For non-GET methods, check if origin is in allowed list
			isAllowedOrigin := false
			for _, allowedOrigin := range allowedOriginsForAllMethods {
				if origin == allowedOrigin {
					isAllowedOrigin = true
					break
				}
			}

			if isAllowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				// Reject non-GET requests from non-allowed origins
				w.Header().Set("Access-Control-Allow-Origin", "")
			}
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SetupCORS configures CORS middleware for Gin
// Allows all origins for GET requests only
// Only specific origins (dashboard.ifalabs.com and localhost:3000) can use all HTTP methods
func SetupCORS() gin.HandlerFunc {
	allowedOriginsForAllMethods := []string{
		"https://dashboard.ifalabs.com",
		"http://localhost:3000",
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		method := c.Request.Method

		// Always allow GET requests from any origin
		if method == "GET" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
			c.Header("Access-Control-Expose-Headers", "Content-Length")
		} else {
			// For non-GET methods, check if origin is in allowed list
			isAllowedOrigin := false
			for _, allowedOrigin := range allowedOriginsForAllMethods {
				if origin == allowedOrigin {
					isAllowedOrigin = true
					break
				}
			}

			if isAllowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
				c.Header("Access-Control-Expose-Headers", "Content-Length")
				c.Header("Access-Control-Allow-Credentials", "true")
			} else {
				// Reject non-GET requests from non-allowed origins
				c.Header("Access-Control-Allow-Origin", "")
			}
		}

		// Handle preflight OPTIONS requests
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLogger logs HTTP requests for Gin
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/api/health"},
	})
}

// PlaceholderLimiter is a legacy placeholder to maintain compatibility where the old
// function-based rate limiter was referenced. Prefer using the RateLimiter struct.
func PlaceholderLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting logic
		c.Next()
	}
}

// Logging middleware for standard http
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logging.Logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

// Recovery middleware for standard http
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logging.Logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", r.URL.Path),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
