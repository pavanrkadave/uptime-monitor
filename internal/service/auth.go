package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pavanrkadave/uptime-monitor/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type AuthService struct {
	repo      UserRepository
	jwtSecret []byte
	log       *slog.Logger
}

func NewAuthService(repo UserRepository, jwtSecret string, log *slog.Logger) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		log:       log.With(slog.String("component", "auth-service")),
	}
}

func (a *AuthService) Register(ctx context.Context, email string, password string, role domain.Role) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	return a.repo.Create(ctx, user)
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			a.log.Warn("login failed: user not found", slog.String("email", email))
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		a.log.Warn("login failed: incorrect password", slog.String("email", email))
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Minute * 30).Unix(),
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
