package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"shared/middleware/claims"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// KeycloakClaims represents the claims from a Keycloak access token

var keycloakJWKS *keyfunc.JWKS

// Initialize JWKS once globally
func InitKeycloakMiddleware(keycloakURL, realm string) error {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, "master")

	var err error
	keycloakJWKS, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshTimeout:    30 * time.Second,
		RefreshUnknownKID: true,
		RefreshErrorHandler: func(err error) {
			fmt.Printf("❌ JWKS refresh error: %v\n", err)
		},
	})

	if err != nil {
		fmt.Printf("Err ", err)
		return fmt.Errorf("❌ Failed to initialize JWKS from %s: %v\n", jwksURL, err)
	} else {
		return nil
	}
}

// GetKeycloakClaims returns the KeycloakClaims from context
func GetKeycloakClaims(c *gin.Context) (*claims.KeycloakClaims, error) {
	clms, exists := c.Get("keycloak_claims")
	if !exists {
		return nil, errors.New("claims for the user not found")
	}
	keycloakClaims, ok := clms.(*claims.KeycloakClaims)

	if !ok {
		return nil, errors.New("claims for the user not found")
	}

	return keycloakClaims, nil
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
