package oauth

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// KeycloakClaims represents the claims from a Keycloak access token
type KeycloakClaims struct {
	jwt.RegisteredClaims

	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	SessionState      string `json:"session_state"`
	Scope             string `json:"scope"`
	ACR               string `json:"acr"`
	SID               string `json:"sid"`
	AZP               string `json:"azp"` // Authorized party (client ID)
	AUD               any    `json:"aud"` // Audience can be string or []string
	SUB               string `json:"sub"` // Subject
	TYP               string `json:"typ"` // Token type

	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`

	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`

	// Optional field: If you're using custom attributes in Keycloak
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

var keycloakJWKS *keyfunc.JWKS

// Initialize JWKS once globally
func InitKeycloakMiddleware(keycloakURL, realm string) {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, realm)

	var err error
	keycloakJWKS, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshTimeout:    10 * time.Second,
		RefreshUnknownKID: true,
		RefreshErrorHandler: func(err error) {
			fmt.Printf("❌ JWKS refresh error: %v\n", err)
		},
	})

	if err != nil {
		fmt.Printf("❌ Failed to initialize JWKS from %s: %v\n", jwksURL, err)
	} else {
		fmt.Printf("🔐 JWKS initialized from: %s\n", jwksURL)
	}
}

// AuthMiddleware validates JWT and stores KeycloakClaims in contex

// GetKeycloakClaims returns the KeycloakClaims from context
func GetKeycloakClaims(c *gin.Context) (*KeycloakClaims, bool) {
	claims, exists := c.Get("keycloak_claims")
	if !exists {
		return nil, false
	}
	keycloakClaims, ok := claims.(*KeycloakClaims)
	return keycloakClaims, ok
}

// ValidateToken returns the validated JWT token
func ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, keycloakJWKS.Keyfunc)
	if err != nil {
		fmt.Printf("❌ Token parse error: %v\n", err)

		// Try parsing without validation to inspect/debug
		parser := jwt.NewParser()
		unverifiedToken, _, parseErr := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		if parseErr == nil {
			fmt.Printf("🔍 Token payload (unverified): %+v\n", unverifiedToken.Claims)
		} else {
			fmt.Printf("❌ Failed to parse token unverified: %v\n", parseErr)
		}

		return nil, err
	}
	if !token.Valid {
		fmt.Println("❌ Token is invalid (signature check failed or expired)")
		return nil, fmt.Errorf("token is not valid")
	}
	return token, nil
}

// ValidateTokenFromContext pulls JWT token from context header and validates
func ValidateTokenFromContext(c *gin.Context) (*jwt.Token, error) {
	tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	return ValidateToken(tokenString)
}

// mapToStruct converts jwt.MapClaims into custom struct
func mapToStruct(m jwt.MapClaims, out interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
