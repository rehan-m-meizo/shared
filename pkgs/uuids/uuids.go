package uuids

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

const (
	DnsNamespace  = "DNS"
	UrlNamespace  = "URL"
	OidNamespace  = "OID"
	X500Namespace = "X500"
)

var sha1Pool = sync.Pool{
	New: func() interface{} {
		return sha1.New()
	},
}

// NewUUID generates a version 4 UUID (random-based)
func NewUUID() string {
	u := make([]byte, 16)
	if _, err := rand.Read(u); err != nil {
		panic(fmt.Errorf("failed to generate random bytes: %v", err))
	}

	u[6] = (u[6] & 0x0f) | 0x40 // Version 4
	u[8] = (u[8] & 0x3f) | 0x80 // Variant is 10

	return encodeUUID(u)
}

func encodeUUID(u []byte) string {
	var buf [36]byte
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])
	return string(buf[:])
}

var defaultNamespaces = map[string]string{
	"DNS":  "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	"URL":  "6ba7b811-9dad-11d1-80b4-00c04fd430c8",
	"OID":  "6ba7b812-9dad-11d1-80b4-00c04fd430c8",
	"X500": "6ba7b814-9dad-11d1-80b4-00c04fd430c8",
}

// NewUUID5 generates a version 5 UUID using SHA-1 based namespace and name
func NewUUID5(name, namespace string) (string, error) {
	var namespaceUUID string
	if namespace == "" {
		namespaceUUID = defaultNamespaces["URL"]
	} else {
		namespaceUUID = defaultNamespaces[namespace]
	}

	nsBytes, err := parseUUID(namespaceUUID)
	if err != nil {
		return "", fmt.Errorf("invalid namespace UUID: %v", err)
	}

	h := sha1Pool.Get().(interface {
		Write([]byte) (int, error)
		Sum([]byte) []byte
		Reset()
	})
	defer sha1Pool.Put(h)

	h.Reset()
	h.Write(nsBytes)
	h.Write([]byte(name))
	hash := h.Sum(nil)

	// Create a 16-byte UUID from the SHA-1 hash
	var uuid [16]byte
	copy(uuid[:], hash[:16])

	uuid[6] = (uuid[6] & 0x0f) | 0x50 // Version 5
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return encodeUUID(uuid[:]), nil
}

// Parses a UUID string and returns a 16-byte array
func parseUUID(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return nil, errors.New("invalid UUID length")
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != 16 {
		return nil, errors.New("invalid UUID byte length")
	}
	return b, nil
}

func GenerateRandomFilename(fileBytes []byte) (string, error) {
	kind := http.DetectContentType(fileBytes)

	if kind == "" {
		return "", errors.New("failed to detect file type")
	}

	fileName := NewUUID() + "." + kind

	fileName, err := NewUUID5(fileName, DnsNamespace)

	if err != nil {
		return "", err
	}

	return fileName + "." + kind, nil
}
