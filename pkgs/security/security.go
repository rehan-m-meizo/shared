package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"os"
	"time"
)

type JWTClaims struct {
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTClaims() *JWTClaims {
	return &JWTClaims{}
}

func CreateJWT(field, secret, issuer string, expiration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserId: field,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func VerifyToken(tokenString, secret string) (*JWTClaims, error) {
	jwtKey := []byte(secret)

	claims := NewJWTClaims()

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("Token valid ")
	}

	return claims, nil
}
func AESEncrypt(key, file string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, fileBytes, nil)

	return ciphertext, nil
}

func AESDecrypt(key string, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func AESEncryptString(key, str string) (string, error) {
	ciphertext, err := AESEncrypt(key, str)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func AESDecryptString(key, str string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}

	plaintext, err := AESDecrypt(key, ciphertext)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func AESFileEncrypt(key string, file string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, fileBytes, nil)

	return ciphertext, nil
}

func AESFileDecrypt(key string, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
