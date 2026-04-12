package domain

import (
	"errors"
	"time"
)

var (
	ErrDuplicateEmail  = errors.New("user with this email already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidEmail    = errors.New("email address is invalid")
	ErrInvalidPassword = errors.New("password is invalid")
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

type User struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         Role       `json:"role"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}
