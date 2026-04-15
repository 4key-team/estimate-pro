package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VDV001/estimate-pro/backend/internal/modules/document/domain"
	"github.com/VDV001/estimate-pro/backend/internal/modules/testutil"
)

// helpers

func createDocTestUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.New().String()
	now := time.Now()
	_, err := pool.Exec(t.Context(),
		`INSERT INTO users (id, email, password_hash, name, preferred_locale, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, uuid.New().String()+"@test.com", "$2a$10$xx", "Doc User", "ru", now, now)
	if err != nil {
		t.Fatalf("createDocTestUser: %v", err)
	}
	return id
}

func createDocTestProject(t *testing.T, pool *pgxpool.Pool, ownerID string) string {
	t.Helper()
	wsID := uuid.New().String()
	projID := uuid.New().String()
	now := time.Now()
	_, err := pool.Exec(t.Context(),
		`INSERT INTO workspaces (id, name, owner_id, created_at) VALUES ($1,$2,$3,$4)`,
		wsID, "ws", ownerID, now)
	if err != nil {
		t.Fatalf("createDocTestProject ws: %v", err)
	}
	_, err = pool.Exec(t.Context(),
		`INSERT INTO projects (id, workspace_id, name, status, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		projID, wsID, "doc-proj", "active", ownerID, now, now)
	if err != nil {
		t.Fatalf("createDocTestProject: %v", err)
	}
	return projID
}

func createDocTestDocument(t *testing.T, pool *pgxpool.Pool, projID, userID string) string {
	t.Helper()
	docID := uuid.New().String()
	_, err := pool.Exec(t.Context(),
		`INSERT INTO documents (id, project_id, title, uploaded_by, created_at) VALUES ($1,$2,$3,$4,$5)`,
		docID, projID, "Test Doc", userID, time.Now())
	if err != nil {
		t.Fatalf("createDocTestDocument: %v", err)
	}
	return docID
}

// ---------- Document Repository ----------

func TestDocumentRepository_CreateAndGetByID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresDocumentRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)

	doc := &domain.Document{
		ID:         uuid.New().String(),
		ProjectID:  projID,
		Title:      "Architecture Doc",
		UploadedBy: userID,
		CreatedAt:  time.Now().Truncate(time.Microsecond),
	}

	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Architecture Doc" {
		t.Errorf("title: got %q, want %q", got.Title, "Architecture Doc")
	}
}

func TestDocumentRepository_GetByID_NotFound(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresDocumentRepository(pool)

	_, err := repo.GetByID(ctx, uuid.New().String())
	if !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("got %v, want %v", err, domain.ErrDocumentNotFound)
	}
}

func TestDocumentRepository_ListByProject(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresDocumentRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)

	for i := range 3 {
		doc := &domain.Document{
			ID:         uuid.New().String(),
			ProjectID:  projID,
			Title:      "Doc " + uuid.New().String()[:4],
			UploadedBy: userID,
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := repo.Create(ctx, doc); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	docs, err := repo.ListByProject(ctx, projID)
	if err != nil {
		t.Fatalf("ListByProject: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("got %d, want 3", len(docs))
	}
}

func TestDocumentRepository_Delete(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresDocumentRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)

	doc := &domain.Document{
		ID:         uuid.New().String(),
		ProjectID:  projID,
		Title:      "To Delete",
		UploadedBy: userID,
		CreatedAt:  time.Now(),
	}
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Delete(ctx, doc.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := repo.GetByID(ctx, doc.ID)
	if !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("after delete: got %v, want %v", err, domain.ErrDocumentNotFound)
	}
}

// ---------- Version Repository ----------

func TestVersionRepository_CreateAndGetByID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresVersionRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)
	docID := createDocTestDocument(t, pool, projID, userID)

	v := &domain.DocumentVersion{
		ID:              uuid.New().String(),
		DocumentID:      docID,
		VersionNumber:   1,
		FileKey:         "docs/test.pdf",
		FileType:        domain.FileTypePDF,
		FileSize:        1024,
		ParsedStatus:    domain.ParsedStatusPending,
		ConfidenceScore: 0.0,
		UploadedBy:      userID,
		UploadedAt:      time.Now().Truncate(time.Microsecond),
	}

	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, v.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.FileKey != "docs/test.pdf" {
		t.Errorf("file_key: got %q, want %q", got.FileKey, "docs/test.pdf")
	}
	if got.FileType != domain.FileTypePDF {
		t.Errorf("file_type: got %q, want %q", got.FileType, domain.FileTypePDF)
	}
}

func TestVersionRepository_GetByID_NotFound(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresVersionRepository(pool)

	_, err := repo.GetByID(ctx, uuid.New().String())
	if !errors.Is(err, domain.ErrVersionNotFound) {
		t.Errorf("got %v, want %v", err, domain.ErrVersionNotFound)
	}
}

