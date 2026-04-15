package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "GET request logs method path status duration",
			method:     http.MethodGet,
			path:       "/api/v1/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST request logs correctly",
			method:     http.MethodPost,
			path:       "/api/v1/login",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "error status logged",
			method:     http.MethodGet,
			path:       "/api/v1/missing",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.wantStatus)
			})

			handler := Logger(logger)(next)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// Parse the JSON log line.
			var logEntry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("parsing log output: %v\nraw: %s", err, buf.String())
			}

			// Verify required fields.
			if got, ok := logEntry["method"].(string); !ok || got != tt.method {
				t.Errorf("log method = %v, want %q", logEntry["method"], tt.method)
			}
			if got, ok := logEntry["path"].(string); !ok || got != tt.path {
				t.Errorf("log path = %v, want %q", logEntry["path"], tt.path)
			}
			// Status is logged as a float64 from JSON.
			if got, ok := logEntry["status"].(float64); !ok || int(got) != tt.wantStatus {
				t.Errorf("log status = %v, want %d", logEntry["status"], tt.wantStatus)
			}
			if _, ok := logEntry["duration"].(string); !ok {
				t.Error("log duration field missing or not a string")
			}
		})
	}
}
