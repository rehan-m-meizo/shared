package randoms

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

func GenerateSecureRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GenerateSecureRandomInt(max int64) (int64, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return nBig.Int64(), nil
}

func GenerateSecureRandomAlphaNumericString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	alphaNum := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = alphaNum[b[i]%byte(len(alphaNum))]
	}
	return string(buf), nil
}

// Create merkel tree based Id generator that takes a prefix and a suffix and add integer that incremented by 1 to it
type MerkleTreeIDGenerator struct {
	prefix string
	suffix string
	id     int64
}

func NewMerkleTreeIDGenerator(prefix, suffix string, startID int64) *MerkleTreeIDGenerator {
	return &MerkleTreeIDGenerator{
		prefix: prefix,
		suffix: suffix,
		id:     startID - 1,
	}
}

func (g *MerkleTreeIDGenerator) Generate() string {
	g.id++
	return fmt.Sprintf("%s%d%s", g.prefix, g.id, g.suffix)
}
