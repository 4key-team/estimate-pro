package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/VDV001/estimate-pro/backend/internal/modules/document/domain"
)

// --- Inline mocks ---

type mockDocumentRepo struct {
	createFn        func(ctx context.Context, doc *domain.Document) error
	getByIDFn       func(ctx context.Context, id string) (*domain.Document, error)
	listByProjectFn func(ctx context.Context, projectID string) ([]*domain.Document, error)
	deleteFn        func(ctx context.Context, id string) error
}

func (m *mockDocumentRepo) Create(ctx context.Context, doc *domain.Document) error {
	if m.createFn != nil {
		return m.createFn(ctx, doc)
	}
	return nil
}

func (m *mockDocumentRepo) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockDocumentRepo) ListByProject(ctx context.Context, projectID string) ([]*domain.Document, error) {
	if m.listByProjectFn != nil {
		return m.listByProjectFn(ctx, projectID)
	}
	return nil, nil
}

func (m *mockDocumentRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockVersionRepo struct {
	createFn              func(ctx context.Context, version *domain.DocumentVersion) error
	getByIDFn             func(ctx context.Context, id string) (*domain.DocumentVersion, error)
	listByDocumentFn      func(ctx context.Context, documentID string) ([]*domain.DocumentVersion, error)
	getLatestByDocumentFn func(ctx context.Context, documentID string) (*domain.DocumentVersion, error)
	updateFlagsFn         func(ctx context.Context, id string, isSigned, isFinal bool) error
	clearFinalFn          func(ctx context.Context, documentID string) error
	clearFinalByProjectFn func(ctx context.Context, projectID string) error
	setTagsFn             func(ctx context.Context, versionID string, tags []string) error
	getTagsFn             func(ctx context.Context, versionID string) ([]string, error)
}

func (m *mockVersionRepo) Create(ctx context.Context, version *domain.DocumentVersion) error {
	if m.createFn != nil {
		return m.createFn(ctx, version)
	}
	return nil
}

func (m *mockVersionRepo) GetByID(ctx context.Context, id string) (*domain.DocumentVersion, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockVersionRepo) ListByDocument(ctx context.Context, documentID string) ([]*domain.DocumentVersion, error) {
	if m.listByDocumentFn != nil {
		return m.listByDocumentFn(ctx, documentID)
	}
	return nil, nil
}

func (m *mockVersionRepo) GetLatestByDocument(ctx context.Context, documentID string) (*domain.DocumentVersion, error) {
	if m.getLatestByDocumentFn != nil {
		return m.getLatestByDocumentFn(ctx, documentID)
	}
	return nil, nil
}

func (m *mockVersionRepo) UpdateFlags(ctx context.Context, id string, isSigned, isFinal bool) error {
	if m.updateFlagsFn != nil {
		return m.updateFlagsFn(ctx, id, isSigned, isFinal)
	}
	return nil
}

func (m *mockVersionRepo) ClearFinal(ctx context.Context, documentID string) error {
	if m.clearFinalFn != nil {
		return m.clearFinalFn(ctx, documentID)
	}
	return nil
}

func (m *mockVersionRepo) ClearFinalByProject(ctx context.Context, projectID string) error {
	if m.clearFinalByProjectFn != nil {
		return m.clearFinalByProjectFn(ctx, projectID)
	}
	return nil
}

func (m *mockVersionRepo) SetTags(ctx context.Context, versionID string, tags []string) error {
	if m.setTagsFn != nil {
		return m.setTagsFn(ctx, versionID, tags)
	}
	return nil
}

func (m *mockVersionRepo) GetTags(ctx context.Context, versionID string) ([]string, error) {
	if m.getTagsFn != nil {
		return m.getTagsFn(ctx, versionID)
	}
	return nil, nil
}

type mockFileStorage struct {
	uploadFn   func(ctx context.Context, key string, data io.Reader, size int64, contentType string) error
	downloadFn func(ctx context.Context, key string) (io.ReadCloser, error)
	deleteFn   func(ctx context.Context, key string) error
}

func (m *mockFileStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, key, data, size, contentType)
	}
	return nil
}

