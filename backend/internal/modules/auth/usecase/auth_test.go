package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/VDV001/estimate-pro/backend/internal/modules/auth/domain"
	"github.com/VDV001/estimate-pro/backend/pkg/jwt"
)

// --- Mock UserRepository ---

type mockUserRepo struct {
	createFn           func(ctx context.Context, user *domain.User) error
	getByIDFn          func(ctx context.Context, id string) (*domain.User, error)
	getByEmailFn       func(ctx context.Context, email string) (*domain.User, error)
	updateFn           func(ctx context.Context, user *domain.User) error
	searchFn           func(ctx context.Context, query, excludeUserID string, limit int) ([]*domain.UserSearchResult, error)
	listColleaguesFn   func(ctx context.Context, userID string, limit int) ([]*domain.UserSearchResult, error)
	listRecentlyAddedFn func(ctx context.Context, addedByUserID string, limit int) ([]*domain.UserSearchResult, error)
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

func (m *mockUserRepo) Search(ctx context.Context, query string, excludeUserID string, limit int) ([]*domain.UserSearchResult, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, excludeUserID, limit)
	}
	return nil, nil
}

func (m *mockUserRepo) ListColleagues(ctx context.Context, userID string, limit int) ([]*domain.UserSearchResult, error) {
	if m.listColleaguesFn != nil {
		return m.listColleaguesFn(ctx, userID, limit)
	}
	return nil, nil
}

func (m *mockUserRepo) ListRecentlyAdded(ctx context.Context, addedByUserID string, limit int) ([]*domain.UserSearchResult, error) {
	if m.listRecentlyAddedFn != nil {
		return m.listRecentlyAddedFn(ctx, addedByUserID, limit)
	}
	return nil, nil
}

// --- Mock WorkspaceRepository ---

type mockWorkspaceCreator struct {
	createFn func(ctx context.Context, userID, name string) error
}

func (m *mockWorkspaceCreator) CreatePersonalWorkspace(ctx context.Context, userID, name string) error {
	if m.createFn != nil {
		return m.createFn(ctx, userID, name)
	}
	return nil
}

type mockTokenStore struct {
	tokens map[string]bool
}

func newMockTokenStore() *mockTokenStore {
	return &mockTokenStore{tokens: make(map[string]bool)}
}

func (m *mockTokenStore) Save(_ context.Context, userID, tokenID string, _ time.Duration) error {
	m.tokens[userID+":"+tokenID] = true
	return nil
}

func (m *mockTokenStore) Exists(_ context.Context, userID, tokenID string) (bool, error) {
	return m.tokens[userID+":"+tokenID], nil
}

func (m *mockTokenStore) Delete(_ context.Context, userID, tokenID string) error {
	delete(m.tokens, userID+":"+tokenID)
	return nil
}

func (m *mockTokenStore) DeleteAll(_ context.Context, userID string) error {
	for k := range m.tokens {
		if len(k) > len(userID) && k[:len(userID)+1] == userID+":" {
			delete(m.tokens, k)
		}
	}
	return nil
}

type mockAvatarStorage struct {
	uploadFn   func(ctx context.Context, key string, data []byte, contentType string) (string, error)
	downloadFn func(ctx context.Context, key string) ([]byte, string, error)
}

func (m *mockAvatarStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, key, data, contentType)
	}
	return "/avatars/" + key, nil
}

func (m *mockAvatarStorage) Download(ctx context.Context, key string) ([]byte, string, error) {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, key)
	}
	return []byte("fake-image"), "image/jpeg", nil
}

type mockMembershipChecker struct {
	shareProjectFn func(ctx context.Context, userA, userB string) (bool, error)
}

