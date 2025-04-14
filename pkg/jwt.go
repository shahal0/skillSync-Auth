package pkg

import (
	"errors"
	"log"
	"strings"
	"time"

	//"github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/v4"
)

// JWTMaker is responsible for generating and validating JWT tokens
type JWTMaker struct {
	secretKey string
}

// CustomClaims holds the data encoded in the JWT
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// NewJWTMaker creates a new JWT utility
func NewJWTMaker(secretKey string) *JWTMaker {
	return &JWTMaker{secretKey: secretKey}
}

// GenerateToken creates a JWT token for the given user ID and role
func (j *JWTMaker) GenerateToken(userID, role string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	log.Println("Generated Token:", signedToken)
	return signedToken, nil
}

// ValidateToken parses and validates a JWT token
func (j *JWTMaker) ValidateToken(tokenStr string) (*CustomClaims, error) {
	log.Println("Validating Token:", tokenStr)
	log.Println("Secret Key Used:", j.secretKey)

	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		log.Println("Token Validation Error:", err)
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	log.Println("Token Claims:", claims)
	return claims, nil
}

// ExtractUserIDFromToken extracts the user ID from a JWT token
func (j *JWTMaker) ExtractUserIDFromToken(tokenStr string) (string, error) {
	log.Println("Extracting User ID from Token:", tokenStr)

	claims := &CustomClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
	if err != nil {
		log.Println("Token Parsing Error:", err)
		return "", errors.New("invalid token")
	}

	log.Println("Extracted Claims:", claims)
	return claims.UserID, nil
}

// ExtractTokenFromHeader extracts the token from the Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	// Check if the header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid Authorization header format: missing 'Bearer ' prefix")
	}

	// Extract the token by removing the "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}
