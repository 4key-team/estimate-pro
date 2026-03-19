package domain

import "context"

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
