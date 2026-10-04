package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	config "github.com/sebnow/chud/platform/config"
)

type UserClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	Version  int    `json:"token_version"`
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

	if claims.UserID == "" || claims.Username == "" {
		return nil, fmt.Errorf("JWT token is missing user claims")
	}

	return claims, nil
}

func NewJwt(userID, username string, isAdmin bool, version int) (string, error) {
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
			UserID:   userID,
			Username: username,
			IsAdmin:  isAdmin,
			Version:  version,
		},
	)
	token, err := jwtToken.SignedString([]byte(os.Getenv(config.JwtSecret)))
	if err != nil {
		return "", fmt.Errorf("failed to create JWT token: %w", err)
	}

	return token, nil
}
