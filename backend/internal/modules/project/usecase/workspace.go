// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package usecase

import (
	"context"
	"fmt"

	"github.com/VDV001/estimate-pro/backend/internal/modules/project/domain"
)

// WorkspaceUsecase encapsulates workspace-related business logic: creation,
// ownership-guarded rename, listing. All invariants live in domain.Workspace.
type WorkspaceUsecase struct {
	repo domain.WorkspaceRepository
}

func NewWorkspaceUsecase(repo domain.WorkspaceRepository) *WorkspaceUsecase {
	return &WorkspaceUsecase{repo: repo}
}

// Create builds a new Workspace via domain constructor and persists it.
// Domain errors (ErrInvalidWorkspaceName, ErrMissingOwner) are returned as-is.
func (uc *WorkspaceUsecase) Create(ctx context.Context, name, ownerID string) (*domain.Workspace, error) {
	ws, err := domain.NewWorkspace(name, ownerID)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, ws); err != nil {
		return nil, fmt.Errorf("workspace.Create: %w", err)
	}
	return ws, nil
}

// Rename enforces owner authorization then delegates validation to
// Workspace.Rename. Returns ErrWorkspaceNotFound / ErrWorkspaceForbidden /
// ErrInvalidWorkspaceName as typed domain errors.
func (uc *WorkspaceUsecase) Rename(ctx context.Context, workspaceID, name, userID string) (*domain.Workspace, error) {
	ws, err := uc.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if err := ws.AuthorizeOwner(userID); err != nil {
		return nil, err
	}
	if err := ws.Rename(name); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, ws); err != nil {
		return nil, fmt.Errorf("workspace.Rename: %w", err)
	}
	return ws, nil
}

// ListByUser returns all workspaces the user is a member of.
func (uc *WorkspaceUsecase) ListByUser(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	return uc.repo.ListByUser(ctx, userID)
}
