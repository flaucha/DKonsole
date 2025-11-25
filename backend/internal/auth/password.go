package auth

import (
	"encoding/base64"
	"fmt"
	"strings"

	"crypto/subtle"

	"golang.org/x/crypto/argon2"
)

// VerifyPassword verifies a password against an Argon2 hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var hashFunc func([]byte, []byte, uint32, uint32, uint8, uint32) []byte

	switch parts[1] {
	case "argon2id":
		hashFunc = argon2.IDKey
	case "argon2i":
		hashFunc = argon2.Key
	default:
		return false, fmt.Errorf("unsupported variant: %s", parts[1])
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version")
	}

	var memory uint32
	var time uint32
	var threads uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	keyLen := uint32(len(decodedHash))

	comparisonHash := hashFunc([]byte(password), salt, time, memory, threads, keyLen)

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}



