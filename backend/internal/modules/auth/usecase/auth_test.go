package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/daniilrusanov/estimate-pro/backend/internal/modules/auth/domain"
	projectDomain "github.com/daniilrusanov/estimate-pro/backend/internal/modules/project/domain"
	"github.com/daniilrusanov/estimate-pro/backend/pkg/jwt"
)

// --- Mock UserRepository ---

type mockUserRepo struct {
	createFn     func(ctx context.Context, user *domain.User) error
	getByIDFn    func(ctx context.Context, id string) (*domain.User, error)
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	updateFn     func(ctx context.Context, user *domain.User) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}

// --- Mock WorkspaceRepository ---

type mockWorkspaceRepo struct {
	createFn   func(ctx context.Context, workspace *projectDomain.Workspace) error
	getByIDFn  func(ctx context.Context, id string) (*projectDomain.Workspace, error)
	listByUser func(ctx context.Context, userID string) ([]*projectDomain.Workspace, error)
}

func (m *mockWorkspaceRepo) Create(ctx context.Context, workspace *projectDomain.Workspace) error {
	if m.createFn != nil {
		return m.createFn(ctx, workspace)
	}
	return nil
}

func (m *mockWorkspaceRepo) GetByID(ctx context.Context, id string) (*projectDomain.Workspace, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("workspace not found")
}

func (m *mockWorkspaceRepo) ListByUser(ctx context.Context, userID string) ([]*projectDomain.Workspace, error) {
	if m.listByUser != nil {
		return m.listByUser(ctx, userID)
	}
	return nil, nil
}

// --- Helper ---

func newTestJWT() *jwt.Service {
	return jwt.NewService("test-secret-key-for-unit-tests", 15*time.Minute, 7*24*time.Hour)
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

// --- Tests ---

func TestRegister(t *testing.T) {
	tests := []struct {
		name      string
		input     RegisterInput
		userRepo  *mockUserRepo
		wsRepo    *mockWorkspaceRepo
		wantErr   error
		wantUser  bool
		wantToken bool
	}{
		{
			name: "Success",
			input: RegisterInput{
				Email:    "alice@example.com",
				Password: "strongpassword",
				Name:     "Alice",
			},
			userRepo: &mockUserRepo{
				getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
					return nil, domain.ErrUserNotFound
				},
				createFn: func(_ context.Context, _ *domain.User) error {
					return nil
				},
			},
			wsRepo: &mockWorkspaceRepo{
				createFn: func(_ context.Context, _ *projectDomain.Workspace) error {
					return nil
				},
			},
			wantErr:   nil,
			wantUser:  true,
			wantToken: true,
		},
		{
			name: "DuplicateEmail",
			input: RegisterInput{
				Email:    "taken@example.com",
				Password: "password123",
				Name:     "Bob",
			},
			userRepo: &mockUserRepo{
				getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "existing-id", Email: "taken@example.com"}, nil
				},
			},
			wsRepo:    &mockWorkspaceRepo{},
			wantErr:   domain.ErrEmailTaken,
			wantUser:  false,
			wantToken: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.userRepo, tc.wsRepo, newTestJWT())

			result, err := uc.Register(t.Context(), tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if result != nil {
					t.Fatal("expected nil result on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantUser && result.User == nil {
				t.Fatal("expected user in result")
			}
			if tc.wantUser {
				if result.User.Email != tc.input.Email {
					t.Errorf("expected email %q, got %q", tc.input.Email, result.User.Email)
				}
				if result.User.Name != tc.input.Name {
					t.Errorf("expected name %q, got %q", tc.input.Name, result.User.Name)
				}
				if result.User.ID == "" {
					t.Error("expected non-empty user ID")
				}
			}
			if tc.wantToken {
				if result.TokenPair == nil {
					t.Fatal("expected token pair")
				}
				if result.TokenPair.AccessToken == "" {
					t.Error("expected non-empty access token")
				}
				if result.TokenPair.RefreshToken == "" {
					t.Error("expected non-empty refresh token")
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	const testPassword = "correct-password"

	tests := []struct {
		name    string
		input   LoginInput
		repo    *mockUserRepo
		wantErr error
	}{
		{
			name: "Success",
			input: LoginInput{
				Email:    "alice@example.com",
				Password: testPassword,
			},
			repo: &mockUserRepo{
				getByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
					return &domain.User{
						ID:           "user-1",
						Email:        email,
						PasswordHash: hashPassword(t, testPassword),
						Name:         "Alice",
					}, nil
				},
			},
			wantErr: nil,
		},
		{
			name: "WrongPassword",
			input: LoginInput{
				Email:    "alice@example.com",
				Password: "wrong-password",
			},
			repo: &mockUserRepo{
				getByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
					return &domain.User{
						ID:           "user-1",
						Email:        email,
						PasswordHash: hashPassword(t, testPassword),
						Name:         "Alice",
					}, nil
				},
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name: "UserNotFound",
			input: LoginInput{
				Email:    "nobody@example.com",
				Password: "anything",
			},
			repo: &mockUserRepo{
				getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
					return nil, domain.ErrUserNotFound
				},
			},
			wantErr: domain.ErrInvalidCredentials,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceRepo{}, newTestJWT())

			result, err := uc.Login(t.Context(), tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if result != nil {
					t.Fatal("expected nil result on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.User == nil {
				t.Fatal("expected user in result")
			}
			if result.User.Email != tc.input.Email {
				t.Errorf("expected email %q, got %q", tc.input.Email, result.User.Email)
			}
			if result.TokenPair == nil {
				t.Fatal("expected token pair")
			}
			if result.TokenPair.AccessToken == "" {
				t.Error("expected non-empty access token")
			}
			if result.TokenPair.RefreshToken == "" {
				t.Error("expected non-empty refresh token")
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	jwtSvc := newTestJWT()

	// Generate a valid refresh token for testing
	validPair, err := jwtSvc.GeneratePair("user-123")
	if err != nil {
		t.Fatalf("failed to generate test tokens: %v", err)
	}

	tests := []struct {
		name         string
		refreshToken string
		repo         *mockUserRepo
		wantErr      error
	}{
		{
			name:         "Success",
			refreshToken: validPair.RefreshToken,
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, id string) (*domain.User, error) {
					if id == "user-123" {
						return &domain.User{ID: id, Email: "alice@example.com"}, nil
					}
					return nil, domain.ErrUserNotFound
				},
			},
			wantErr: nil,
		},
		{
			name:         "InvalidToken",
			refreshToken: "not-a-valid-jwt-token",
			repo:         &mockUserRepo{},
			wantErr:      domain.ErrInvalidCredentials,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceRepo{}, jwtSvc)

			tokens, err := uc.Refresh(t.Context(), tc.refreshToken)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if tokens != nil {
					t.Fatal("expected nil tokens on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tokens == nil {
				t.Fatal("expected non-nil tokens")
			}
			if tokens.AccessToken == "" {
				t.Error("expected non-empty access token")
			}
			if tokens.RefreshToken == "" {
				t.Error("expected non-empty refresh token")
			}
		})
	}
}
