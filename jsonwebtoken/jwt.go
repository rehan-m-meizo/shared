package jsonwebtoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

type SecretClaims struct {
	Key string `json:"key"`
	IV  string `json:"iv"`
	jwt.RegisteredClaims
}

func GenerateCustomJWT(claimString, secretKey string) (string, error) {
	jwtKey := []byte(secretKey)

	claims := &Claims{
		UserId: claimString,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "Signing error", err
	}

	return tokenString, nil
}

func VerifyCustomJWT(jwtToken, secretKey string) (string, error) {
	jwtKey := []byte(secretKey)
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "Authorization error", err
	}

	if !tkn.Valid {
		return "Authorization error", errors.New("invalid token")
	}

	return claims.UserId, nil
}

func GenerateSecretJWT(field string, secret string, expiration time.Duration, issuer string) (string, error) {
	jwtKey := []byte(secret)

	expirationTime := time.Now().Add(expiration)

	claims := &Claims{
		UserId: field,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "Signing error", err
	}

	return tokenString, nil
}

func GenerateJWTNOExpiry(userId string, secret string, expiration time.Duration, issuer string) (string, error) {
	jwtKey := []byte(secret)

	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "Signing Error", err
	}

	return tokenString, nil
}

func GenerateRefreshAndAccessToken(
	userId, secret, issuer string,
	accessExpiration, refreshExpiraton time.Duration,
) (string, string, error) {

	jwtKey := []byte(secret)

	accessExpirationTime := time.Now().Add(accessExpiration)
	refreshExpirationTime := time.Now().Add(refreshExpiraton)

	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpirationTime),
			Issuer:    issuer,
		},
	}

	refreshClaims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpirationTime),
			Issuer:    issuer,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessTokenString, err := accessToken.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func GenerateJWT(userId string, secret string, expiration time.Duration, issuer string) (string, error) {
	jwtKey := []byte(secret)

	expirationTime := time.Now().Add(expiration)

	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "Signing error", err
	}

	return tokenString, nil
}

func VerifySecretJWT(tokenString string, secret string) (string, error) {
	jwtKey := []byte(secret)

	claims := &Claims{} // Use StandardClaims instead of RegisteredClaims

	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "", err
	}

	if !tkn.Valid {
		return "", errors.New("invalid token")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return "", errors.New("token expired")
	}

	return claims.UserId, nil
}

func VerifyAccessAndRefreshToken(accessTokenString, refreshTokenString string, secret string, refreshDuration time.Duration) (string, string, error) {

	jwtKey := []byte(secret)

	accessTokenClaims := &Claims{}
	refreshTokenClaims := &Claims{}

	tkn, err := jwt.ParseWithClaims(accessTokenString, accessTokenClaims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "", "", err
	}

	if !tkn.Valid {
		return "", "", errors.New("invalid token")
	}

	if accessTokenClaims.ExpiresAt != nil && accessTokenClaims.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("token expired")
	}

	tkn, err = jwt.ParseWithClaims(refreshTokenString, refreshTokenClaims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "", "", err
	}

	if !tkn.Valid {
		return "", "", errors.New("invalid token")
	}

	if refreshTokenClaims.ExpiresAt != nil && refreshTokenClaims.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("token expired")
	}

	refreshTokenClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(refreshDuration))

	refreshTokenString, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims).SignedString(jwtKey)

	if err != nil {
		return "", "", err
	}

	return accessTokenClaims.UserId, refreshTokenString, nil
}

func VerifyJwt(tokenString string, secret string) (string, error) {
	jwtKey := []byte(secret)
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "Authorization error", err
	}

	if !tkn.Valid {
		return "Authorization error", errors.New("invalid token")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return "Authorization error", errors.New("token expired")
	}

	return claims.UserId, nil
}

func VerifyJWTNOExpiry(tokenString string, secret string) (string, error) {
	jwtKey := []byte(secret)
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return "Authorization error", err
	}

	if !tkn.Valid {
		return "Authorization error", errors.New("invalid token")
	}

	return claims.UserId, nil
}

func VerifyWithClaims(tokenString string, secret string) (*Claims, error) {
	jwtKey := []byte(secret)
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !tkn.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
