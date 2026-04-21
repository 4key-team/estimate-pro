package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		method         string
		wantStatus     int
		wantOrigin     string // expected Access-Control-Allow-Origin, "" means absent
		wantNextCalled bool
	}{
		{
			name:           "allowed origin sets header",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://localhost:3000",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantOrigin:     "http://localhost:3000",
			wantNextCalled: true,
		},
		{
			name:           "disallowed origin no header",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://evil.com",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantOrigin:     "",
			wantNextCalled: true,
		},
		{
			name:           "OPTIONS preflight returns 204",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://localhost:3000",
			method:         http.MethodOptions,
			wantStatus:     http.StatusNoContent,
			wantOrigin:     "http://localhost:3000",
			wantNextCalled: false,
		},
		{
			name:           "no origin header passes through",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantOrigin:     "",
			wantNextCalled: true,
		},
		{
			name:           "multiple allowed origins first",
			allowedOrigins: []string{"http://localhost:3000", "https://app.example.com"},
			origin:         "http://localhost:3000",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantOrigin:     "http://localhost:3000",
			wantNextCalled: true,
		},
		{
			name:           "multiple allowed origins second",
			allowedOrigins: []string{"http://localhost:3000", "https://app.example.com"},
			origin:         "https://app.example.com",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantOrigin:     "https://app.example.com",
			wantNextCalled: true,
		},
		{
			name:           "default origins when none provided",
			allowedOrigins: nil,
			origin:         "http://localhost:3000",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantOrigin:     "http://localhost:3000",
			wantNextCalled: true,
		},
		{
			name:           "OPTIONS without allowed origin still returns 204",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://evil.com",
			method:         http.MethodOptions,
			wantStatus:     http.StatusNoContent,
			wantOrigin:     "",
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nextCalled bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			var handler http.Handler
			if tt.allowedOrigins == nil {
				handler = CORS()(next)
			} else {
				handler = CORS(tt.allowedOrigins...)(next)
			}

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			gotOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if gotOrigin != tt.wantOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", gotOrigin, tt.wantOrigin)
			}

			if nextCalled != tt.wantNextCalled {
				t.Errorf("nextCalled = %v, want %v", nextCalled, tt.wantNextCalled)
			}

			// Common headers should always be set.
			if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
				t.Error("Access-Control-Allow-Methods header not set")
			}
			if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
				t.Error("Access-Control-Allow-Headers header not set")
			}
			if got := rec.Header().Get("Access-Control-Max-Age"); got != "86400" {
				t.Errorf("Access-Control-Max-Age = %q, want %q", got, "86400")
			}

})
	}
}
