package repo

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VDV001/estimate-pro/backend/internal/modules/estimation/domain"
	"github.com/VDV001/estimate-pro/backend/internal/modules/testutil"
)

// helpers

func createEstTestUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.New().String()
	now := time.Now()
	_, err := pool.Exec(t.Context(),
		`INSERT INTO users (id, email, password_hash, name, preferred_locale, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, uuid.New().String()+"@test.com", "$2a$10$xx", "Est User", "ru", now, now)
	if err != nil {
		t.Fatalf("createEstTestUser: %v", err)
	}
	return id
}

func createEstTestProject(t *testing.T, pool *pgxpool.Pool, ownerID string) string {
	t.Helper()
	wsID := uuid.New().String()
	projID := uuid.New().String()
	now := time.Now()
	_, err := pool.Exec(t.Context(),
		`INSERT INTO workspaces (id, name, owner_id, created_at) VALUES ($1,$2,$3,$4)`,
		wsID, "ws", ownerID, now)
	if err != nil {
		t.Fatalf("createEstTestProject ws: %v", err)
	}
	_, err = pool.Exec(t.Context(),
		`INSERT INTO projects (id, workspace_id, name, status, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		projID, wsID, "est-proj", "active", ownerID, now, now)
	if err != nil {
		t.Fatalf("createEstTestProject: %v", err)
	}
	return projID
}

// ---------- Estimation Repository ----------

func TestEstimationRepository_CreateAndGetByID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresEstimationRepository(pool)

	userID := createEstTestUser(t, pool)
	projID := createEstTestProject(t, pool, userID)

	est := &domain.Estimation{
		ID:          uuid.New().String(),
		ProjectID:   projID,
		SubmittedBy: userID,
		Status:      domain.StatusDraft,
		CreatedAt:   time.Now().Truncate(time.Microsecond),
	}

	if err := repo.Create(ctx, est); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, est.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Status != domain.StatusDraft {
		t.Errorf("status: got %q, want %q", got.Status, domain.StatusDraft)
	}
	if got.ProjectID != projID {
		t.Errorf("project_id: got %q, want %q", got.ProjectID, projID)
	}
}

func TestEstimationRepository_GetByID_NotFound(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresEstimationRepository(pool)

	_, err := repo.GetByID(ctx, uuid.New().String())
	if !errors.Is(err, domain.ErrEstimationNotFound) {
		t.Errorf("got %v, want %v", err, domain.ErrEstimationNotFound)
	}
}

