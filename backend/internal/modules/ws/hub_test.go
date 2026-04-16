package ws

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHub_RegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		UserID:     "user-1",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1", hub.ClientCount())
	}

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("client count = %d, want 0", hub.ClientCount())
	}
}

func TestHub_BroadcastToProjectMembers(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Client in proj-1
	c1 := &Client{
		UserID:     "user-1",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}
	// Client in proj-2
	c2 := &Client{
		UserID:     "user-2",
		ProjectIDs: map[string]bool{"proj-2": true},
		Send:       make(chan []byte, 10),
	}
	// Client in both
	c3 := &Client{
		UserID:     "user-3",
		ProjectIDs: map[string]bool{"proj-1": true, "proj-2": true},
		Send:       make(chan []byte, 10),
	}

	hub.Register(c1)
	hub.Register(c2)
	hub.Register(c3)
	time.Sleep(10 * time.Millisecond)

	// Broadcast to proj-1
	hub.Broadcast(Event{Type: "test.event", ProjectID: "proj-1", Payload: "hello"})
	time.Sleep(10 * time.Millisecond)

	// c1 and c3 should receive, c2 should not
	if len(c1.Send) != 1 {
		t.Errorf("c1 got %d messages, want 1", len(c1.Send))
	}
	if len(c2.Send) != 0 {
		t.Errorf("c2 got %d messages, want 0", len(c2.Send))
	}
	if len(c3.Send) != 1 {
		t.Errorf("c3 got %d messages, want 1", len(c3.Send))
	}

	// Verify event content
	msg := <-c1.Send
	var event Event
	json.Unmarshal(msg, &event)
	if event.Type != "test.event" {
		t.Errorf("event type = %q, want test.event", event.Type)
	}
}

func TestHub_BroadcastGlobal(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	c1 := &Client{UserID: "u1", ProjectIDs: map[string]bool{"p1": true}, Send: make(chan []byte, 10)}
	c2 := &Client{UserID: "u2", ProjectIDs: map[string]bool{"p2": true}, Send: make(chan []byte, 10)}

	hub.Register(c1)
	hub.Register(c2)
	time.Sleep(10 * time.Millisecond)

	// Broadcast with empty ProjectID — goes to everyone
	hub.Broadcast(Event{Type: "global.event", ProjectID: ""})
	time.Sleep(10 * time.Millisecond)

	if len(c1.Send) != 1 {
		t.Errorf("c1 got %d messages, want 1", len(c1.Send))
	}
	if len(c2.Send) != 1 {
		t.Errorf("c2 got %d messages, want 1", len(c2.Send))
	}
}

func TestHub_BroadcastSkipsSender(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	sender := &Client{
		UserID:     "sender-1",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}
	receiver := &Client{
		UserID:     "receiver-1",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}

	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	hub.Broadcast(Event{
		Type:      "estimation.submitted",
		ProjectID: "proj-1",
		UserID:    "sender-1",
		Payload:   "data",
	})
	time.Sleep(10 * time.Millisecond)

	if len(sender.Send) != 0 {
		t.Errorf("sender got %d messages, want 0 (should be skipped)", len(sender.Send))
	}
	if len(receiver.Send) != 1 {
		t.Errorf("receiver got %d messages, want 1", len(receiver.Send))
	}
}

func TestHub_BroadcastFullBuffer(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Client with buffer size 0 — will always be full
	slow := &Client{
		UserID:     "slow-user",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte), // unbuffered = always blocks
	}
	fast := &Client{
		UserID:     "fast-user",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}

	hub.Register(slow)
	hub.Register(fast)
	time.Sleep(10 * time.Millisecond)

	hub.Broadcast(Event{Type: "test.event", ProjectID: "proj-1"})
	time.Sleep(10 * time.Millisecond)

	// fast client should still get the message even though slow client's buffer is full
	if len(fast.Send) != 1 {
		t.Errorf("fast client got %d messages, want 1", len(fast.Send))
	}
	// slow client should have 0 (message was dropped due to full buffer)
	if len(slow.Send) != 0 {
		t.Errorf("slow client got %d messages, want 0 (should be dropped)", len(slow.Send))
	}
}

