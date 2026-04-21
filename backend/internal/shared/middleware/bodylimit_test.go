package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMaxBodySize(t *testing.T) {
	const limit int64 = 16 // 16 bytes

	tests := []struct {
		name       string
		bodySize   int
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "body within limit passes",
			bodySize:   10,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "body exactly at limit passes",
			bodySize:   int(limit),
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "body exceeds limit fails",
			bodySize:   int(limit) + 10,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.Repeat("x", tt.bodySize)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := io.ReadAll(r.Body)
				if err != nil {
					// MaxBytesReader triggers http.Error with 413 on the ResponseWriter
					// when the handler tries to read beyond the limit.
					http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			handler := MaxBodySize(limit)(next)

			req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
