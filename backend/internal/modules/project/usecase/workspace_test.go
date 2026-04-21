// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VDV001/estimate-pro/backend/internal/modules/project/domain"
	"github.com/VDV001/estimate-pro/backend/internal/modules/project/usecase"
)

// --- function-field mock ---

type stubWorkspaceRepo struct {
	createFn  func(ctx context.Context, w *domain.Workspace) error
	getByID   func(ctx context.Context, id string) (*domain.Workspace, error)
	listByUsr func(ctx context.Context, userID string) ([]*domain.Workspace, error)
	updateFn  func(ctx context.Context, w *domain.Workspace) error
}

func (m *stubWorkspaceRepo) Create(ctx context.Context, w *domain.Workspace) error {
	if m.createFn != nil {
		return m.createFn(ctx, w)
	}
	return nil
}
func (m *stubWorkspaceRepo) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, domain.ErrWorkspaceNotFound
}
func (m *stubWorkspaceRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	if m.listByUsr != nil {
		return m.listByUsr(ctx, userID)
	}
	return nil, nil
}
func (m *stubWorkspaceRepo) Update(ctx context.Context, w *domain.Workspace) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, w)
	}
	return nil
}

func TestWorkspaceUsecase_Create_Success(t *testing.T) {
	var saved *domain.Workspace
	repo := &stubWorkspaceRepo{
		createFn: func(_ context.Context, w *domain.Workspace) error {
			saved = w
			return nil
		},
	}
	uc := usecase.NewWorkspaceUsecase(repo)

	ws, err := uc.Create(context.Background(), "My Project", "owner-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ws == nil || saved == nil || ws.ID != saved.ID {
		t.Fatal("Create: workspace was not persisted")
	}
	if ws.Name != "My Project" || ws.OwnerID != "owner-1" {
		t.Errorf("Create: wrong fields, got %+v", ws)
	}
}

func TestWorkspaceUsecase_Create_InvalidName(t *testing.T) {
	uc := usecase.NewWorkspaceUsecase(&stubWorkspaceRepo{})

	_, err := uc.Create(context.Background(), "", "owner-1")
	if !errors.Is(err, domain.ErrInvalidWorkspaceName) {
		t.Errorf("err = %v, want ErrInvalidWorkspaceName", err)
	}
}

func TestWorkspaceUsecase_Rename_Success(t *testing.T) {
	existing := &domain.Workspace{ID: "ws-1", Name: "Old", OwnerID: "owner-1"}
	var updated *domain.Workspace
	repo := &stubWorkspaceRepo{
		getByID: func(_ context.Context, id string) (*domain.Workspace, error) {
			if id == "ws-1" {
				return existing, nil
			}
			return nil, domain.ErrWorkspaceNotFound
		},
		updateFn: func(_ context.Context, w *domain.Workspace) error {
			updated = w
			return nil
		},
	}
	uc := usecase.NewWorkspaceUsecase(repo)

	ws, err := uc.Rename(context.Background(), "ws-1", "New Name", "owner-1")
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if ws.Name != "New Name" || updated.Name != "New Name" {
		t.Errorf("Rename: name not updated, got %q", ws.Name)
	}
}

func TestWorkspaceUsecase_Rename_NotOwner(t *testing.T) {
	existing := &domain.Workspace{ID: "ws-1", Name: "Old", OwnerID: "owner-1"}
	repo := &stubWorkspaceRepo{
		getByID: func(_ context.Context, _ string) (*domain.Workspace, error) {
			return existing, nil
		},
	}
	uc := usecase.NewWorkspaceUsecase(repo)

	_, err := uc.Rename(context.Background(), "ws-1", "New", "other-user")
	if !errors.Is(err, domain.ErrWorkspaceForbidden) {
		t.Errorf("err = %v, want ErrWorkspaceForbidden", err)
	}
	if existing.Name != "Old" {
		t.Errorf("Name must not change on forbidden, got %q", existing.Name)
	}
}

func TestWorkspaceUsecase_Rename_NotFound(t *testing.T) {
	repo := &stubWorkspaceRepo{
		getByID: func(_ context.Context, _ string) (*domain.Workspace, error) {
			return nil, domain.ErrWorkspaceNotFound
		},
	}
	uc := usecase.NewWorkspaceUsecase(repo)

	_, err := uc.Rename(context.Background(), "missing", "New", "owner")
	if !errors.Is(err, domain.ErrWorkspaceNotFound) {
		t.Errorf("err = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestWorkspaceUsecase_Rename_InvalidName(t *testing.T) {
	existing := &domain.Workspace{ID: "ws-1", Name: "Old", OwnerID: "owner-1"}
	repo := &stubWorkspaceRepo{
		getByID: func(_ context.Context, _ string) (*domain.Workspace, error) {
			return existing, nil
		},
	}
	uc := usecase.NewWorkspaceUsecase(repo)

	_, err := uc.Rename(context.Background(), "ws-1", "", "owner-1")
	if !errors.Is(err, domain.ErrInvalidWorkspaceName) {
		t.Errorf("err = %v, want ErrInvalidWorkspaceName", err)
	}
}

func TestWorkspaceUsecase_ListByUser(t *testing.T) {
	want := []*domain.Workspace{{ID: "ws-1"}, {ID: "ws-2"}}
	repo := &stubWorkspaceRepo{
		listByUsr: func(_ context.Context, userID string) ([]*domain.Workspace, error) {
			if userID != "u1" {
				t.Errorf("userID = %q, want u1", userID)
			}
			return want, nil
		},
	}
	uc := usecase.NewWorkspaceUsecase(repo)

	got, err := uc.ListByUser(context.Background(), "u1")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}
