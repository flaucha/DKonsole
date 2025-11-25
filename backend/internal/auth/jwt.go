package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/example/k8s-view/internal/models"
)

// JWTService provides JWT token operations
type JWTService struct {
	jwtSecret []byte
}

// NewJWTService creates a new JWTService
func NewJWTService(jwtSecret []byte) *JWTService {
	return &JWTService{
		jwtSecret: jwtSecret,
	}
}

// AuthClaims extends models.Claims with JWT registered claims
type AuthClaims struct {
	models.Claims
	jwt.RegisteredClaims
}

// ExtractToken extracts JWT token from request (header, query, or cookie)
func (s *JWTService) ExtractToken(r *http.Request) (string, error) {
	tokenString := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
	} else if q := r.URL.Query().Get("token"); q != "" {
		tokenString = q
	} else if c, err := r.Cookie("token"); err == nil {
		tokenString = c.Value
	}

	if tokenString == "" {
		return "", fmt.Errorf("authorization token required")
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*AuthClaims, error) {
	claims := &AuthClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// AuthenticateRequest extracts and validates JWT from request
func (s *JWTService) AuthenticateRequest(r *http.Request) (*AuthClaims, error) {
	tokenString, err := s.ExtractToken(r)
	if err != nil {
		return nil, err
	}

	return s.ValidateToken(tokenString)
}