func (m *mockMembershipChecker) ShareProject(ctx context.Context, userA, userB string) (bool, error) {
	if m.shareProjectFn != nil {
		return m.shareProjectFn(ctx, userA, userB)
	}
	return true, nil
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
		wsRepo    *mockWorkspaceCreator
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
			wsRepo: &mockWorkspaceCreator{
				createFn: func(_ context.Context, _, _ string) error {
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
			wsRepo:    &mockWorkspaceCreator{},
			wantErr:   domain.ErrEmailTaken,
			wantUser:  false,
			wantToken: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.userRepo, tc.wsRepo, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

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
			uc := New(tc.repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

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
			store := newMockTokenStore()
			// Pre-save the valid token so Refresh finds it
			if tc.wantErr == nil {
				store.Save(t.Context(), "user-123", validPair.RefreshID, 7*24*time.Hour)
			}
			uc := New(tc.repo, &mockWorkspaceCreator{}, jwtSvc, store, &mockAvatarStorage{}, &mockMembershipChecker{})

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

func TestRefresh_RevokedToken(t *testing.T) {
	jwtSvc := newTestJWT()
	store := newMockTokenStore()
	repo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "user-1"}, nil
		},
	}
	uc := New(repo, &mockWorkspaceCreator{}, jwtSvc, store, &mockAvatarStorage{}, &mockMembershipChecker{})

	// Generate a token but do NOT save it to the store (simulates revocation)
	pair, _ := jwtSvc.GeneratePair("user-1")

	_, err := uc.Refresh(t.Context(), pair.RefreshToken)
	if !errors.Is(err, domain.ErrTokenRevoked) {
		t.Fatalf("expected ErrTokenRevoked, got: %v", err)
	}
}

func TestRefresh_RotatesTokens(t *testing.T) {
	jwtSvc := newTestJWT()
	store := newMockTokenStore()
	repo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "user-1"}, nil
		},
	}
	uc := New(repo, &mockWorkspaceCreator{}, jwtSvc, store, &mockAvatarStorage{}, &mockMembershipChecker{})

	// Generate and save initial token
	pair, _ := jwtSvc.GeneratePair("user-1")
	store.Save(t.Context(), "user-1", pair.RefreshID, 7*24*time.Hour)

	// Refresh should rotate
	newPair, err := uc.Refresh(t.Context(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh error: %v", err)
	}

	// Old token should be deleted
	exists, _ := store.Exists(t.Context(), "user-1", pair.RefreshID)
	if exists {
		t.Error("old token should be deleted after rotation")
	}

	// New token should exist
	if newPair.AccessToken == "" || newPair.RefreshToken == "" {
		t.Error("expected non-empty new tokens")
	}
}

func TestLogout_DeletesToken(t *testing.T) {
	jwtSvc := newTestJWT()
	store := newMockTokenStore()
	uc := New(&mockUserRepo{}, &mockWorkspaceCreator{}, jwtSvc, store, &mockAvatarStorage{}, &mockMembershipChecker{})

	pair, _ := jwtSvc.GeneratePair("user-1")
	store.Save(t.Context(), "user-1", pair.RefreshID, 7*24*time.Hour)

	err := uc.Logout(t.Context(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("Logout error: %v", err)
	}

	exists, _ := store.Exists(t.Context(), "user-1", pair.RefreshID)
	if exists {
		t.Error("token should be deleted after logout")
	}
}

func TestOAuthLogin_NewUser(t *testing.T) {
	store := newMockTokenStore()
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
		createFn: func(_ context.Context, _ *domain.User) error {
			return nil
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{
		createFn: func(_ context.Context, _, _ string) error { return nil },
	}, newTestJWT(), store, &mockAvatarStorage{}, &mockMembershipChecker{})

	result, err := uc.OAuthLogin(t.Context(), OAuthLoginInput{
		Email:    "oauth@example.com",
		Name:     "OAuth User",
		Provider: "google",
	})
	if err != nil {
		t.Fatalf("OAuthLogin error: %v", err)
	}
	if result.User.Email != "oauth@example.com" {
		t.Errorf("email = %q, want oauth@example.com", result.User.Email)
	}
	if result.TokenPair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestOAuthLogin_ExistingUser(t *testing.T) {
	store := newMockTokenStore()
	existingUser := &domain.User{ID: "user-1", Email: "existing@example.com", Name: "Existing"}
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return existingUser, nil
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), store, &mockAvatarStorage{}, &mockMembershipChecker{})

	result, err := uc.OAuthLogin(t.Context(), OAuthLoginInput{
		Email:    "existing@example.com",
		Name:     "Existing",
		Provider: "github",
	})
	if err != nil {
		t.Fatalf("OAuthLogin error: %v", err)
	}
	if result.User.ID != "user-1" {
		t.Errorf("should return existing user, got ID %q", result.User.ID)
	}
}

