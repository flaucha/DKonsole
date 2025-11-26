package main

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <password> <hash>\n", os.Args[0])
		fmt.Println("Example: go run debug-auth.go mypassword '$argon2i$v=19$m=4096,t=3,p=1$salt$hash'")
		os.Exit(1)
	}

	password := os.Args[1]
	encodedHash := os.Args[2]

	fmt.Printf("Verifying password: '%s'\n", password)
	fmt.Printf("Against hash: '%s'\n", encodedHash)

	match, err := VerifyPassword(password, encodedHash)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if match {
		fmt.Println("SUCCESS: Password matches hash!")
	} else {
		fmt.Println("FAILURE: Password does NOT match hash.")
	}
}

// VerifyPassword verifies a password against an Argon2 hash.
// Copied from backend/internal/auth/password.go
func VerifyPassword(password, encodedHash string) (bool, error) {
	// Trim whitespace in case of formatting issues
	encodedHash = strings.TrimSpace(encodedHash)

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format: expected 6 parts separated by '$', got %d", len(parts))
	}

	// Empty first part is expected (format starts with $)
	if parts[0] != "" {
		return false, fmt.Errorf("invalid hash format: must start with '$'")
	}

	var hashFunc func([]byte, []byte, uint32, uint32, uint8, uint32) []byte

	// Support both argon2id and argon2i variants
	variant := strings.ToLower(parts[1])
	switch variant {
	case "argon2id":
		hashFunc = argon2.IDKey
	case "argon2i":
		hashFunc = argon2.Key
	default:
		return false, fmt.Errorf("unsupported variant: %s (supported: argon2id, argon2i)", parts[1])
	}

	// Parse version
	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("invalid version format: %w", err)
	}
	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version: got %d, expected %d", version, argon2.Version)
	}

	// Parse parameters - be flexible with whitespace
	params := strings.TrimSpace(parts[3])
	var memory uint32
	var time uint32
	var threads uint8
	_, err = fmt.Sscanf(params, "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, fmt.Errorf("invalid parameters format: %w", err)
	}

	// Validate reasonable parameter ranges
	if memory == 0 || time == 0 || threads == 0 {
		return false, fmt.Errorf("invalid parameters: memory, time, and threads must be > 0")
	}

	// Decode salt - try RawStdEncoding first (standard), then fallback to StdEncoding
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		// Try standard base64 encoding as fallback
		salt, err = base64.StdEncoding.DecodeString(parts[4])
		if err != nil {
			return false, fmt.Errorf("invalid salt encoding: %w", err)
		}
	}

	// Decode hash - try RawStdEncoding first (standard), then fallback to StdEncoding
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		// Try standard base64 encoding as fallback
		decodedHash, err = base64.StdEncoding.DecodeString(parts[5])
		if err != nil {
			return false, fmt.Errorf("invalid hash encoding: %w", err)
		}
	}

	keyLen := uint32(len(decodedHash))

	// Generate comparison hash using the same parameters from the stored hash
	comparisonHash := hashFunc([]byte(password), salt, time, memory, threads, keyLen)

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
