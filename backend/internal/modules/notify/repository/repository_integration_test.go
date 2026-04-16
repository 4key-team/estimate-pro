package repository

import (
	"testing"
	"time"

	"github.com/VDV001/estimate-pro/backend/internal/modules/notify/domain"
	"github.com/VDV001/estimate-pro/backend/internal/testutil"
)

func TestNotifyRepository_Integration(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	now := time.Now().Truncate(time.Microsecond)

	// Seed user
	userID := "ffffffff-0000-0000-0000-000000000001"
	pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,name,preferred_locale,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		userID, "notify@test.com", "$2a$10$abc", "Notify Tester", "ru", now, now)

	// Seed workspace + project for FK
	wsID := "ffffffff-1111-0000-0000-000000000001"
	pool.Exec(ctx, `INSERT INTO workspaces (id,name,owner_id,created_at) VALUES ($1,$2,$3,$4)`, wsID, "WS", userID, now)
	pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id,user_id,role,invited_at,joined_at) VALUES ($1,$2,$3,$4,$4)`, wsID, userID, "admin", now)
	projID := "ffffffff-2222-0000-0000-000000000001"
	pool.Exec(ctx, `INSERT INTO projects (id,workspace_id,name,description,status,created_by,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		projID, wsID, "Proj", "", "active", userID, now, now)

	notifRepo := NewPostgresNotificationRepository(pool)
	prefRepo := NewPostgresPreferenceRepository(pool)
	logRepo := NewPostgresDeliveryLogRepository(pool)

	n1 := &domain.Notification{
		ID: "aaaaaaaa-1111-0000-0000-000000000001", UserID: userID, EventType: "member.added",
		Title: "Added", Message: "You were added", ProjectID: projID, Read: false, CreatedAt: now,
	}
	n2 := &domain.Notification{
		ID: "aaaaaaaa-1111-0000-0000-000000000002", UserID: userID, EventType: "estimation.submitted",
		Title: "Submitted", Message: "Estimation submitted", ProjectID: projID, Read: false, CreatedAt: now.Add(time.Second),
	}

	t.Run("Create", func(t *testing.T) {
		if err := notifRepo.Create(ctx, n1); err != nil {
			t.Fatalf("Create: %v", err)
		}
	})

	t.Run("CreateBatch", func(t *testing.T) {
		if err := notifRepo.CreateBatch(ctx, []*domain.Notification{n2}); err != nil {
			t.Fatalf("CreateBatch: %v", err)
		}
	})

	t.Run("ListByUser", func(t *testing.T) {
		list, total, err := notifRepo.ListByUser(ctx, userID, 10, 0)
		if err != nil { t.Fatalf("ListByUser: %v", err) }
		if total != 2 { t.Errorf("total = %d, want 2", total) }
		if len(list) != 2 { t.Errorf("len = %d, want 2", len(list)) }
	})

	t.Run("CountUnread", func(t *testing.T) {
		count, err := notifRepo.CountUnread(ctx, userID)
		if err != nil { t.Fatalf("CountUnread: %v", err) }
		if count != 2 { t.Errorf("count = %d, want 2", count) }
	})

	t.Run("MarkRead", func(t *testing.T) {
		if err := notifRepo.MarkRead(ctx, userID, n1.ID); err != nil {
			t.Fatalf("MarkRead: %v", err)
		}
		count, _ := notifRepo.CountUnread(ctx, userID)
		if count != 1 { t.Errorf("unread = %d, want 1", count) }
	})

	t.Run("MarkRead_NotFound", func(t *testing.T) {
		err := notifRepo.MarkRead(ctx, userID, "aaaaaaaa-1111-0000-0000-999999999999")
		if err != domain.ErrNotificationNotFound {
			t.Errorf("err = %v, want ErrNotificationNotFound", err)
		}
	})

	t.Run("MarkAllRead", func(t *testing.T) {
		notifRepo.MarkAllRead(ctx, userID)
		count, _ := notifRepo.CountUnread(ctx, userID)
		if count != 0 { t.Errorf("unread = %d, want 0", count) }
	})

	// Preferences
	t.Run("Upsert_and_Get", func(t *testing.T) {
		pref := &domain.Preference{UserID: userID, Channel: "email", Enabled: true}
		if err := prefRepo.Upsert(ctx, pref); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
		prefs, err := prefRepo.Get(ctx, userID)
		if err != nil { t.Fatalf("Get: %v", err) }
		if len(prefs) != 1 { t.Errorf("len = %d, want 1", len(prefs)) }
		if prefs[0].Channel != "email" || !prefs[0].Enabled {
			t.Errorf("pref = %+v", prefs[0])
		}
	})

	t.Run("Upsert_Update", func(t *testing.T) {
		pref := &domain.Preference{UserID: userID, Channel: "email", Enabled: false}
		prefRepo.Upsert(ctx, pref)
		prefs, _ := prefRepo.Get(ctx, userID)
		if prefs[0].Enabled { t.Error("expected disabled") }
	})

	// DeliveryLog
	t.Run("DeliveryLog_Create", func(t *testing.T) {
		dl := &domain.DeliveryLog{
			ID: "bbbbbbbb-1111-0000-0000-000000000001", UserID: userID,
			EventType: "member.added", Channel: "email", SentAt: now, Status: "sent",
		}
		if err := logRepo.Create(ctx, dl); err != nil {
			t.Fatalf("Create: %v", err)
		}
	})

	// Lookups
	t.Run("EmailLookup", func(t *testing.T) {
		lookup := NewEmailLookup(pool)
		email, err := lookup.GetEmail(ctx, userID)
		if err != nil { t.Fatalf("GetEmail: %v", err) }
		if email != "notify@test.com" { t.Errorf("email = %q", email) }
	})

	t.Run("TelegramChatLookup_NoChat", func(t *testing.T) {
		lookup := NewTelegramChatLookup(pool)
		_, err := lookup.GetTelegramChatID(ctx, userID)
		if err == nil { t.Error("expected error for user without telegram") }
	})

	t.Run("TelegramChatLookup_WithChat", func(t *testing.T) {
		pool.Exec(ctx, `UPDATE users SET telegram_chat_id = '99999' WHERE id = $1`, userID)
		lookup := NewTelegramChatLookup(pool)
		chatID, err := lookup.GetTelegramChatID(ctx, userID)
		if err != nil { t.Fatalf("GetTelegramChatID: %v", err) }
		if chatID != "99999" { t.Errorf("chatID = %q", chatID) }
	})

	t.Run("UserNameLookup", func(t *testing.T) {
		lookup := NewUserNameLookup(pool)
		name, err := lookup.GetName(ctx, userID)
		if err != nil { t.Fatalf("GetName: %v", err) }
		if name != "Notify Tester" { t.Errorf("name = %q", name) }
	})

	t.Run("MemberLister", func(t *testing.T) {
		pool.Exec(ctx, `INSERT INTO project_members (project_id,user_id,role,added_at) VALUES ($1,$2,$3,$4)`, projID, userID, "admin", now)
		lister := NewMemberListerAdapter(pool)
		ids, err := lister.ListMemberUserIDs(ctx, projID)
		if err != nil { t.Fatalf("ListMemberUserIDs: %v", err) }
		if len(ids) != 1 || ids[0] != userID { t.Errorf("ids = %v", ids) }
	})
}
