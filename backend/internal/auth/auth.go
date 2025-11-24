package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/subtle"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"

	"github.com/example/k8s-view/internal/models"
)

var (
	// jwtSecret is the secret key for signing tokens
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

func init() {
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if len(jwtSecretStr) == 0 {
		fmt.Println("CRITICAL: JWT_SECRET environment variable must be set")
		if os.Getenv("GO_ENV") == "production" {
			log.Fatal("JWT_SECRET is required in production")
		}
		if len(jwtSecretStr) == 0 {
			log.Fatal("CRITICAL: JWT_SECRET environment variable must be set")
		}
	}
	if len(jwtSecretStr) < 32 {
		log.Fatal("CRITICAL: JWT_SECRET must be at least 32 characters long")
	}
	jwtSecret = []byte(jwtSecretStr)
}

// AuthClaims extends models.Claims with JWT registered claims
type AuthClaims struct {
	models.Claims
	jwt.RegisteredClaims
}

// Service provides authentication operations
type Service struct {
	handlers *models.Handlers
}

// NewService creates a new auth service
func NewService(h *models.Handlers) *Service {
	return &Service{
		handlers: h,
	}
}

// LoginHandler handles user authentication
func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get admin credentials from env
	adminUser := os.Getenv("ADMIN_USER")
	adminPassHash := os.Getenv("ADMIN_PASSWORD")

	if adminUser == "" || adminPassHash == "" {
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
	claims := &AuthClaims{
		Claims: models.Claims{
			Username: creds.Username,
			Role:     "admin",
		},
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
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	userVal := r.Context().Value("user")
	if userVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to extract username and role from the claims
	var username, role string
	if claims, ok := userVal.(*AuthClaims); ok {
		username = claims.Username
		role = claims.Role
	} else if claims, ok := userVal.(map[string]interface{}); ok {
		if u, ok := claims["username"].(string); ok {
			username = u
		}
		if r, ok := claims["role"].(string); ok {
			role = r
		}
	}

	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
		"role":     role,
	})
}

// AuthMiddleware protects routes
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow OPTIONS for CORS
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		claims, err := s.AuthenticateRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user", claims)
		next(w, r.WithContext(ctx))
	}
}

// AuthenticateRequest extracts and validates JWT from header, query, or cookie
func (s *Service) AuthenticateRequest(r *http.Request) (*AuthClaims, error) {
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

	claims := &AuthClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
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
