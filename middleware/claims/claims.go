package claims

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type JWTConfig struct {
	Secret     string // for HS256
	ClientID   string // keycloak client ID
	SigningKey []byte // for HS256
}

// ExtractedClaims is a structured way to access token claims
type ExtractedClaims struct {
	Username    string
	Email       string
	RealmRoles  []string
	ClientRoles []string
	Raw         jwt.MapClaims
}

type KeycloakClaims struct {
	jwt.RegisteredClaims

	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Role              []string `json:"role"`
	GroupCode         []string `json:"group_code"` // Note the typo in your token: it's "company_cdoe"
	ProductCode       []string `json:"product_code"`
}

// Attach parsed claims to context using this key
const ContextKey = "keycloak_claims"

// FromContext retrieves KeycloakClaims from Gin context
func FromContext(c *gin.Context) (*KeycloakClaims, error) {
	val, ok := c.Get(ContextKey)
	if !ok {
		return nil, errors.New("keycloak claims not found in context")
	}
	claims, ok := val.(*KeycloakClaims)
	if !ok {
		return nil, errors.New("invalid keycloak claims format")
	}
	return claims, nil
}
