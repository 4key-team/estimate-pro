// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxWorkspaceNameLen = 255

// NewWorkspace creates a Workspace enforcing domain invariants: non-empty
// trimmed name up to 255 chars, non-empty owner. Caller must not bypass this
// constructor by using the struct literal directly.
func NewWorkspace(name, ownerID string) (*Workspace, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > maxWorkspaceNameLen {
		return nil, ErrInvalidWorkspaceName
	}
	if ownerID == "" {
		return nil, ErrMissingOwner
	}
	return &Workspace{
		ID:        uuid.New().String(),
		Name:      trimmed,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
	}, nil
}

// Rename updates the workspace name enforcing the same name invariant.
// If the name is invalid, the current value is not mutated.
func (w *Workspace) Rename(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > maxWorkspaceNameLen {
		return ErrInvalidWorkspaceName
	}
	w.Name = trimmed
	return nil
}

// AuthorizeOwner reports whether the given userID is the workspace owner.
// Returns ErrWorkspaceForbidden otherwise.
func (w *Workspace) AuthorizeOwner(userID string) error {
	if w.OwnerID != userID {
		return ErrWorkspaceForbidden
	}
	return nil
}
