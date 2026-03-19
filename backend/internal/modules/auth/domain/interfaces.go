package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

// WorkspaceCreator creates a personal workspace for a newly registered user.
// Defined here to avoid cross-module import of project domain.
type WorkspaceCreator interface {
	CreatePersonalWorkspace(ctx context.Context, userID, name string) error
}

// AvatarStorage uploads and serves avatar images.
type AvatarStorage interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
}

// TokenStore manages refresh token persistence (Redis-backed).
type TokenStore interface {
	Save(ctx context.Context, userID, tokenID string, ttl time.Duration) error
	Exists(ctx context.Context, userID, tokenID string) (bool, error)
	Delete(ctx context.Context, userID, tokenID string) error
	DeleteAll(ctx context.Context, userID string) error
}
