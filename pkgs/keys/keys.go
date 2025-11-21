package keys

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
)

var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

func InitKeyPair() error {
	rootDir, err := getGoWorkRoot()
	if err != nil {
		return err
	}

	keyDir := filepath.Join(rootDir, "config", "keys")
	privateKeyPath := filepath.Join(keyDir, "private.pem")
	publicKeyPath := filepath.Join(keyDir, "public.pem")

	// Ensure directory exists
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return err
	}

	// Generate keys if they don't exist
	if !fileExists(privateKeyPath) || !fileExists(publicKeyPath) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		privateFile, err := os.Create(privateKeyPath)
		if err != nil {
			return err
		}
		defer privateFile.Close()

		privatePem := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}
		if err := pem.Encode(privateFile, privatePem); err != nil {
			return err
		}

		publicFile, err := os.Create(publicKeyPath)
		if err != nil {
			return err
		}
		defer publicFile.Close()

		pubASN1, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return err
		}

		publicPem := &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubASN1,
		}
		if err := pem.Encode(publicFile, publicPem); err != nil {
			return err
		}
	}

	// Load keys into memory
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}
	privBlock, _ := pem.Decode(privBytes)
	if privBlock == nil {
		return errors.New("failed to parse private key")
	}
	privateKey, err = x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return err
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return err
	}
	pubBlock, _ := pem.Decode(pubBytes)
	if pubBlock == nil {
		return errors.New("failed to parse public key")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}
	publicKey = pubKey

	return nil
}

func GetPrivateKey() *rsa.PrivateKey {
	return privateKey
}

func GetPublicKey() *rsa.PublicKey {
	return publicKey
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func getGoWorkRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("go.work file not found")
}

func SignPayload(payload interface{}) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifySignature(payload any, base64Sig string) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	sigBytes, err := base64.StdEncoding.DecodeString(base64Sig)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sigBytes)
}
