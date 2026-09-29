package utils

import (
	"errors"
	"time"

	"suppliment/backend/internal/config"
	"suppliment/backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(user models.User, cfg config.Config) (string, error) {
	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTAccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
}

func GenerateRefreshToken(user models.User, cfg config.Config) (string, error) {
	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTRefreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
}

func ParseToken(tokenString, secret string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("missing token")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func MustUUID(value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return id
}
