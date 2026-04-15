package repository

import (
	"testing"
	"time"

	"github.com/VDV001/estimate-pro/backend/internal/modules/auth/domain"
	"github.com/VDV001/estimate-pro/backend/internal/testutil"
)

func TestPostgresUserRepository_Integration(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	repo := NewPostgresUserRepository(pool)
	ctx := t.Context()

	now := time.Now().Truncate(time.Microsecond)
	user := &domain.User{
		ID: "aaaaaaaa-0000-0000-0000-000000000001", Email: "test@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuuv", Name: "Test User",
		PreferredLocale: "ru", CreatedAt: now, UpdatedAt: now,
	}

	t.Run("Create", func(t *testing.T) {
		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Create: %v", err)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		got, err := repo.GetByID(ctx, user.ID)
		if err != nil { t.Fatalf("GetByID: %v", err) }
		if got.Email != user.Email { t.Errorf("email = %q, want %q", got.Email, user.Email) }
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "aaaaaaaa-0000-0000-0000-999999999999")
		if err != domain.ErrUserNotFound { t.Errorf("err = %v, want ErrUserNotFound", err) }
	})

	t.Run("GetByEmail", func(t *testing.T) {
		got, err := repo.GetByEmail(ctx, "test@example.com")
		if err != nil { t.Fatalf("GetByEmail: %v", err) }
		if got.ID != user.ID { t.Errorf("id = %q, want %q", got.ID, user.ID) }
	})

	t.Run("Update", func(t *testing.T) {
		user.Name = "Updated"
		user.TelegramChatID = "123"
		user.UpdatedAt = time.Now().Truncate(time.Microsecond)
		if err := repo.Update(ctx, user); err != nil { t.Fatalf("Update: %v", err) }
		got, _ := repo.GetByID(ctx, user.ID)
		if got.Name != "Updated" { t.Errorf("name = %q", got.Name) }
		if got.TelegramChatID != "123" { t.Errorf("telegram = %q", got.TelegramChatID) }
	})

	t.Run("Search", func(t *testing.T) {
		results, err := repo.Search(ctx, "Updated", "aaaaaaaa-0000-0000-0000-999999999999", 10)
		if err != nil { t.Fatalf("Search: %v", err) }
		if len(results) != 1 { t.Errorf("len = %d, want 1", len(results)) }
	})

	t.Run("Search_ExcludesSelf", func(t *testing.T) {
		results, _ := repo.Search(ctx, "Updated", user.ID, 10)
		if len(results) != 0 { t.Errorf("len = %d, want 0", len(results)) }
	})
}

func TestRedisTokenStore_Integration(t *testing.T) {
	client := testutil.SetupRedis(t)
	store := NewRedisTokenStore(client)
	ctx := t.Context()

	t.Run("Save_Exists", func(t *testing.T) {
		store.Save(ctx, "u1", "t1", 5*time.Minute)
		exists, _ := store.Exists(ctx, "u1", "t1")
		if !exists { t.Error("expected exists") }
	})

	t.Run("NotFound", func(t *testing.T) {
		exists, _ := store.Exists(ctx, "u1", "nope")
		if exists { t.Error("expected not exists") }
	})

	t.Run("Delete", func(t *testing.T) {
		store.Save(ctx, "u1", "del", 5*time.Minute)
		store.Delete(ctx, "u1", "del")
		exists, _ := store.Exists(ctx, "u1", "del")
		if exists { t.Error("expected deleted") }
	})

	t.Run("DeleteAll", func(t *testing.T) {
		store.Save(ctx, "u2", "a", 5*time.Minute)
		store.Save(ctx, "u2", "b", 5*time.Minute)
		store.DeleteAll(ctx, "u2")
		for _, tok := range []string{"a", "b"} {
			exists, _ := store.Exists(ctx, "u2", tok)
			if exists { t.Errorf("%q still exists", tok) }
		}
	})
}

func TestResetTokenStore_Integration(t *testing.T) {
	client := testutil.SetupRedis(t)
	store := NewRedisResetTokenStore(client)
	ctx := t.Context()

	t.Run("Save_Consume", func(t *testing.T) {
		store.Save(ctx, "rt1", "user-abc", 5*time.Minute)
		uid, err := store.Consume(ctx, "rt1")
		if err != nil { t.Fatalf("Consume: %v", err) }
		if uid != "user-abc" { t.Errorf("uid = %q", uid) }
	})

	t.Run("Consume_Twice", func(t *testing.T) {
		store.Save(ctx, "rt2", "user-xyz", 5*time.Minute)
		store.Consume(ctx, "rt2")
		_, err := store.Consume(ctx, "rt2")
		if err == nil { t.Error("expected error") }
	})
}

func TestMembershipChecker_Integration(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	now := time.Now().Truncate(time.Microsecond)

	userA := "aaaaaaaa-ac00-0000-0000-000000000001"
	userB := "aaaaaaaa-ac00-0000-0000-000000000002"
	pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,name,preferred_locale,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		userA, "a@test.com", "$2a$10$x", "A", "ru", now, now)
	pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,name,preferred_locale,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		userB, "b@test.com", "$2a$10$x", "B", "ru", now, now)

	wsID := "aaaaaaaa-ac01-0000-0000-000000000001"
	pool.Exec(ctx, `INSERT INTO workspaces (id,name,owner_id,created_at) VALUES ($1,$2,$3,$4)`, wsID, "WS", userA, now)
	pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id,user_id,role,invited_at,joined_at) VALUES ($1,$2,$3,$4,$4)`, wsID, userA, "admin", now)

	projID := "aaaaaaaa-ac02-0000-0000-000000000001"
	pool.Exec(ctx, `INSERT INTO projects (id,workspace_id,name,description,status,created_by,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		projID, wsID, "P", "", "active", userA, now, now)

	checker := NewPostgresMembershipChecker(pool)

	t.Run("NoSharedProject", func(t *testing.T) {
		shared, err := checker.ShareProject(ctx, userA, userB)
		if err != nil { t.Fatalf("ShareProject: %v", err) }
		if shared { t.Error("expected false") }
	})

	t.Run("SharedProject", func(t *testing.T) {
		pool.Exec(ctx, `INSERT INTO project_members (project_id,user_id,role,added_at) VALUES ($1,$2,$3,$4)`, projID, userA, "admin", now)
		pool.Exec(ctx, `INSERT INTO project_members (project_id,user_id,role,added_at) VALUES ($1,$2,$3,$4)`, projID, userB, "developer", now)
		shared, err := checker.ShareProject(ctx, userA, userB)
		if err != nil { t.Fatalf("ShareProject: %v", err) }
		if !shared { t.Error("expected true") }
	})
}
