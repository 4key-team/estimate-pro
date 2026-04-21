package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/VDV001/estimate-pro/backend/internal/shared/errors"
)

// mockRoleGetter implements RoleGetter with configurable behavior.
type mockRoleGetter struct {
	GetRoleFn func(ctx context.Context, projectID, userID string) (string, error)
}

func (m *mockRoleGetter) GetRole(ctx context.Context, projectID, userID string) (string, error) {
	return m.GetRoleFn(ctx, projectID, userID)
}

func TestRequireProjectMember(t *testing.T) {
	const (
		testUserID    = "usr_123"
		testProjectID = "prj_456"
	)

	tests := []struct {
		name           string
		userID         string // empty means no user in context
		urlParam       string // "id" or "projectId"
		paramValue     string
		roleFn         func(ctx context.Context, projectID, userID string) (string, error)
		wantStatus     int
		wantNextCalled bool
		wantErrCode    string
		wantRole       string
	}{
		{
			name:       "user is member passes through",
			userID:     testUserID,
			urlParam:   "id",
			paramValue: testProjectID,
			roleFn: func(_ context.Context, _, _ string) (string, error) {
				return "owner", nil
			},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantRole:       "owner",
		},
		{
			name:       "user is not member returns 403",
			userID:     testUserID,
			urlParam:   "id",
			paramValue: testProjectID,
			roleFn: func(_ context.Context, _, _ string) (string, error) {
				return "", fmt.Errorf("not found")
			},
			wantStatus:  http.StatusForbidden,
			wantErrCode: "FORBIDDEN",
		},
		{
			name:        "missing user in context returns 401",
			userID:      "",
			urlParam:    "id",
			paramValue:  testProjectID,
			wantStatus:  http.StatusUnauthorized,
			wantErrCode: "UNAUTHORIZED",
		},
		{
			name:        "missing project ID returns 400",
			userID:      testUserID,
			urlParam:    "",
			paramValue:  "",
			wantStatus:  http.StatusBadRequest,
			wantErrCode: "BAD_REQUEST",
		},
		{
			name:       "projectId URL param works",
			userID:     testUserID,
			urlParam:   "projectId",
			paramValue: testProjectID,
			roleFn: func(_ context.Context, _, _ string) (string, error) {
				return "member", nil
			},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantRole:       "member",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				nextCalled   bool
				capturedRole string
			)

			mock := &mockRoleGetter{}
			if tt.roleFn != nil {
				mock.GetRoleFn = tt.roleFn
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				role, _ := ProjectRoleFromContext(r.Context())
				capturedRole = role
				w.WriteHeader(http.StatusOK)
			})

			mw := RequireProjectMember(mock)

			// Build a chi router so URL params work.
			r := chi.NewRouter()
			if tt.urlParam == "projectId" {
				r.With(mw).Get("/projects/{projectId}/estimations", next)
			} else if tt.urlParam == "id" {
				r.With(mw).Get("/projects/{id}", next)
			} else {
				// No URL param — use a route without params.
				r.With(mw).Get("/projects", next)
			}

			var path string
			switch tt.urlParam {
			case "projectId":
				path = "/projects/" + tt.paramValue + "/estimations"
			case "id":
				path = "/projects/" + tt.paramValue
			default:
				path = "/projects"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("nextCalled = %v, want %v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantNextCalled && capturedRole != tt.wantRole {
				t.Errorf("role = %q, want %q", capturedRole, tt.wantRole)
			}
			if tt.wantErrCode != "" {
				var errResp errors.ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Fatalf("decoding error response: %v", err)
				}
				if errResp.Error.Code != tt.wantErrCode {
					t.Errorf("error code = %q, want %q", errResp.Error.Code, tt.wantErrCode)
				}
			}
		})
	}
}

func TestProjectRoleFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantRole string
		wantOK   bool
	}{
		{
			name:     "role set",
			ctx:      context.WithValue(context.Background(), ProjectRoleKey, "owner"),
			wantRole: "owner",
			wantOK:   true,
		},
		{
			name:     "empty context",
			ctx:      context.Background(),
			wantRole: "",
			wantOK:   false,
		},
		{
			name:     "wrong type",
			ctx:      context.WithValue(context.Background(), ProjectRoleKey, 42),
			wantRole: "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, ok := ProjectRoleFromContext(tt.ctx)
			if role != tt.wantRole {
				t.Errorf("role = %q, want %q", role, tt.wantRole)
			}
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}
