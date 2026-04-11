package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	adminPassword string
	jwtSecret     []byte
	log           *slog.Logger
}

func NewAuthService(adminPassword, jwtSecret string, log *slog.Logger) *AuthService {
	return &AuthService{
		adminPassword: adminPassword,
		jwtSecret:     []byte(jwtSecret),
		log:           log.With(slog.String("component", "auth-service")),
	}
}

func (a *AuthService) Login(ctx context.Context, password string) (string, error) {
	if password != a.adminPassword {
		a.log.Warn("failed login attempt")
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub": "admin",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute * 30).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.jwtSecret)
	if err != nil {
		a.log.Warn("failed to sign token", slog.Any("error", err))
		return "", err
	}

	a.log.Info("token generated successfully")
	return signedToken, nil
}
