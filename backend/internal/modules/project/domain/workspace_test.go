// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/VDV001/estimate-pro/backend/internal/modules/project/domain"
)

func TestNewWorkspace_Valid(t *testing.T) {
	ws, err := domain.NewWorkspace("My Project", "owner-123")
	if err != nil {
		t.Fatalf("NewWorkspace: unexpected error: %v", err)
	}
	if ws == nil {
		t.Fatal("NewWorkspace: got nil workspace")
	}
	if ws.Name != "My Project" {
		t.Errorf("Name = %q, want %q", ws.Name, "My Project")
	}
	if ws.OwnerID != "owner-123" {
		t.Errorf("OwnerID = %q, want %q", ws.OwnerID, "owner-123")
	}
	if ws.ID == "" {
		t.Error("ID must be auto-generated, got empty")
	}
	if ws.CreatedAt.IsZero() {
		t.Error("CreatedAt must be set, got zero")
	}
}

func TestNewWorkspace_Validation(t *testing.T) {
	tests := []struct {
		name    string
		wsName  string
		ownerID string
		want    error
	}{
		{"empty name", "", "owner-1", domain.ErrInvalidWorkspaceName},
		{"whitespace name", "   ", "owner-1", domain.ErrInvalidWorkspaceName},
		{"name too long", strings.Repeat("x", 256), "owner-1", domain.ErrInvalidWorkspaceName},
		{"empty owner", "ok", "", domain.ErrMissingOwner},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewWorkspace(tc.wsName, tc.ownerID)
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestWorkspace_Rename(t *testing.T) {
	ws, _ := domain.NewWorkspace("Old", "owner-1")

	if err := ws.Rename("New Name"); err != nil {
		t.Fatalf("Rename: unexpected error: %v", err)
	}
	if ws.Name != "New Name" {
		t.Errorf("Name = %q, want %q", ws.Name, "New Name")
	}
}

func TestWorkspace_Rename_InvalidName(t *testing.T) {
	ws, _ := domain.NewWorkspace("Old", "owner-1")

	if err := ws.Rename(""); !errors.Is(err, domain.ErrInvalidWorkspaceName) {
		t.Errorf("Rename empty: err = %v, want ErrInvalidWorkspaceName", err)
	}
	if err := ws.Rename(strings.Repeat("x", 256)); !errors.Is(err, domain.ErrInvalidWorkspaceName) {
		t.Errorf("Rename too long: err = %v, want ErrInvalidWorkspaceName", err)
	}
	if ws.Name != "Old" {
		t.Errorf("Name should stay 'Old' after failed Rename, got %q", ws.Name)
	}
}