func TestVersionRepository_ListByDocument(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresVersionRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)
	docID := createDocTestDocument(t, pool, projID, userID)

	for i := 1; i <= 3; i++ {
		v := &domain.DocumentVersion{
			ID:            uuid.New().String(),
			DocumentID:    docID,
			VersionNumber: i,
			FileKey:       "docs/v" + uuid.New().String()[:4],
			FileType:      domain.FileTypePDF,
			FileSize:      int64(1024 * i),
			ParsedStatus:  domain.ParsedStatusPending,
			UploadedBy:    userID,
			UploadedAt:    time.Now(),
		}
		if err := repo.Create(ctx, v); err != nil {
			t.Fatalf("Create v%d: %v", i, err)
		}
	}

	versions, err := repo.ListByDocument(ctx, docID)
	if err != nil {
		t.Fatalf("ListByDocument: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("got %d versions, want 3", len(versions))
	}
	// Should be sorted by version_number DESC.
	if versions[0].VersionNumber != 3 {
		t.Errorf("first version_number: got %d, want 3", versions[0].VersionNumber)
	}
}

func TestVersionRepository_GetLatestByDocument(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresVersionRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)
	docID := createDocTestDocument(t, pool, projID, userID)

	for i := 1; i <= 2; i++ {
		v := &domain.DocumentVersion{
			ID:            uuid.New().String(),
			DocumentID:    docID,
			VersionNumber: i,
			FileKey:       "docs/v" + uuid.New().String()[:4],
			FileType:      domain.FileTypePDF,
			FileSize:      int64(1024 * i),
			ParsedStatus:  domain.ParsedStatusPending,
			UploadedBy:    userID,
			UploadedAt:    time.Now(),
		}
		if err := repo.Create(ctx, v); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	latest, err := repo.GetLatestByDocument(ctx, docID)
	if err != nil {
		t.Fatalf("GetLatestByDocument: %v", err)
	}
	if latest.VersionNumber != 2 {
		t.Errorf("version_number: got %d, want 2", latest.VersionNumber)
	}
}

func TestVersionRepository_UpdateFlags(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresVersionRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)
	docID := createDocTestDocument(t, pool, projID, userID)

	v := &domain.DocumentVersion{
		ID:            uuid.New().String(),
		DocumentID:    docID,
		VersionNumber: 1,
		FileKey:       "docs/flags.pdf",
		FileType:      domain.FileTypePDF,
		FileSize:      512,
		ParsedStatus:  domain.ParsedStatusPending,
		UploadedBy:    userID,
		UploadedAt:    time.Now(),
	}
	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateFlags(ctx, v.ID, true, true); err != nil {
		t.Fatalf("UpdateFlags: %v", err)
	}

	got, _ := repo.GetByID(ctx, v.ID)
	if !got.IsSigned {
		t.Error("expected is_signed=true")
	}
	if !got.IsFinal {
		t.Error("expected is_final=true")
	}
}

func TestVersionRepository_ClearFinal(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresVersionRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)
	docID := createDocTestDocument(t, pool, projID, userID)

	v := &domain.DocumentVersion{
		ID:            uuid.New().String(),
		DocumentID:    docID,
		VersionNumber: 1,
		FileKey:       "docs/final.pdf",
		FileType:      domain.FileTypePDF,
		FileSize:      256,
		ParsedStatus:  domain.ParsedStatusPending,
		UploadedBy:    userID,
		UploadedAt:    time.Now(),
	}
	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.UpdateFlags(ctx, v.ID, false, true); err != nil {
		t.Fatalf("UpdateFlags: %v", err)
	}

	if err := repo.ClearFinal(ctx, docID); err != nil {
		t.Fatalf("ClearFinal: %v", err)
	}

	got, _ := repo.GetByID(ctx, v.ID)
	if got.IsFinal {
		t.Error("expected is_final=false after ClearFinal")
	}
}

func TestVersionRepository_SetAndGetTags(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	ctx := t.Context()
	repo := NewPostgresVersionRepository(pool)

	userID := createDocTestUser(t, pool)
	projID := createDocTestProject(t, pool, userID)
	docID := createDocTestDocument(t, pool, projID, userID)

	v := &domain.DocumentVersion{
		ID:            uuid.New().String(),
		DocumentID:    docID,
		VersionNumber: 1,
		FileKey:       "docs/tagged.pdf",
		FileType:      domain.FileTypePDF,
		FileSize:      128,
		ParsedStatus:  domain.ParsedStatusPending,
		UploadedBy:    userID,
		UploadedAt:    time.Now(),
	}
	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tags := []string{"срочно", "черновик"}
	if err := repo.SetTags(ctx, v.ID, tags); err != nil {
		t.Fatalf("SetTags: %v", err)
	}

	got, err := repo.GetTags(ctx, v.ID)
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d tags, want 2", len(got))
	}

	// Replace tags.
	if err := repo.SetTags(ctx, v.ID, []string{"архив"}); err != nil {
		t.Fatalf("SetTags replace: %v", err)
	}
	got, _ = repo.GetTags(ctx, v.ID)
	if len(got) != 1 || got[0] != "архив" {
		t.Errorf("after replace: got %v, want [архив]", got)
	}
}