func (m *mockFileStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, key)
	}
	return nil, nil
}

func (m *mockFileStorage) Delete(ctx context.Context, key string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, key)
	}
	return nil
}

// --- Tests ---

func TestUpload_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    UploadInput
		fileType domain.FileType
	}{
		{
			name: "valid PDF upload",
			input: UploadInput{
				ProjectID: "proj-1",
				Title:     "Requirements",
				FileName:  "requirements.pdf",
				FileSize:  1024,
				FileType:  domain.FileTypePDF,
				Content:   strings.NewReader("fake content"),
				UserID:    "user-1",
			},
			fileType: domain.FileTypePDF,
		},
		{
			name: "valid DOCX upload",
			input: UploadInput{
				ProjectID: "proj-2",
				Title:     "Spec",
				FileName:  "spec.docx",
				FileSize:  2048,
				FileType:  domain.FileTypeDOCX,
				Content:   strings.NewReader("fake content"),
				UserID:    "user-2",
			},
			fileType: domain.FileTypeDOCX,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var createdDoc *domain.Document
			var createdVersion *domain.DocumentVersion
			var uploadedKey string

			docRepo := &mockDocumentRepo{
				createFn: func(_ context.Context, doc *domain.Document) error {
					createdDoc = doc
					return nil
				},
			}
			versionRepo := &mockVersionRepo{
				createFn: func(_ context.Context, v *domain.DocumentVersion) error {
					createdVersion = v
					return nil
				},
			}
			storage := &mockFileStorage{
				uploadFn: func(_ context.Context, key string, _ io.Reader, _ int64, _ string) error {
					uploadedKey = key
					return nil
				},
			}

			uc := New(docRepo, versionRepo, storage)
			doc, version, err := uc.Upload(t.Context(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if doc == nil {
				t.Fatal("expected document, got nil")
			}
			if version == nil {
				t.Fatal("expected version, got nil")
			}
			if doc.ProjectID != tt.input.ProjectID {
				t.Errorf("ProjectID = %q, want %q", doc.ProjectID, tt.input.ProjectID)
			}
			if doc.Title != tt.input.Title {
				t.Errorf("Title = %q, want %q", doc.Title, tt.input.Title)
			}
			if doc.UploadedBy != tt.input.UserID {
				t.Errorf("UploadedBy = %q, want %q", doc.UploadedBy, tt.input.UserID)
			}
			if version.VersionNumber != 1 {
				t.Errorf("VersionNumber = %d, want 1", version.VersionNumber)
			}
			if version.FileType != tt.fileType {
				t.Errorf("FileType = %q, want %q", version.FileType, tt.fileType)
			}
			if version.FileSize != tt.input.FileSize {
				t.Errorf("FileSize = %d, want %d", version.FileSize, tt.input.FileSize)
			}
			if version.ParsedStatus != domain.ParsedStatusPending {
				t.Errorf("ParsedStatus = %q, want %q", version.ParsedStatus, domain.ParsedStatusPending)
			}
			if version.DocumentID != doc.ID {
				t.Errorf("version.DocumentID = %q, want %q", version.DocumentID, doc.ID)
			}
			if createdDoc == nil {
				t.Fatal("docRepo.Create was not called")
			}
			if createdVersion == nil {
				t.Fatal("versionRepo.Create was not called")
			}
			if uploadedKey == "" {
				t.Fatal("storage.Upload was not called")
			}
		})
	}
}

