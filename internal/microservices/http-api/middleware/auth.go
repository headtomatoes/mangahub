package middleware

import (
	"mangahub/internal/microservices/http-api/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a Gin middleware for JWT authentication of API requests
// It checks for the presence and validity of a JWT token in the Authorization header
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization") // extract the Authorization header
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ") // split by space , 0 is Bearer, 1 is token
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user info in context for handlers to use
		c.Set("claims", claims)
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("scopes", claims.Scopes)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// All under here are scope-related middlewares use in route protection
// RequireScopes middleware checks if token has required scopes
func RequireScopes(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get scopes from context (set by AuthMiddleware)
		scopesInterface, exists := c.Get("scopes")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scopes not found in token"})
			c.Abort()
			return
		}

		tokenScopes, ok := scopesInterface.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid scope format"})
			c.Abort()
			return
		}

		// Check if token has all required scopes
		if !hasAllScopes(tokenScopes, requiredScopes) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "Insufficient scopes",
				"required": requiredScopes,
				"granted":  tokenScopes,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasAllScopes checks if token has all required scopes
func hasAllScopes(tokenScopes, requiredScopes []string) bool {
	// Create a map for efficient lookup rather than nested loops
	scopeMap := make(map[string]bool)
	for _, scope := range tokenScopes {
		scopeMap[scope] = true
	}

	// Check for wildcard admin scope
	if scopeMap["*"] || scopeMap["admin:*"] {
		return true
	}

	// Check each required scope
	for _, required := range requiredScopes {
		if !scopeMap[required] {
			// Check for wildcard matches (e.g., "read:*" for "read:products")
			if !matchesWildcardScope(tokenScopes, required) {
				return false
			}
		}
	}

	return true
}

// matchesWildcardScope handles wildcard scope matching
func matchesWildcardScope(tokenScopes []string, required string) bool {
	for _, scope := range tokenScopes {
		if len(scope) > 0 && scope[len(scope)-1] == '*' {
			prefix := scope[:len(scope)-1]
			if strings.HasPrefix(required, prefix) {
				return true
			}
		}
	}
	return false
}

// RequireAnyScope checks if token has ANY of the specified scopes
func RequireAnyScope(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopesInterface, exists := c.Get("scopes")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Scopes not found"})
			c.Abort()
			return
		}

		tokenScopes := scopesInterface.([]string)

		for _, required := range scopes {
			for _, granted := range tokenScopes {
				if granted == required {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient scopes"})
		c.Abort()
	}
}

// RequireRole checks if the user has the specified role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by AuthMiddleware)
		roleInterface, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		userRole, ok := roleInterface.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role format"})
			c.Abort()
			return
		}

		// Check if user has the required role
		if userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "Insufficient permissions",
				"required": requiredRole,
				"current":  userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience function for requiring admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}
