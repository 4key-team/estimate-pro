package ws

import "testing"

// Verify that *Hub implements EventPublisher interface.
func TestHub_ImplementsEventPublisher(t *testing.T) {
	var _ EventPublisher = (*Hub)(nil)
}