func TestUpload_UnsupportedFileType(t *testing.T) {
	tests := []struct {
		name     string
		fileType domain.FileType
	}{
		{name: "exe file", fileType: domain.FileType("exe")},
		{name: "unknown file", fileType: domain.FileType("unknown")},
		{name: "empty file type", fileType: domain.FileType("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := New(&mockDocumentRepo{}, &mockVersionRepo{}, &mockFileStorage{})
			_, _, err := uc.Upload(t.Context(), UploadInput{
				ProjectID: "proj-1",
				Title:     "Bad file",
				FileName:  "virus.exe",
				FileSize:  100,
				FileType:  tt.fileType,
				Content:   strings.NewReader("fake content"),
				UserID:    "user-1",
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, domain.ErrUnsupportedFileType) {
				t.Errorf("error = %v, want %v", err, domain.ErrUnsupportedFileType)
			}
		})
	}
}

func TestUpload_FileTooLarge(t *testing.T) {
	tests := []struct {
		name string
		size int64
	}{
		{name: "exactly over limit", size: domain.MaxFileSize + 1},
		{name: "way over limit", size: 100 << 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := New(&mockDocumentRepo{}, &mockVersionRepo{}, &mockFileStorage{})
			_, _, err := uc.Upload(t.Context(), UploadInput{
				ProjectID: "proj-1",
				Title:     "Huge file",
				FileName:  "huge.pdf",
				FileSize:  tt.size,
				FileType:  domain.FileTypePDF,
				Content:   strings.NewReader("fake content"),
				UserID:    "user-1",
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, domain.ErrFileTooLarge) {
				t.Errorf("error = %v, want %v", err, domain.ErrFileTooLarge)
			}
		})
	}
}

func TestList_Success(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		docs      []*domain.Document
	}{
		{
			name:      "two documents",
			projectID: "proj-1",
			docs: []*domain.Document{
				{ID: "doc-1", ProjectID: "proj-1", Title: "Doc 1"},
				{ID: "doc-2", ProjectID: "proj-1", Title: "Doc 2"},
			},
		},
		{
			name:      "empty list",
			projectID: "proj-empty",
			docs:      []*domain.Document{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docRepo := &mockDocumentRepo{
				listByProjectFn: func(_ context.Context, projectID string) ([]*domain.Document, error) {
					if projectID != tt.projectID {
						t.Errorf("projectID = %q, want %q", projectID, tt.projectID)
					}
					return tt.docs, nil
				},
			}

			uc := New(docRepo, &mockVersionRepo{}, &mockFileStorage{})
			result, err := uc.List(t.Context(), tt.projectID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != len(tt.docs) {
				t.Errorf("len(result) = %d, want %d", len(result), len(tt.docs))
			}
		})
	}
}

func TestGet_Success(t *testing.T) {
	tests := []struct {
		name    string
		docID   string
		doc     *domain.Document
		version *domain.DocumentVersion
	}{
		{
			name:  "document with latest version",
			docID: "doc-1",
			doc:   &domain.Document{ID: "doc-1", ProjectID: "proj-1", Title: "Requirements"},
			version: &domain.DocumentVersion{
				ID:            "ver-1",
				DocumentID:    "doc-1",
				VersionNumber: 3,
				FileType:      domain.FileTypePDF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docRepo := &mockDocumentRepo{
				getByIDFn: func(_ context.Context, id string) (*domain.Document, error) {
					if id != tt.docID {
						t.Errorf("id = %q, want %q", id, tt.docID)
					}
					return tt.doc, nil
				},
			}
			versionRepo := &mockVersionRepo{
				getLatestByDocumentFn: func(_ context.Context, docID string) (*domain.DocumentVersion, error) {
					if docID != tt.docID {
						t.Errorf("docID = %q, want %q", docID, tt.docID)
					}
					return tt.version, nil
				},
			}

			uc := New(docRepo, versionRepo, &mockFileStorage{})
			result, err := uc.Get(t.Context(), tt.docID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Document.ID != tt.doc.ID {
				t.Errorf("Document.ID = %q, want %q", result.Document.ID, tt.doc.ID)
			}
			if result.LatestVersion.ID != tt.version.ID {
				t.Errorf("LatestVersion.ID = %q, want %q", result.LatestVersion.ID, tt.version.ID)
			}
			if result.LatestVersion.VersionNumber != tt.version.VersionNumber {
				t.Errorf("VersionNumber = %d, want %d", result.LatestVersion.VersionNumber, tt.version.VersionNumber)
			}
		})
	}
}

func TestGet_NotFound(t *testing.T) {
	tests := []struct {
		name  string
		docID string
	}{
		{name: "nonexistent document", docID: "doc-nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docRepo := &mockDocumentRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.Document, error) {
					return nil, domain.ErrDocumentNotFound
				},
			}

			uc := New(docRepo, &mockVersionRepo{}, &mockFileStorage{})
			_, err := uc.Get(t.Context(), tt.docID)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, domain.ErrDocumentNotFound) {
				t.Errorf("error = %v, want %v", err, domain.ErrDocumentNotFound)
			}
		})
	}
}

