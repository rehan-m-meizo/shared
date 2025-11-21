package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Errors returned by this package.
var (
	ErrInvalidHash         = errors.New("passwords: invalid hash format")
	ErrIncompatibleVersion = errors.New("passwords: incompatible argon2 version")
)

// params holds the parameters used for Argon2id.
type params struct {
	memory      uint32 // in kibibytes
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// DefaultParams are sensible defaults for production use.
// Feel free to tune depending on available memory / threat model.
var DefaultParams = &params{
	memory:      64 * 1024, // 64 MB in KiB
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

// GenerateFromPassword returns an encoded hash string containing:
// $argon2id$v=<version>$m=<memory>,t=<iterations>,p=<parallelism>$<base64Salt>$<base64Hash>
//
// Store this returned string in your database. Use ComparePasswordAndHash to verify.
func GenerateFromPassword(password string) (encodedHash string, err error) {
	p := DefaultParams

	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	// Derive key
	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Encode components
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// ComparePasswordAndHash checks whether the provided password matches the encodedHash.
// Returns true if password is correct.
func ComparePasswordAndHash(password, encodedHash string) (bool, error) {
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// decodeHash extracts params, salt and hash from an encoded string.
func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	// Expected format:
	// $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	// parts[0] == "" because string starts with $
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	// version
	var version int
	if _, err = fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	// params part: m=...,t=...,p=...
	// safer parsing: split by comma and parse each key=val
	paramMap := map[string]uint64{}
	for _, kv := range strings.Split(parts[3], ",") {
		kv = strings.TrimSpace(kv)
		pair := strings.SplitN(kv, "=", 2)
		if len(pair) != 2 {
			return nil, nil, nil, ErrInvalidHash
		}
		key := pair[0]
		valStr := pair[1]
		val, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			return nil, nil, nil, ErrInvalidHash
		}
		paramMap[key] = val
	}

	// Required params m, t, p
	mem, ok1 := paramMap["m"]
	iter, ok2 := paramMap["t"]
	par, ok3 := paramMap["p"]
	if !ok1 || !ok2 || !ok3 {
		return nil, nil, nil, ErrInvalidHash
	}

	// decode salt and hash (base64 raw std)
	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	hash, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	p = &params{
		memory:      uint32(mem),
		iterations:  uint32(iter),
		parallelism: uint8(par),
		saltLength:  uint32(len(salt)),
		keyLength:   uint32(len(hash)),
	}

	return p, salt, hash, nil
}

// generateRandomBytes returns securely generated random bytes.
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