func TestLogout_InvalidToken_NoError(t *testing.T) {
	jwtSvc := newTestJWT()
	uc := New(&mockUserRepo{}, &mockWorkspaceCreator{}, jwtSvc, newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	err := uc.Logout(t.Context(), "invalid.token.string")
	if err != nil {
		t.Fatalf("expected nil error for invalid token on logout, got: %v", err)
	}
}

// --- GetCurrentUser ---

func TestGetCurrentUser(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		repo    *mockUserRepo
		wantErr bool
	}{
		{
			name:   "Found",
			userID: "user-1",
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, id string) (*domain.User, error) {
					return &domain.User{ID: id, Email: "alice@example.com", Name: "Alice"}, nil
				},
			},
		},
		{
			name:   "NotFound",
			userID: "nonexistent",
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return nil, domain.ErrUserNotFound
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
			user, err := uc.GetCurrentUser(t.Context(), tc.userID)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.ID != tc.userID {
				t.Errorf("expected ID %q, got %q", tc.userID, user.ID)
			}
		})
	}
}

// --- UpdateProfile ---

func TestUpdateProfile(t *testing.T) {
	telegramID := "12345"
	notifEmail := "notify@example.com"
	empty := ""

	tests := []struct {
		name    string
		input   UpdateProfileInput
		repo    *mockUserRepo
		wantErr bool
		check   func(t *testing.T, user *domain.User)
	}{
		{
			name: "UpdateName",
			input: UpdateProfileInput{
				UserID: "user-1",
				Name:   "New Name",
			},
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "user-1", Name: "Old Name", Email: "a@b.com"}, nil
				},
			},
			check: func(t *testing.T, user *domain.User) {
				if user.Name != "New Name" {
					t.Errorf("expected name 'New Name', got %q", user.Name)
				}
			},
		},
		{
			name: "UpdateTelegramChatID",
			input: UpdateProfileInput{
				UserID:         "user-1",
				TelegramChatID: &telegramID,
			},
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "user-1", Name: "Alice", Email: "a@b.com"}, nil
				},
			},
			check: func(t *testing.T, user *domain.User) {
				if user.TelegramChatID != "12345" {
					t.Errorf("expected telegram_chat_id '12345', got %q", user.TelegramChatID)
				}
			},
		},
		{
			name: "UpdateNotificationEmail",
			input: UpdateProfileInput{
				UserID:            "user-1",
				NotificationEmail: &notifEmail,
			},
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "user-1", Name: "Alice", Email: "a@b.com"}, nil
				},
			},
			check: func(t *testing.T, user *domain.User) {
				if user.NotificationEmail != "notify@example.com" {
					t.Errorf("expected notification_email 'notify@example.com', got %q", user.NotificationEmail)
				}
			},
		},
		{
			name: "ClearTelegramChatID",
			input: UpdateProfileInput{
				UserID:         "user-1",
				TelegramChatID: &empty,
			},
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "user-1", Name: "Alice", TelegramChatID: "old-id"}, nil
				},
			},
			check: func(t *testing.T, user *domain.User) {
				if user.TelegramChatID != "" {
					t.Errorf("expected empty telegram_chat_id, got %q", user.TelegramChatID)
				}
			},
		},
		{
			name: "UserNotFound",
			input: UpdateProfileInput{
				UserID: "nonexistent",
				Name:   "X",
			},
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return nil, domain.ErrUserNotFound
				},
			},
			wantErr: true,
		},
		{
			name: "UpdateError",
			input: UpdateProfileInput{
				UserID: "user-1",
				Name:   "X",
			},
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "user-1", Name: "Old"}, nil
				},
				updateFn: func(_ context.Context, _ *domain.User) error {
					return fmt.Errorf("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
			user, err := uc.UpdateProfile(t.Context(), tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, user)
			}
		})
	}
}

// --- UploadAvatar ---

