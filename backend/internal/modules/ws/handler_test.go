package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/VDV001/estimate-pro/backend/pkg/jwt"
)

func newTestJWTService() *jwt.Service {
	return jwt.NewService("test-secret-key-at-least-32-chars!!", 15*time.Minute, 24*time.Hour)
}

func TestHandler_ServeWS_MissingToken(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	jwtSvc := newTestJWTService()
	h := NewHandler(hub, jwtSvc, func(string) []string { return nil })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws", nil)
	w := httptest.NewRecorder()

	h.ServeWS(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "missing token") {
		t.Errorf("body = %q, want 'missing token'", w.Body.String())
	}
}

func TestHandler_ServeWS_InvalidToken(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	jwtSvc := newTestJWTService()
	h := NewHandler(hub, jwtSvc, func(string) []string { return nil })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws?token=invalid.jwt.token", nil)
	w := httptest.NewRecorder()

	h.ServeWS(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "invalid token") {
		t.Errorf("body = %q, want 'invalid token'", w.Body.String())
	}
}

func TestHandler_ServeWS_ValidToken_FullConnection(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	jwtSvc := newTestJWTService()
	pair, err := jwtSvc.GeneratePair("user-ws-1")
	if err != nil {
		t.Fatalf("GeneratePair: %v", err)
	}

	getProjects := func(userID string) []string {
		if userID == "user-ws-1" {
			return []string{"proj-A", "proj-B"}
		}
		return nil
	}

	h := NewHandler(hub, jwtSvc, getProjects)

	// Create test server with the handler
	srv := httptest.NewServer(http.HandlerFunc(h.ServeWS))
	defer srv.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "?token=" + pair.AccessToken

	// Connect via WebSocket
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial: %v (resp: %v)", err, resp)
	}
	defer conn.Close()

	// Give hub time to register the client
	time.Sleep(20 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1", hub.ClientCount())
	}

	// Broadcast an event to proj-A
	hub.Broadcast(Event{Type: "test.event", ProjectID: "proj-A", Payload: "ws-data"})

	// Read the message from WebSocket
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}

	if !strings.Contains(string(msg), "test.event") {
		t.Errorf("message = %q, want to contain 'test.event'", string(msg))
	}

	// Broadcast to a different project — should NOT be received
	hub.Broadcast(Event{Type: "other.event", ProjectID: "proj-C"})
	time.Sleep(20 * time.Millisecond)

	// Close connection to trigger unregister
	conn.Close()
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("client count after close = %d, want 0", hub.ClientCount())
	}
}

func TestHandler_NewHandler(t *testing.T) {
	hub := NewHub()
	jwtSvc := newTestJWTService()
	getProjects := func(string) []string { return nil }

	h := NewHandler(hub, jwtSvc, getProjects)
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
	if h.hub != hub {
		t.Error("hub not set correctly")
	}
	if h.jwtService != jwtSvc {
		t.Error("jwtService not set correctly")
	}
}