func TestEstimationRepository_ListByProject(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresEstimationRepository(pool)

	userID := createEstTestUser(t, pool)
	projID := createEstTestProject(t, pool, userID)

	for i := range 3 {
		est := &domain.Estimation{
			ID:          uuid.New().String(),
			ProjectID:   projID,
			SubmittedBy: userID,
			Status:      domain.StatusDraft,
			CreatedAt:   time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := repo.Create(ctx, est); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	list, err := repo.ListByProject(ctx, projID)
	if err != nil {
		t.Fatalf("ListByProject: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("got %d, want 3", len(list))
	}
}

func TestEstimationRepository_UpdateStatus(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresEstimationRepository(pool)

	userID := createEstTestUser(t, pool)
	projID := createEstTestProject(t, pool, userID)

	est := &domain.Estimation{
		ID:          uuid.New().String(),
		ProjectID:   projID,
		SubmittedBy: userID,
		Status:      domain.StatusDraft,
		CreatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, est); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, est.ID, domain.StatusSubmitted); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, _ := repo.GetByID(ctx, est.ID)
	if got.Status != domain.StatusSubmitted {
		t.Errorf("status: got %q, want %q", got.Status, domain.StatusSubmitted)
	}
	if got.SubmittedAt.IsZero() {
		t.Error("submitted_at should be set after submit")
	}
}

func TestEstimationRepository_UpdateStatus_NotFound(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresEstimationRepository(pool)

	err := repo.UpdateStatus(ctx, uuid.New().String(), domain.StatusSubmitted)
	if !errors.Is(err, domain.ErrEstimationNotFound) {
		t.Errorf("got %v, want %v", err, domain.ErrEstimationNotFound)
	}
}

func TestEstimationRepository_Delete(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresEstimationRepository(pool)

	userID := createEstTestUser(t, pool)
	projID := createEstTestProject(t, pool, userID)

	est := &domain.Estimation{
		ID:          uuid.New().String(),
		ProjectID:   projID,
		SubmittedBy: userID,
		Status:      domain.StatusDraft,
		CreatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, est); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Delete(ctx, est.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := repo.GetByID(ctx, est.ID)
	if !errors.Is(err, domain.ErrEstimationNotFound) {
		t.Errorf("after delete: got %v, want %v", err, domain.ErrEstimationNotFound)
	}
}

// ---------- Item Repository ----------

func TestItemRepository_CreateBatchAndList(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	estRepo := NewPostgresEstimationRepository(pool)
	itemRepo := NewPostgresItemRepository(pool)

	userID := createEstTestUser(t, pool)
	projID := createEstTestProject(t, pool, userID)

	est := &domain.Estimation{
		ID:          uuid.New().String(),
		ProjectID:   projID,
		SubmittedBy: userID,
		Status:      domain.StatusDraft,
		CreatedAt:   time.Now(),
	}
	if err := estRepo.Create(ctx, est); err != nil {
		t.Fatalf("Create estimation: %v", err)
	}

	items := []*domain.EstimationItem{
		{ID: uuid.New().String(), EstimationID: est.ID, TaskName: "Backend API", MinHours: 8, LikelyHours: 16, MaxHours: 32, SortOrder: 0, Note: "REST endpoints"},
		{ID: uuid.New().String(), EstimationID: est.ID, TaskName: "Frontend UI", MinHours: 4, LikelyHours: 8, MaxHours: 16, SortOrder: 1, Note: ""},
		{ID: uuid.New().String(), EstimationID: est.ID, TaskName: "Testing", MinHours: 2, LikelyHours: 4, MaxHours: 8, SortOrder: 2, Note: "unit + integration"},
	}

	if err := itemRepo.CreateBatch(ctx, items); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	got, err := itemRepo.ListByEstimation(ctx, est.ID)
	if err != nil {
		t.Fatalf("ListByEstimation: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d items, want 3", len(got))
	}
	// Should be sorted by sort_order.
	if got[0].TaskName != "Backend API" {
		t.Errorf("first item: got %q, want %q", got[0].TaskName, "Backend API")
	}
	if got[2].TaskName != "Testing" {
		t.Errorf("last item: got %q, want %q", got[2].TaskName, "Testing")
	}
	// Verify PERT values round-trip.
	if got[0].MinHours != 8 || got[0].LikelyHours != 16 || got[0].MaxHours != 32 {
		t.Errorf("hours mismatch: min=%v likely=%v max=%v", got[0].MinHours, got[0].LikelyHours, got[0].MaxHours)
	}
}

func TestItemRepository_CreateBatch_Empty(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	itemRepo := NewPostgresItemRepository(pool)

	// Empty batch should be a no-op.
	if err := itemRepo.CreateBatch(ctx, nil); err != nil {
		t.Fatalf("CreateBatch empty: %v", err)
	}
}

func TestItemRepository_DeleteByEstimation(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	estRepo := NewPostgresEstimationRepository(pool)
	itemRepo := NewPostgresItemRepository(pool)

	userID := createEstTestUser(t, pool)
	projID := createEstTestProject(t, pool, userID)

	est := &domain.Estimation{
		ID:          uuid.New().String(),
		ProjectID:   projID,
		SubmittedBy: userID,
		Status:      domain.StatusDraft,
		CreatedAt:   time.Now(),
	}
	if err := estRepo.Create(ctx, est); err != nil {
		t.Fatalf("Create: %v", err)
	}

	items := []*domain.EstimationItem{
		{ID: uuid.New().String(), EstimationID: est.ID, TaskName: "Task", MinHours: 1, LikelyHours: 2, MaxHours: 3, SortOrder: 0},
	}
	if err := itemRepo.CreateBatch(ctx, items); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	if err := itemRepo.DeleteByEstimation(ctx, est.ID); err != nil {
		t.Fatalf("DeleteByEstimation: %v", err)
	}

	got, err := itemRepo.ListByEstimation(ctx, est.ID)
	if err != nil {
		t.Fatalf("ListByEstimation: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d items after delete, want 0", len(got))
	}
}