func TestUploadAvatar(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		storage *mockAvatarStorage
		wantErr bool
	}{
		{
			name: "Success",
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "user-1", Name: "Alice"}, nil
				},
			},
			storage: &mockAvatarStorage{},
		},
		{
			name: "StorageFailure",
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return &domain.User{ID: "user-1", Name: "Alice"}, nil
				},
			},
			storage: &mockAvatarStorage{
				uploadFn: func(_ context.Context, _ string, _ []byte, _ string) (string, error) {
					return "", fmt.Errorf("s3 unavailable")
				},
			},
			wantErr: true,
		},
		{
			name: "UserNotFound",
			repo: &mockUserRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
					return nil, domain.ErrUserNotFound
				},
			},
			storage: &mockAvatarStorage{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), tc.storage, &mockMembershipChecker{})
			user, err := uc.UploadAvatar(t.Context(), "user-1", []byte("img-data"), "image/png")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.AvatarURL == "" {
				t.Error("expected non-empty avatar URL")
			}
		})
	}
}

// --- GetAvatar ---

func TestGetAvatar(t *testing.T) {
	tests := []struct {
		name       string
		callerID   string
		targetID   string
		membership *mockMembershipChecker
		storage    *mockAvatarStorage
		wantErr    bool
	}{
		{
			name:       "SelfAccess",
			callerID:   "user-1",
			targetID:   "user-1",
			membership: &mockMembershipChecker{},
			storage:    &mockAvatarStorage{},
		},
		{
			name:     "SharedProject",
			callerID: "user-1",
			targetID: "user-2",
			membership: &mockMembershipChecker{
				shareProjectFn: func(_ context.Context, _, _ string) (bool, error) {
					return true, nil
				},
			},
			storage: &mockAvatarStorage{},
		},
		{
			name:     "AccessDenied",
			callerID: "user-1",
			targetID: "user-2",
			membership: &mockMembershipChecker{
				shareProjectFn: func(_ context.Context, _, _ string) (bool, error) {
					return false, nil
				},
			},
			storage: &mockAvatarStorage{},
			wantErr: true,
		},
		{
			name:     "MembershipCheckError",
			callerID: "user-1",
			targetID: "user-2",
			membership: &mockMembershipChecker{
				shareProjectFn: func(_ context.Context, _, _ string) (bool, error) {
					return false, fmt.Errorf("db error")
				},
			},
			storage: &mockAvatarStorage{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(&mockUserRepo{}, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), tc.storage, tc.membership)
			data, contentType, err := uc.GetAvatar(t.Context(), tc.callerID, tc.targetID)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(data) == 0 {
				t.Error("expected non-empty data")
			}
			if contentType == "" {
				t.Error("expected non-empty content type")
			}
		})
	}
}

// --- SearchUsers ---

func TestSearchUsers(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		wantLen int
		wantErr bool
	}{
		{
			name: "ReturnsResults",
			repo: &mockUserRepo{
				searchFn: func(_ context.Context, _ string, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return []*domain.UserSearchResult{
						{ID: "u1", Email: "a@b.com", Name: "Alice"},
						{ID: "u2", Email: "c@d.com", Name: "Bob"},
					}, nil
				},
			},
			wantLen: 2,
		},
		{
			name: "EmptyResults",
			repo: &mockUserRepo{
				searchFn: func(_ context.Context, _ string, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return nil, nil
				},
			},
			wantLen: 0,
		},
		{
			name: "Error",
			repo: &mockUserRepo{
				searchFn: func(_ context.Context, _ string, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
			results, err := uc.SearchUsers(t.Context(), "alice", "caller-1", 10)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != tc.wantLen {
				t.Errorf("expected %d results, got %d", tc.wantLen, len(results))
			}
		})
	}
}

// --- ListColleagues ---

