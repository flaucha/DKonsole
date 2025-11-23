package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if len(jwtSecretStr) == 0 {
		// In production this must be fatal.
		// For development convenience we might warn, but the request was to fix it.
		// Given the critical severity, we should enforce it or at least generate a random one if not present (but persistence issues).
		// The requirement says: "log.Fatal".
		// However, to avoid breaking the user's immediate local run if they haven't set it,
		// I will check if we are in a "production" mode or similar, but the prompt asked to fix it.
		// I will follow the solution:
		fmt.Println("CRITICAL: JWT_SECRET environment variable must be set")
		// If we want to be strict as requested:
		if os.Getenv("GO_ENV") == "production" {
			log.Fatal("JWT_SECRET is required in production")
		}
		// Fallback for dev only if absolutely necessary, but better to fail or warn loudly.
		// The prompt solution used log.Fatal. I will use log.Fatal to be safe as requested.
		// But wait, if I break the app now, the user might complain.
		// I'll use a strong warning and maybe a default for dev if not set, OR just Fatal as requested.
		// The user said "soluciona todo lo que se pueda".
		// I'll stick to the requested solution but maybe allow a bypass for dev if needed?
		// No, let's be secure.
		if len(jwtSecretStr) == 0 {
			log.Fatal("CRITICAL: JWT_SECRET environment variable must be set")
		}
	}
	if len(jwtSecretStr) < 32 {
		log.Fatal("CRITICAL: JWT_SECRET must be at least 32 characters long")
	}
	jwtSecret = []byte(jwtSecretStr)
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

	if adminUser == "" || adminPassHash == "" {
		// In production, this should be a fatal error.
		// For now, we return 500 to prevent unauthorized access.
		fmt.Println("Critical: ADMIN_USER or ADMIN_PASSWORD not set")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	if creds.Username != adminUser {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password using Argon2
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
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   true, // Should be true in production (HTTPS)
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// Do not return token in body to prevent storage in localStorage
	json.NewEncoder(w).Encode(map[string]string{
		"role": "admin",
	})
}

// LogoutHandler clears the session cookie
func (h *Handlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out"))
}

// MeHandler returns current user info if authenticated
func (h *Handlers) MeHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"username": claims.Username,
		"role":     claims.Role,
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
		return nil, fmt.Errorf("authorization token required")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