func TestHub_UnregisterNonExistentClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	registered := &Client{
		UserID:     "user-1",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}
	unknown := &Client{
		UserID:     "user-unknown",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}

	hub.Register(registered)
	time.Sleep(10 * time.Millisecond)

	// Unregister a client that was never registered — should be a no-op
	hub.Unregister(unknown)
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Errorf("client count = %d, want 1 (registered client should remain)", hub.ClientCount())
	}

	// Verify unknown client's Send channel is still open (not closed)
	select {
	case unknown.Send <- []byte("test"):
		<-unknown.Send // drain
	default:
		t.Error("unknown client's Send channel should still be open")
	}
}

func TestHub_MultipleClientsOnSameProject(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	clients := make([]*Client, 5)
	for i := range clients {
		clients[i] = &Client{
			UserID:     "user-" + string(rune('A'+i)),
			ProjectIDs: map[string]bool{"shared-proj": true},
			Send:       make(chan []byte, 10),
		}
		hub.Register(clients[i])
	}
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 5 {
		t.Fatalf("client count = %d, want 5", hub.ClientCount())
	}

	hub.Broadcast(Event{Type: "update", ProjectID: "shared-proj", Payload: "v2"})
	time.Sleep(10 * time.Millisecond)

	for i, c := range clients {
		if len(c.Send) != 1 {
			t.Errorf("client[%d] got %d messages, want 1", i, len(c.Send))
		}
	}
}

func TestHub_BroadcastEventContent(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	c := &Client{
		UserID:     "user-1",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}
	hub.Register(c)
	time.Sleep(10 * time.Millisecond)

	original := Event{
		Type:      "document.uploaded",
		ProjectID: "proj-1",
		Payload:   map[string]string{"filename": "spec.pdf"},
	}
	hub.Broadcast(original)
	time.Sleep(10 * time.Millisecond)

	msg := <-c.Send
	var received Event
	if err := json.Unmarshal(msg, &received); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}
	if received.Type != "document.uploaded" {
		t.Errorf("type = %q, want document.uploaded", received.Type)
	}
	if received.ProjectID != "proj-1" {
		t.Errorf("project_id = %q, want proj-1", received.ProjectID)
	}
}

func TestHub_BroadcastMultipleEvents(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	c := &Client{
		UserID:     "user-1",
		ProjectIDs: map[string]bool{"proj-1": true},
		Send:       make(chan []byte, 10),
	}
	hub.Register(c)
	time.Sleep(10 * time.Millisecond)

	for i := range 3 {
		hub.Broadcast(Event{
			Type:      "event",
			ProjectID: "proj-1",
			Payload:   i,
		})
	}
	time.Sleep(20 * time.Millisecond)

	if len(c.Send) != 3 {
		t.Errorf("got %d messages, want 3", len(c.Send))
	}
}

func TestHub_RegisterMultipleThenUnregisterOne(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	c1 := &Client{UserID: "u1", ProjectIDs: map[string]bool{"p1": true}, Send: make(chan []byte, 10)}
	c2 := &Client{UserID: "u2", ProjectIDs: map[string]bool{"p1": true}, Send: make(chan []byte, 10)}
	c3 := &Client{UserID: "u3", ProjectIDs: map[string]bool{"p1": true}, Send: make(chan []byte, 10)}

	hub.Register(c1)
	hub.Register(c2)
	hub.Register(c3)
	time.Sleep(10 * time.Millisecond)

	hub.Unregister(c2)
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 2 {
		t.Errorf("client count = %d, want 2", hub.ClientCount())
	}

	hub.Broadcast(Event{Type: "test", ProjectID: "p1"})
	time.Sleep(10 * time.Millisecond)

	if len(c1.Send) != 1 {
		t.Errorf("c1 got %d messages, want 1", len(c1.Send))
	}
	if len(c3.Send) != 1 {
		t.Errorf("c3 got %d messages, want 1", len(c3.Send))
	}
	// c2 was unregistered and channel closed — don't read from it
}
