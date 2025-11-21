package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"shared/constants"
	"shared/pkgs/jwtmanager"

	"github.com/gin-gonic/gin"
)

type AccessClaims struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	GroupCode   string `json:"group_code"`
	CompanyCode string `json:"company_code"`
}

// AuthcMiddleware verifies JWT access tokens and sets claims in context
func AuthcMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("[AuthcMiddleware] Incoming requests:", c.Request.Method, c.Request.URL.String())

		// 1️⃣ Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("[AuthcMiddleware] Missing Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		// 2️⃣ Validate Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Println("[AuthcMiddleware] Invalid Authorization header format:", authHeader)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}
		tokenStr := parts[1]
		log.Println("[AuthcMiddleware] Extracted token:", tokenStr[:10]+"...") // only print first 10 chars for safety

		// 3️⃣ Verify JWT token
		claims, err := jwtmanager.VerifyToken(tokenStr)
		if err != nil {
			log.Println("[AuthcMiddleware] Token verification failed:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "details": err.Error()})
			return
		}
		log.Println("[AuthcMiddleware] Token verified successfully")

		// 4️⃣ Check expiration
		if claims.ExpiresAt == nil {
			log.Println("[AuthcMiddleware] Token missing expiration")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing expiration"})
			return
		}
		if time.Now().After(claims.ExpiresAt.Time) {
			log.Println("[AuthcMiddleware] Token expired at:", claims.ExpiresAt.Time)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}
		log.Println("[AuthcMiddleware] Token is valid, expires at:", claims.ExpiresAt.Time)

		// 5️⃣ Set claims in context for handlers
		c.Set(constants.KeyClaims, claims)
		log.Println("[AuthcMiddleware] Claims set in context:", claims.Custom)

		// ✅ Continue to next middleware/handler
		c.Next()
	}
}

// Helper to extract claims from context
func ExtractClaims(c *gin.Context) (*jwtmanager.CustomClaims, error) {
	if clm, exists := c.Get(constants.KeyClaims); exists {
		if claims, ok := clm.(*jwtmanager.CustomClaims); ok {
			return claims, nil
		}
	}

	return nil, errors.New("no claims found in context")
}

func GetAccessClaims(c *gin.Context) (*AccessClaims, error) {
	claims, err := ExtractClaims(c)
	if err != nil {
		return nil, err
	}

	userID, ok := claims.Custom["user_id"].(string)

	if !ok {
		return nil, errors.New("no user id found in claims")
	}

	username, ok := claims.Custom["username"].(string)

	if !ok {
		return nil, errors.New("username not found in claims")
	}

	email, ok := claims.Custom["email"].(string)
	if !ok {
		return nil, errors.New("email not found in claims")
	}

	groupCode, ok := claims.Custom["group_code"].(string)
	if !ok {
		return nil, errors.New("group code not found in claims")
	}

	companyCode, ok := claims.Custom["company_code"].(string)

	if !ok {
		return nil, errors.New("company code not found in claims")
	}

	accessClaims := &AccessClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		GroupCode:   groupCode,
		CompanyCode: companyCode,
	}

	return accessClaims, nil
}
