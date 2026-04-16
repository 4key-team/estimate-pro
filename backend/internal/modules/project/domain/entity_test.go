// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/VDV001/estimate-pro/backend/internal/modules/project/domain"
)

// --- Project ---

func TestNewProject_Valid(t *testing.T) {
	p, err := domain.NewProject("ws-1", "My App", "desc", "user-1")
	if err != nil {
		t.Fatalf("NewProject: unexpected error: %v", err)
	}
	if p.ID == "" {
		t.Error("ID must be auto-generated")
	}
	if p.WorkspaceID != "ws-1" || p.Name != "My App" || p.CreatedBy != "user-1" {
		t.Errorf("fields wrong: %+v", p)
	}
	if p.Status != domain.ProjectStatusActive {
		t.Errorf("Status = %q, want active", p.Status)
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		t.Error("timestamps must be set")
	}
}

func TestNewProject_Validation(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		projectName string
		createdBy   string
		want        error
	}{
		{"empty workspace", "", "ok", "u1", domain.ErrMissingWorkspace},
		{"empty name", "ws-1", "", "u1", domain.ErrInvalidProjectName},
		{"whitespace name", "ws-1", "   ", "u1", domain.ErrInvalidProjectName},
		{"too long name", "ws-1", strings.Repeat("x", 256), "u1", domain.ErrInvalidProjectName},
		{"empty creator", "ws-1", "ok", "", domain.ErrMissingCreator},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewProject(tc.workspaceID, tc.projectName, "", tc.createdBy)
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestProject_UpdateDetails(t *testing.T) {
	p, _ := domain.NewProject("ws-1", "Old", "old desc", "u1")
	before := p.UpdatedAt
	time.Sleep(1 * time.Millisecond) // force UpdatedAt change

	if err := p.UpdateDetails("New Name", "new desc"); err != nil {
		t.Fatalf("UpdateDetails: %v", err)
	}
	if p.Name != "New Name" || p.Description != "new desc" {
		t.Errorf("fields not updated: %+v", p)
	}
	if !p.UpdatedAt.After(before) {
		t.Error("UpdatedAt must advance")
	}
}

func TestProject_UpdateDetails_PartialUpdate(t *testing.T) {
	p, _ := domain.NewProject("ws-1", "Old", "orig", "u1")

	if err := p.UpdateDetails("", "only desc"); err != nil {
		t.Fatalf("UpdateDetails: %v", err)
	}
	if p.Name != "Old" {
		t.Errorf("Name should stay 'Old' when empty passed, got %q", p.Name)
	}
	if p.Description != "only desc" {
		t.Errorf("Description = %q, want 'only desc'", p.Description)
	}
}

func TestProject_UpdateDetails_InvalidName(t *testing.T) {
	p, _ := domain.NewProject("ws-1", "Old", "", "u1")

	err := p.UpdateDetails(strings.Repeat("x", 256), "")
	if !errors.Is(err, domain.ErrInvalidProjectName) {
		t.Errorf("err = %v, want ErrInvalidProjectName", err)
	}
	if p.Name != "Old" {
		t.Error("Name must not mutate on validation error")
	}
}

func TestProject_Archive_Restore(t *testing.T) {
	p, _ := domain.NewProject("ws-1", "X", "", "u1")

	p.Archive()
	if p.Status != domain.ProjectStatusArchived {
		t.Errorf("after Archive Status = %q, want archived", p.Status)
	}

	p.Restore()
	if p.Status != domain.ProjectStatusActive {
		t.Errorf("after Restore Status = %q, want active", p.Status)
	}
}

// --- Member ---

func TestNewMember_Valid(t *testing.T) {
	m, err := domain.NewMember("proj-1", "user-1", domain.RoleDeveloper, "caller")
	if err != nil {
		t.Fatalf("NewMember: %v", err)
	}
	if m.ProjectID != "proj-1" || m.UserID != "user-1" || m.Role != domain.RoleDeveloper {
		t.Errorf("fields wrong: %+v", m)
	}
	if m.AddedBy != "caller" {
		t.Errorf("AddedBy = %q, want caller", m.AddedBy)
	}
	if m.AddedAt.IsZero() {
		t.Error("AddedAt must be set")
	}
}

func TestNewMember_EmptyAddedByAllowed(t *testing.T) {
	// System creation (e.g., project creator as initial admin) may omit AddedBy.
	m, err := domain.NewMember("proj-1", "user-1", domain.RoleAdmin, "")
	if err != nil {
		t.Fatalf("NewMember: %v", err)
	}
	if m.AddedBy != "" {
		t.Errorf("AddedBy = %q, want empty", m.AddedBy)
	}
}

func TestNewMember_Validation(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		userID    string
		role      domain.Role
		want      error
	}{
		{"empty project", "", "u1", domain.RoleDeveloper, domain.ErrMissingProject},
		{"empty user", "p1", "", domain.RoleDeveloper, domain.ErrMissingUser},
		{"invalid role", "p1", "u1", domain.Role("hacker"), domain.ErrInvalidRole},
		{"empty role", "p1", "u1", domain.Role(""), domain.ErrInvalidRole},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewMember(tc.projectID, tc.userID, tc.role, "caller")
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
