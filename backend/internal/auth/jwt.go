package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// JWTService provides JWT token operations including extraction, validation, and authentication.
type JWTService struct {
	jwtSecret []byte // Secret key used for signing and verifying JWT tokens
}

// NewJWTService creates a new JWTService with the provided JWT secret.
// The secret should be a secure random byte array, typically loaded from environment variables.
func NewJWTService(jwtSecret []byte) *JWTService {
	return &JWTService{
		jwtSecret: jwtSecret,
	}
}

// AuthClaims extends models.Claims with JWT registered claims.
// It contains user information (username, role) and standard JWT claims (expiration, etc.).
type AuthClaims struct {
	models.Claims
	jwt.RegisteredClaims
}

// ExtractToken extracts JWT token from HTTP request.
// It checks for the token in the following order:
//  1. Authorization header (Bearer token)
//  2. Query parameter "token"
//  3. Cookie named "token"
//
// Returns an error if no token is found in any of these locations.
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

// ValidateToken validates a JWT token and returns the claims if valid.
// It verifies the token signature using the JWT secret and checks token expiration.
//
// Returns an error if the token is invalid, expired, or malformed.
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

// AuthenticateRequest extracts and validates JWT from HTTP request.
// This is a convenience method that combines ExtractToken and ValidateToken.
//
// Returns the authenticated claims if successful, or an error if authentication fails.
func (s *JWTService) AuthenticateRequest(r *http.Request) (*AuthClaims, error) {
	tokenString, err := s.ExtractToken(r)
	if err != nil {
		return nil, err
	}

	return s.ValidateToken(tokenString)
}