func TestListColleagues(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		wantLen int
		wantErr bool
	}{
		{
			name: "ReturnsList",
			repo: &mockUserRepo{
				listColleaguesFn: func(_ context.Context, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return []*domain.UserSearchResult{
						{ID: "u1", Name: "Alice"},
					}, nil
				},
			},
			wantLen: 1,
		},
		{
			name: "EmptyList",
			repo: &mockUserRepo{
				listColleaguesFn: func(_ context.Context, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return nil, nil
				},
			},
			wantLen: 0,
		},
		{
			name: "Error",
			repo: &mockUserRepo{
				listColleaguesFn: func(_ context.Context, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
			results, err := uc.ListColleagues(t.Context(), "user-1", 10)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != tc.wantLen {
				t.Errorf("expected %d results, got %d", tc.wantLen, len(results))
			}
		})
	}
}

// --- ListRecentlyAdded ---

func TestListRecentlyAdded(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		wantLen int
		wantErr bool
	}{
		{
			name: "ReturnsList",
			repo: &mockUserRepo{
				listRecentlyAddedFn: func(_ context.Context, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return []*domain.UserSearchResult{
						{ID: "u1", Name: "Alice"},
						{ID: "u2", Name: "Bob"},
					}, nil
				},
			},
			wantLen: 2,
		},
		{
			name: "Error",
			repo: &mockUserRepo{
				listRecentlyAddedFn: func(_ context.Context, _ string, _ int) ([]*domain.UserSearchResult, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(tc.repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
			results, err := uc.ListRecentlyAdded(t.Context(), "user-1", 10)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != tc.wantLen {
				t.Errorf("expected %d results, got %d", tc.wantLen, len(results))
			}
		})
	}
}

// --- OAuthLogin extra branches ---

func TestOAuthLogin_UpdatesExistingUserProfile(t *testing.T) {
	var updatedUser *domain.User
	existingUser := &domain.User{ID: "user-1", Email: "e@x.com", Name: "Old Name", AvatarURL: "old-url"}
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return existingUser, nil
		},
		updateFn: func(_ context.Context, u *domain.User) error {
			updatedUser = u
			return nil
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	result, err := uc.OAuthLogin(t.Context(), OAuthLoginInput{
		Email:     "e@x.com",
		Name:      "New Name",
		AvatarURL: "new-url",
		Provider:  "google",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %q", result.User.Name)
	}
	if result.User.AvatarURL != "new-url" {
		t.Errorf("expected avatar 'new-url', got %q", result.User.AvatarURL)
	}
	if updatedUser == nil {
		t.Fatal("expected Update to be called")
	}
}

func TestOAuthLogin_CreateUserError(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
		createFn: func(_ context.Context, _ *domain.User) error {
			return fmt.Errorf("db error")
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.OAuthLogin(t.Context(), OAuthLoginInput{
		Email: "new@example.com", Name: "User", Provider: "github",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOAuthLogin_GetByEmailDBError(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, fmt.Errorf("db connection failed")
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.OAuthLogin(t.Context(), OAuthLoginInput{
		Email: "a@b.com", Name: "X", Provider: "google",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Register extra branches ---

func TestRegister_GetByEmailDBError(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, fmt.Errorf("db connection failed")
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.Register(t.Context(), RegisterInput{Email: "a@b.com", Password: "pass", Name: "X"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegister_CreateUserError(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
		createFn: func(_ context.Context, _ *domain.User) error {
			return fmt.Errorf("insert failed")
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.Register(t.Context(), RegisterInput{Email: "a@b.com", Password: "pass", Name: "X"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegister_WorkspaceCreateError(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	wsCreator := &mockWorkspaceCreator{
		createFn: func(_ context.Context, _, _ string) error {
			return fmt.Errorf("workspace creation failed")
		},
	}
	uc := New(userRepo, wsCreator, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.Register(t.Context(), RegisterInput{Email: "a@b.com", Password: "pass", Name: "X"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Login extra branch ---

func TestLogin_GetByEmailDBError(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, fmt.Errorf("db connection failed")
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.Login(t.Context(), LoginInput{Email: "a@b.com", Password: "pass"})
	if err == nil {
		t.Fatal("expected error")
	}
	// Should NOT be ErrInvalidCredentials — it's a different error path
	if errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatal("db error should not be mapped to ErrInvalidCredentials")
	}
}

// --- Refresh extra branch: user not found after token valid ---

func TestRefresh_UserNotFound(t *testing.T) {
	jwtSvc := newTestJWT()
	store := newMockTokenStore()
	pair, _ := jwtSvc.GeneratePair("user-gone")
	store.Save(t.Context(), "user-gone", pair.RefreshID, 7*24*time.Hour)

	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, jwtSvc, store, &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.Refresh(t.Context(), pair.RefreshToken)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

// --- ForgotPassword: not configured ---

func TestForgotPassword_NotConfigured(t *testing.T) {
	uc := New(&mockUserRepo{}, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
	// resetTokenStore is nil — not configured

	out, err := uc.ForgotPassword(t.Context(), "alice@example.com")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.Token != "" {
		t.Fatalf("expected empty token when reset not configured, got %q", out.Token)
	}
}

// --- ResetPassword: not configured ---

func TestResetPassword_NotConfigured(t *testing.T) {
	uc := New(&mockUserRepo{}, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	err := uc.ResetPassword(t.Context(), ResetPasswordInput{Token: "any", NewPassword: "pass"})
	if err == nil {
		t.Fatal("expected error when reset not configured")
	}
}

// --- ForgotPasswordByUserID: not configured ---

func TestForgotPasswordByUserID_NotConfigured(t *testing.T) {
	uc := New(&mockUserRepo{}, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.ForgotPasswordByUserID(t.Context(), "user-1")
	if err == nil {
		t.Fatal("expected error when reset not configured")
	}
}

// --- ResetLink ---

func TestResetLink(t *testing.T) {
	uc := New(&mockUserRepo{}, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
	uc.SetResetConfig(newMockResetTokenStore(), "https://app.example.com", 30*time.Minute)

	link := uc.ResetLink("test-token-123")
	expected := "https://app.example.com/reset-password?token=test-token-123"
	if link != expected {
		t.Errorf("expected %q, got %q", expected, link)
	}
}

// --- SetResetNotifier ---

func TestForgotPassword_WithNotifier(t *testing.T) {
	resetStore := newMockResetTokenStore()
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{
				ID:           "user-1",
				Email:        "alice@example.com",
				PasswordHash: hashPassword(t, "secret"),
			}, nil
		},
	}
	uc := New(userRepo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})
	uc.SetResetConfig(resetStore, "https://app.example.com", 30*time.Minute)

	notified := make(chan bool, 1)
	uc.SetResetNotifier(&mockResetNotifier{
		notifyFn: func(_ context.Context, _, _ string) error {
			notified <- true
			return nil
		},
	})

	out, err := uc.ForgotPassword(t.Context(), "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Token == "" {
		t.Fatal("expected non-empty token")
	}

	// Wait briefly for the goroutine to fire
	select {
	case <-notified:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("notifier was not called within timeout")
	}
}

// --- mockResetNotifier ---

type mockResetNotifier struct {
	notifyFn func(ctx context.Context, userID, resetLink string) error
}

func (m *mockResetNotifier) NotifyReset(ctx context.Context, userID, resetLink string) error {
	if m.notifyFn != nil {
		return m.notifyFn(ctx, userID, resetLink)
	}
	return nil
}

// --- OAuthLogin workspace creation error ---

func TestOAuthLogin_WorkspaceCreateError(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	wsCreator := &mockWorkspaceCreator{
		createFn: func(_ context.Context, _, _ string) error {
			return fmt.Errorf("workspace error")
		},
	}
	uc := New(userRepo, wsCreator, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.OAuthLogin(t.Context(), OAuthLoginInput{
		Email: "new@example.com", Name: "User", Provider: "google",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- UploadAvatar: update save error ---

func TestUploadAvatar_SaveError(t *testing.T) {
	repo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "user-1", Name: "Alice"}, nil
		},
		updateFn: func(_ context.Context, _ *domain.User) error {
			return fmt.Errorf("db error")
		},
	}
	uc := New(repo, &mockWorkspaceCreator{}, newTestJWT(), newMockTokenStore(), &mockAvatarStorage{}, &mockMembershipChecker{})

	_, err := uc.UploadAvatar(t.Context(), "user-1", []byte("data"), "image/png")
	if err == nil {
		t.Fatal("expected error when update fails")
	}
}
