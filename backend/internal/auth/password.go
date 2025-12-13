package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// VerifyPassword verifies a password against an Argon2 hash.
// The hash should be in the format: $variant$v=version$m=memory,t=time,p=threads$salt$hash
//
// It supports both Argon2id and Argon2i variants with any valid parameters (m, t, p).
// The function automatically extracts and uses the parameters from the hash itself,
// making it compatible with any Argon2 hash generated with standard tools.
//
// It uses constant-time comparison to prevent timing attacks.
//
// Returns true if the password matches the hash, false otherwise.
// Returns an error if the hash format is invalid or unsupported.
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

// HashPassword generates an Argon2id hash for a password.
// Uses increased parameters for better security: 64MB memory, 1 iteration, 4 threads.
// Returns a hash in the format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
func HashPassword(password string) (string, error) {
	// Generate random 16-byte salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Parameters for Argon2id
	memory := uint32(64 * 1024) // 64 MB
	time := uint32(1)           // 1 iteration
	threads := uint8(4)         // 4 threads
	keyLen := uint32(32)        // 32 bytes output

	// Generate Argon2id hash
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		time,
		threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encodedHash, nil
}
