package composio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteAction(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"success", http.StatusOK, false},
		{"created", http.StatusCreated, false},
		{"bad_request", http.StatusBadRequest, true},
		{"server_error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-API-Key") != "test-key" {
					t.Error("missing or wrong API key")
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing Content-Type")
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			client := NewClient("test-key").WithBaseURL(srv.URL)
			err := client.ExecuteAction(t.Context(), "TEST_ACTION", "acc-123", map[string]any{"key": "value"})

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecuteAction_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient("key").WithBaseURL(srv.URL)
	_ = client.ExecuteAction(t.Context(), "GMAIL_SEND_EMAIL", "acc", nil)

	if gotPath != "/actions/GMAIL_SEND_EMAIL/execute" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

// Verify context cancellation propagates correctly.
func TestExecuteAction_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately

	client := NewClient("key").WithBaseURL(srv.URL)
	err := client.ExecuteAction(ctx, "ACTION", "acc", nil)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