func TestDelete_Success(t *testing.T) {
	tests := []struct {
		name     string
		docID    string
		userID   string
		versions []*domain.DocumentVersion
	}{
		{
			name:   "delete document with two versions",
			docID:  "doc-1",
			userID: "user-1",
			versions: []*domain.DocumentVersion{
				{ID: "ver-1", DocumentID: "doc-1", FileKey: "documents/proj-1/doc-1/ver-1/file.pdf"},
				{ID: "ver-2", DocumentID: "doc-1", FileKey: "documents/proj-1/doc-1/ver-2/file.pdf"},
			},
		},
		{
			name:     "delete document with no versions",
			docID:    "doc-2",
			userID:   "user-1",
			versions: []*domain.DocumentVersion{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deletedKeys []string
			var docDeleted bool

			versionRepo := &mockVersionRepo{
				listByDocumentFn: func(_ context.Context, docID string) ([]*domain.DocumentVersion, error) {
					if docID != tt.docID {
						t.Errorf("docID = %q, want %q", docID, tt.docID)
					}
					return tt.versions, nil
				},
			}
			storage := &mockFileStorage{
				deleteFn: func(_ context.Context, key string) error {
					deletedKeys = append(deletedKeys, key)
					return nil
				},
			}
			docRepo := &mockDocumentRepo{
				getByIDFn: func(_ context.Context, id string) (*domain.Document, error) {
					return &domain.Document{ID: id, UploadedBy: tt.userID}, nil
				},
				deleteFn: func(_ context.Context, id string) error {
					if id != tt.docID {
						t.Errorf("id = %q, want %q", id, tt.docID)
					}
					docDeleted = true
					return nil
				},
			}

			uc := New(docRepo, versionRepo, storage)
			err := uc.Delete(t.Context(), tt.docID, tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(deletedKeys) != len(tt.versions) {
				t.Errorf("deleted %d files, want %d", len(deletedKeys), len(tt.versions))
			}
			for i, v := range tt.versions {
				if deletedKeys[i] != v.FileKey {
					t.Errorf("deletedKeys[%d] = %q, want %q", i, deletedKeys[i], v.FileKey)
				}
			}
			if !docDeleted {
				t.Error("docRepo.Delete was not called")
			}
		})
	}
}

// --- Upload error branches ---

func TestUpload_DocRepoCreateError(t *testing.T) {
	wantErr := errors.New("db down")
	docRepo := &mockDocumentRepo{
		createFn: func(_ context.Context, _ *domain.Document) error { return wantErr },
	}
	uc := New(docRepo, &mockVersionRepo{}, &mockFileStorage{})
	_, _, err := uc.Upload(t.Context(), UploadInput{
		ProjectID: "proj-1", Title: "Doc", FileName: "f.pdf",
		FileSize: 100, FileType: domain.FileTypePDF,
		Content: strings.NewReader("data"), UserID: "u-1",
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestUpload_StorageError(t *testing.T) {
	wantErr := errors.New("s3 unavailable")
	storage := &mockFileStorage{
		uploadFn: func(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error { return wantErr },
	}
	uc := New(&mockDocumentRepo{}, &mockVersionRepo{}, storage)
	_, _, err := uc.Upload(t.Context(), UploadInput{
		ProjectID: "proj-1", Title: "Doc", FileName: "f.pdf",
		FileSize: 100, FileType: domain.FileTypePDF,
		Content: strings.NewReader("data"), UserID: "u-1",
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestUpload_VersionRepoCreateError(t *testing.T) {
	wantErr := errors.New("version insert failed")
	versionRepo := &mockVersionRepo{
		createFn: func(_ context.Context, _ *domain.DocumentVersion) error { return wantErr },
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	_, _, err := uc.Upload(t.Context(), UploadInput{
		ProjectID: "proj-1", Title: "Doc", FileName: "f.pdf",
		FileSize: 100, FileType: domain.FileTypePDF,
		Content: strings.NewReader("data"), UserID: "u-1",
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

// --- List error ---

func TestList_Error(t *testing.T) {
	wantErr := errors.New("db error")
	docRepo := &mockDocumentRepo{
		listByProjectFn: func(_ context.Context, _ string) ([]*domain.Document, error) {
			return nil, wantErr
		},
	}
	uc := New(docRepo, &mockVersionRepo{}, &mockFileStorage{})
	_, err := uc.List(t.Context(), "proj-1")
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

// --- Get latest version error ---

func TestGet_LatestVersionError(t *testing.T) {
	wantErr := errors.New("version error")
	docRepo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Document, error) {
			return &domain.Document{ID: "doc-1"}, nil
		},
	}
	versionRepo := &mockVersionRepo{
		getLatestByDocumentFn: func(_ context.Context, _ string) (*domain.DocumentVersion, error) {
			return nil, wantErr
		},
	}
	uc := New(docRepo, versionRepo, &mockFileStorage{})
	_, err := uc.Get(t.Context(), "doc-1")
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

// --- GetVersions ---

func TestGetVersions_Success(t *testing.T) {
	versions := []*domain.DocumentVersion{
		{ID: "v1", DocumentID: "doc-1", VersionNumber: 1},
		{ID: "v2", DocumentID: "doc-1", VersionNumber: 2},
	}
	versionRepo := &mockVersionRepo{
		listByDocumentFn: func(_ context.Context, docID string) ([]*domain.DocumentVersion, error) {
			if docID != "doc-1" {
				t.Errorf("docID = %q, want %q", docID, "doc-1")
			}
			return versions, nil
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	result, err := uc.GetVersions(t.Context(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}
}

func TestGetVersions_Error(t *testing.T) {
	wantErr := errors.New("db error")
	versionRepo := &mockVersionRepo{
		listByDocumentFn: func(_ context.Context, _ string) ([]*domain.DocumentVersion, error) {
			return nil, wantErr
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	_, err := uc.GetVersions(t.Context(), "doc-1")
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

// --- Download ---

func TestDownload_Success(t *testing.T) {
	versionRepo := &mockVersionRepo{
		getLatestByDocumentFn: func(_ context.Context, _ string) (*domain.DocumentVersion, error) {
			return &domain.DocumentVersion{
				ID: "v1", DocumentID: "doc-1", FileKey: "documents/proj-1/doc-1/v1/file.pdf",
			}, nil
		},
	}
	storage := &mockFileStorage{
		downloadFn: func(_ context.Context, key string) (io.ReadCloser, error) {
			if key != "documents/proj-1/doc-1/v1/file.pdf" {
				t.Errorf("key = %q, want %q", key, "documents/proj-1/doc-1/v1/file.pdf")
			}
			return io.NopCloser(strings.NewReader("pdf content")), nil
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, storage)
	reader, fileKey, err := uc.Download(t.Context(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader == nil {
		t.Fatal("expected reader, got nil")
	}
	defer reader.Close()
	if fileKey != "documents/proj-1/doc-1/v1/file.pdf" {
		t.Errorf("fileKey = %q, want %q", fileKey, "documents/proj-1/doc-1/v1/file.pdf")
	}
}

func TestDownload_VersionNotFound(t *testing.T) {
	versionRepo := &mockVersionRepo{
		getLatestByDocumentFn: func(_ context.Context, _ string) (*domain.DocumentVersion, error) {
			return nil, domain.ErrVersionNotFound
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	_, _, err := uc.Download(t.Context(), "doc-nonexistent")
	if !errors.Is(err, domain.ErrVersionNotFound) {
		t.Errorf("error = %v, want %v", err, domain.ErrVersionNotFound)
	}
}

func TestDownload_StorageError(t *testing.T) {
	wantErr := errors.New("s3 down")
	versionRepo := &mockVersionRepo{
		getLatestByDocumentFn: func(_ context.Context, _ string) (*domain.DocumentVersion, error) {
			return &domain.DocumentVersion{ID: "v1", FileKey: "key"}, nil
		},
	}
	storage := &mockFileStorage{
		downloadFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return nil, wantErr
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, storage)
	_, _, err := uc.Download(t.Context(), "doc-1")
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

// --- Delete error branches ---

func TestDelete_NotFound(t *testing.T) {
	docRepo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Document, error) {
			return nil, domain.ErrDocumentNotFound
		},
	}
	uc := New(docRepo, &mockVersionRepo{}, &mockFileStorage{})
	err := uc.Delete(t.Context(), "doc-x", "user-1")
	if !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("error = %v, want %v", err, domain.ErrDocumentNotFound)
	}
}

func TestDelete_WrongUser(t *testing.T) {
	docRepo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Document, error) {
			return &domain.Document{ID: "doc-1", UploadedBy: "user-1"}, nil
		},
	}
	uc := New(docRepo, &mockVersionRepo{}, &mockFileStorage{})
	err := uc.Delete(t.Context(), "doc-1", "user-other")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDelete_ListVersionsError(t *testing.T) {
	wantErr := errors.New("db error")
	docRepo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Document, error) {
			return &domain.Document{ID: "doc-1", UploadedBy: "user-1"}, nil
		},
	}
	versionRepo := &mockVersionRepo{
		listByDocumentFn: func(_ context.Context, _ string) ([]*domain.DocumentVersion, error) {
			return nil, wantErr
		},
	}
	uc := New(docRepo, versionRepo, &mockFileStorage{})
	err := uc.Delete(t.Context(), "doc-1", "user-1")
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestDelete_StorageDeleteError(t *testing.T) {
	wantErr := errors.New("s3 delete failed")
	docRepo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Document, error) {
			return &domain.Document{ID: "doc-1", UploadedBy: "user-1"}, nil
		},
	}
	versionRepo := &mockVersionRepo{
		listByDocumentFn: func(_ context.Context, _ string) ([]*domain.DocumentVersion, error) {
			return []*domain.DocumentVersion{{ID: "v1", FileKey: "k1"}}, nil
		},
	}
	storage := &mockFileStorage{
		deleteFn: func(_ context.Context, _ string) error { return wantErr },
	}
	uc := New(docRepo, versionRepo, storage)
	err := uc.Delete(t.Context(), "doc-1", "user-1")
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestDelete_DocRepoDeleteError(t *testing.T) {
	wantErr := errors.New("db delete failed")
	docRepo := &mockDocumentRepo{
		getByIDFn: func(_ context.Context, _ string) (*domain.Document, error) {
			return &domain.Document{ID: "doc-1", UploadedBy: "user-1"}, nil
		},
		deleteFn: func(_ context.Context, _ string) error { return wantErr },
	}
	versionRepo := &mockVersionRepo{
		listByDocumentFn: func(_ context.Context, _ string) ([]*domain.DocumentVersion, error) {
			return []*domain.DocumentVersion{}, nil
		},
	}
	uc := New(docRepo, versionRepo, &mockFileStorage{})
	err := uc.Delete(t.Context(), "doc-1", "user-1")
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

// --- UpdateVersionFlags ---

func TestUpdateVersionFlags_WithFinal(t *testing.T) {
	var clearedProject string
	var updatedID string
	var updatedSigned, updatedFinal bool

	versionRepo := &mockVersionRepo{
		clearFinalByProjectFn: func(_ context.Context, projectID string) error {
			clearedProject = projectID
			return nil
		},
		updateFlagsFn: func(_ context.Context, id string, isSigned, isFinal bool) error {
			updatedID = id
			updatedSigned = isSigned
			updatedFinal = isFinal
			return nil
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	err := uc.UpdateVersionFlags(t.Context(), "proj-1", "ver-1", true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clearedProject != "proj-1" {
		t.Errorf("clearedProject = %q, want %q", clearedProject, "proj-1")
	}
	if updatedID != "ver-1" {
		t.Errorf("updatedID = %q, want %q", updatedID, "ver-1")
	}
	if !updatedSigned {
		t.Error("expected isSigned=true")
	}
	if !updatedFinal {
		t.Error("expected isFinal=true")
	}
}

func TestUpdateVersionFlags_WithoutFinal(t *testing.T) {
	var clearCalled bool
	versionRepo := &mockVersionRepo{
		clearFinalByProjectFn: func(_ context.Context, _ string) error {
			clearCalled = true
			return nil
		},
		updateFlagsFn: func(_ context.Context, _ string, _, _ bool) error {
			return nil
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	err := uc.UpdateVersionFlags(t.Context(), "proj-1", "ver-1", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clearCalled {
		t.Error("ClearFinalByProject should not be called when isFinal=false")
	}
}

func TestUpdateVersionFlags_ClearFinalError(t *testing.T) {
	wantErr := errors.New("clear failed")
	versionRepo := &mockVersionRepo{
		clearFinalByProjectFn: func(_ context.Context, _ string) error { return wantErr },
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	err := uc.UpdateVersionFlags(t.Context(), "proj-1", "ver-1", false, true)
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestUpdateVersionFlags_UpdateFlagsError(t *testing.T) {
	wantErr := errors.New("update failed")
	versionRepo := &mockVersionRepo{
		updateFlagsFn: func(_ context.Context, _ string, _, _ bool) error { return wantErr },
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	err := uc.UpdateVersionFlags(t.Context(), "proj-1", "ver-1", true, false)
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

// --- SetVersionTags ---

func TestSetVersionTags_Success(t *testing.T) {
	var savedTags []string
	versionRepo := &mockVersionRepo{
		setTagsFn: func(_ context.Context, _ string, tags []string) error {
			savedTags = tags
			return nil
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	err := uc.SetVersionTags(t.Context(), "ver-1", []string{"срочно", "черновик"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(savedTags) != 2 {
		t.Errorf("len(savedTags) = %d, want 2", len(savedTags))
	}
}

func TestSetVersionTags_MaxExceeded(t *testing.T) {
	uc := New(&mockDocumentRepo{}, &mockVersionRepo{}, &mockFileStorage{})
	err := uc.SetVersionTags(t.Context(), "ver-1", []string{"a", "b", "c", "d"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetVersionTags_RepoError(t *testing.T) {
	wantErr := errors.New("db error")
	versionRepo := &mockVersionRepo{
		setTagsFn: func(_ context.Context, _ string, _ []string) error { return wantErr },
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	err := uc.SetVersionTags(t.Context(), "ver-1", []string{"tag1"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestSetVersionTags_Empty(t *testing.T) {
	var savedTags []string
	versionRepo := &mockVersionRepo{
		setTagsFn: func(_ context.Context, _ string, tags []string) error {
			savedTags = tags
			return nil
		},
	}
	uc := New(&mockDocumentRepo{}, versionRepo, &mockFileStorage{})
	err := uc.SetVersionTags(t.Context(), "ver-1", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(savedTags) != 0 {
		t.Errorf("len(savedTags) = %d, want 0", len(savedTags))
	}
}
