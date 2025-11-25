package auth

import (
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/argon2"
)

// generateArgon2Hash generates a valid Argon2id hash for testing
func generateArgon2Hash(password string) string {
	salt := []byte("testsalt12345678") // 16 bytes salt
	memory := uint32(65536)            // 64 MB
	time := uint32(3)                  // 3 iterations
	threads := uint8(2)                // 2 threads
	keyLen := uint32(32)               // 32 bytes output

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	// Format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		time,
		threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func TestVerifyPassword(t *testing.T) {
	correctPassword := "testpassword123"
	wrongPassword := "wrongpassword"

	// Generate a valid hash for testing
	validHash := generateArgon2Hash(correctPassword)

	tests := []struct {
		name        string
		password    string
		encodedHash string
		wantMatch   bool
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "correct password",
			password:    correctPassword,
			encodedHash: validHash,
			wantMatch:   true,
			wantErr:     false,
		},
		{
			name:        "wrong password",
			password:    wrongPassword,
			encodedHash: validHash,
			wantMatch:   false,
			wantErr:     false,
		},
		{
			name:        "invalid hash format - too few parts",
			password:    correctPassword,
			encodedHash: "$argon2id$v=19$m=65536",
			wantMatch:   false,
			wantErr:     true,
			errMsg:      "invalid hash format",
		},
		{
			name:        "unsupported variant",
			password:    correctPassword,
			encodedHash: "$argon2$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG",
			wantMatch:   false,
			wantErr:     true,
			errMsg:      "unsupported variant",
		},
		{
			name:        "incompatible version",
			password:    correctPassword,
			encodedHash: "$argon2id$v=99$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG",
			wantMatch:   false,
			wantErr:     true,
			errMsg:      "incompatible version",
		},
		{
			name:        "invalid base64 salt",
			password:    correctPassword,
			encodedHash: "$argon2id$v=19$m=65536,t=3,p=2$invalid-salt$RdescudvJCsgt3ub+b+dWRWJTmaaJObG",
			wantMatch:   false,
			wantErr:     true,
		},
		{
			name:        "invalid base64 hash",
			password:    correctPassword,
			encodedHash: "$argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$invalid-hash",
			wantMatch:   false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := VerifyPassword(tt.password, tt.encodedHash)

			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("VerifyPassword() expected error but got nil")
					return
				}
				if tt.errMsg != "" {
					errStr := err.Error()
					found := false
					for i := 0; i <= len(errStr)-len(tt.errMsg); i++ {
						if errStr[i:i+len(tt.errMsg)] == tt.errMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("VerifyPassword() error = %v, want error containing %v", err, tt.errMsg)
					}
				}
				return
			}

			if match != tt.wantMatch {
				t.Errorf("VerifyPassword() match = %v, want %v", match, tt.wantMatch)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
