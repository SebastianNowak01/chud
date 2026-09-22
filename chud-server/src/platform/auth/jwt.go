package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	config "github.com/sebnow/chud/platform/config"
)

// UserClaims extends standard JWT claims with custom fields.
type UserClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

func ValidateJwt(token string) (*UserClaims, error) {
	jwtSecret := os.Getenv(config.JwtSecret)
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret is not configured")
	}

	claims := &UserClaims{}

	parsedJwt, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	if !parsedJwt.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	if claims.Username == "" {
		return nil, fmt.Errorf("JWT token is missing username claim")
	}

	return claims, nil
}

func NewJwt(username string, isAdmin bool) (string, error) {
	expiryHours, err := strconv.Atoi(os.Getenv(config.JwtExpiryHours))
	if err != nil {
		return "", fmt.Errorf("invalid JwtExpiryHours value: %w", err)
	}
	now := time.Now()

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiryHours) * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
			Username: username,
			IsAdmin:  isAdmin,
		},
	)
	token, err := jwtToken.SignedString([]byte(os.Getenv(config.JwtSecret)))
	if err != nil {
		return "", fmt.Errorf("failed to create JWT token: %w", err)
	}

	return token, nil
}
