package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/subtle"
	"encoding/base64"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

var (
	// Secret key for signing tokens. In production, this should be loaded from env.
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-secret-key-change-me")
	}
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// LoginHandler handles user authentication
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get admin credentials from env
	adminUser := os.Getenv("ADMIN_USER")
	adminPassHash := os.Getenv("ADMIN_PASSWORD")

	// Default fallback if not set (FOR DEVELOPMENT ONLY)
	if adminUser == "" {
		adminUser = "admin"
	}

	if creds.Username != adminUser {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password using Argon2
	if adminPassHash == "" {
		// Fallback for dev/testing if no hash is provided
		// WARNING: This is insecure and should be removed in production
		if creds.Password != "password" {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	} else {
		match, err := verifyPassword(creds.Password, adminPassHash)
		if err != nil {
			fmt.Printf("Password verification error: %v\n", err)
			http.Error(w, "Authentication error", http.StatusInternalServerError)
			return
		}
		if !match {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	// Generate JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"role":  "admin",
	})
}

func verifyPassword(password, encodedHash string) (bool, error) {
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

// AuthMiddleware protects routes
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow OPTIONS for CORS
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		claims, err := authenticateRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user", claims)
		next(w, r.WithContext(ctx))
	}
}

// authenticateRequest extracts and validates JWT from header, query, or cookie.
func authenticateRequest(r *http.Request) (*Claims, error) {
	tokenString := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
	} else if q := r.URL.Query().Get("token"); q != "" {
		tokenString = q
	} else if c, err := r.Cookie("token"); err == nil {
		tokenString = c.Value
	}

	if tokenString == "" {
		fmt.Println("Auth Debug: No token found in request")
		return nil, fmt.Errorf("authorization token required")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		fmt.Printf("Auth Debug: Token parse error: %v\n", err)
		return nil, fmt.Errorf("invalid token")
	}
	if !token.Valid {
		fmt.Println("Auth Debug: Token is invalid")
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
