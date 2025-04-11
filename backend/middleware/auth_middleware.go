package middleware

import (
	"errors" // Import errors package
	"net/http"
	"strings"
	"vacation-scheduler/backend/handlers" // Import handlers to access Claims and jwtKey

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TODO: Consider moving jwtKey to a shared config package if middleware grows
var jwtKey = []byte("your_very_secret_key_change_this") // Must match the key in auth_handler.go

// AuthMiddleware validates the JWT token from the Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Check if the header format is "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := parts[1]

		// Parse the token
		claims := &handlers.Claims{} // Use the Claims struct from handlers
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid // Or a more specific error
			}
			return jwtKey, nil
		})

		if err != nil {
			// Check for specific JWT errors using errors.Is
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				return
			}
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token signature"})
 				return
 			}
 			// Fallback for other unexpected parsing errors
 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed: " + err.Error()})
			return
		}

		if !token.Valid { // This check might be redundant if ParseWithClaims returns an error for invalid tokens
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Token is valid, set user info in context
		c.Set("userID", claims.UserID)
		c.Set("isAdmin", claims.IsAdmin)

		// Continue to the next handler
		c.Next()
	}
}

// Optional: AdminMiddleware checks if the authenticated user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, run the standard AuthMiddleware logic (or ensure it ran before this)
		// This assumes AuthMiddleware has already run and set the context values.
		// A better approach might be to chain them in the router setup.

		isAdmin, exists := c.Get("isAdmin")
		if !exists {
			// This shouldn't happen if AuthMiddleware ran correctly
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "isAdmin flag not found in context"})
			return
		}

		isAdminBool, ok := isAdmin.(bool)
		if !ok || !isAdminBool {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			return
		}

		// User is admin, continue
		c.Next()
	}
}
