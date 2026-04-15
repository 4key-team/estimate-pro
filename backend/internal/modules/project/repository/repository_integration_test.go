package repository

import (
	"testing"
	"time"

	projectdomain "github.com/VDV001/estimate-pro/backend/internal/modules/project/domain"
	"github.com/VDV001/estimate-pro/backend/internal/testutil"
)

func TestProjectRepository_Integration(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	now := time.Now().Truncate(time.Microsecond)

	// Seed user
	userID := "bbbbbbbb-0000-0000-0000-000000000001"
	pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,name,preferred_locale,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		userID, "proj@test.com", "$2a$10$abc", "Proj Tester", "ru", now, now)

	// Workspace
	wsRepo := NewPostgresWorkspaceRepository(pool)
	ws := &projectdomain.Workspace{ID: "cccccccc-0000-0000-0000-000000000001", Name: "WS", OwnerID: userID, CreatedAt: now}
	wsRepo.Create(ctx, ws)

	projRepo := NewPostgresProjectRepository(pool)
	proj := &projectdomain.Project{
		ID: "dddddddd-0000-0000-0000-000000000001", WorkspaceID: ws.ID, Name: "Proj",
		Description: "Desc", Status: "active", CreatedBy: userID, CreatedAt: now, UpdatedAt: now,
	}

	t.Run("Create_GetByID", func(t *testing.T) {
		if err := projRepo.Create(ctx, proj); err != nil {
			t.Fatalf("Create: %v", err)
		}
		got, err := projRepo.GetByID(ctx, proj.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Proj" {
			t.Errorf("name = %q, want Proj", got.Name)
		}
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		_, err := projRepo.GetByID(ctx, "dddddddd-0000-0000-0000-999999999999")
		if err != projectdomain.ErrProjectNotFound {
			t.Errorf("err = %v, want ErrProjectNotFound", err)
		}
	})

	t.Run("ListByWorkspace", func(t *testing.T) {
		list, total, err := projRepo.ListByWorkspace(ctx, ws.ID, 10, 0)
		if err != nil {
			t.Fatalf("ListByWorkspace: %v", err)
		}
		if total != 1 || len(list) != 1 {
			t.Errorf("total=%d len=%d, want 1/1", total, len(list))
		}
	})

	t.Run("Update", func(t *testing.T) {
		proj.Name = "Updated"
		proj.UpdatedAt = time.Now().Truncate(time.Microsecond)
		projRepo.Update(ctx, proj)
		got, _ := projRepo.GetByID(ctx, proj.ID)
		if got.Name != "Updated" {
			t.Errorf("name = %q, want Updated", got.Name)
		}
	})

	// Members
	memberRepo := NewPostgresMemberRepository(pool)
	t.Run("AddMember_GetRole_List", func(t *testing.T) {
		m := &projectdomain.Member{ProjectID: proj.ID, UserID: userID, Role: "admin", AddedAt: now}
		if err := memberRepo.Add(ctx, m); err != nil {
			t.Fatalf("Add: %v", err)
		}
		role, err := memberRepo.GetRole(ctx, proj.ID, userID)
		if err != nil {
			t.Fatalf("GetRole: %v", err)
		}
		if role != "admin" {
			t.Errorf("role = %q, want admin", role)
		}
		members, _ := memberRepo.ListByProject(ctx, proj.ID)
		if len(members) != 1 {
			t.Errorf("members len = %d, want 1", len(members))
		}
	})

	t.Run("ListByProjectWithUsers", func(t *testing.T) {
		members, err := memberRepo.ListByProjectWithUsers(ctx, proj.ID)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(members) != 1 || members[0].UserName != "Proj Tester" {
			t.Errorf("unexpected: %+v", members)
		}
	})

	t.Run("ListByUser", func(t *testing.T) {
		list, total, _ := projRepo.ListByUser(ctx, userID, 10, 0)
		if total != 1 || len(list) != 1 {
			t.Errorf("total=%d len=%d, want 1/1", total, len(list))
		}
	})

	t.Run("RemoveMember", func(t *testing.T) {
		memberRepo.Remove(ctx, proj.ID, userID)
		_, err := memberRepo.GetRole(ctx, proj.ID, userID)
		if err != projectdomain.ErrMemberNotFound {
			t.Errorf("err = %v, want ErrMemberNotFound", err)
		}
	})

	t.Run("WorkspaceUpdate", func(t *testing.T) {
		ws.Name = "Renamed"
		wsRepo.Update(ctx, ws)
		got, _ := wsRepo.GetByID(ctx, ws.ID)
		if got.Name != "Renamed" {
			t.Errorf("name = %q, want Renamed", got.Name)
		}
	})

	t.Run("WorkspaceUpdate_NotOwner", func(t *testing.T) {
		bad := &projectdomain.Workspace{ID: ws.ID, Name: "Hack", OwnerID: "eeeeeeee-0000-0000-0000-999999999999"}
		err := wsRepo.Update(ctx, bad)
		if err != projectdomain.ErrWorkspaceNotFound {
			t.Errorf("err = %v, want ErrWorkspaceNotFound", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		projRepo.Delete(ctx, proj.ID)
		_, err := projRepo.GetByID(ctx, proj.ID)
		if err != projectdomain.ErrProjectNotFound {
			t.Errorf("after delete: %v", err)
		}
	})
}
