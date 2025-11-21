package jwtmanager

import (
	"errors"
	"time"

	"shared/pkgs/keys"

	"github.com/golang-jwt/jwt/v4"
)

// CustomClaims combines standard JWT claims with dynamic custom claims
type CustomClaims struct {
	jwt.RegisteredClaims
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// TokenPair holds an access and refresh token
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessExp    time.Time
	RefreshExp   time.Time
}

// TTLs
var (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// GenerateTokenPair creates a signed access + refresh token pair with dynamic claims
func GenerateTokenPair(customClaims map[string]interface{}) (*TokenPair, error) {
	now := time.Now()

	accessClaims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
		Custom: customClaims,
	}

	refreshClaims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenTTL)),
		},
		Custom: customClaims,
	}

	accessToken, err := signToken(accessClaims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := signToken(refreshClaims)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessExp:    now.Add(AccessTokenTTL),
		RefreshExp:   now.Add(RefreshTokenTTL),
	}, nil
}

// signToken signs a token using the private key from keys package
func signToken(claims *CustomClaims) (string, error) {
	priv := keys.GetPrivateKey()
	if priv == nil {
		return "", errors.New("no private key available")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(priv)
}

// VerifyToken parses and validates a token, returning claims
func VerifyToken(tokenString string) (*CustomClaims, error) {
	pub := keys.GetPublicKey()
	if pub == nil {
		return nil, errors.New("no public key available")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return pub, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshTokenPair generates a new token pair using an existing valid refresh token
func RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	claims, err := VerifyToken(refreshToken)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if claims.ExpiresAt == nil || now.After(claims.ExpiresAt.Time) {
		return nil, errors.New("refresh token expired")
	}

	return GenerateTokenPair(claims.Custom)
}
