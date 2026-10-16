package geofence

import (
	"sync"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// EventHub manages real-time broadcast of geofence change events to connected clients (SSE).
type EventHub struct {
	mu          sync.RWMutex
	subscribers map[chan model.GeofenceEvent]struct{}
}

// NewEventHub creates and returns an initialized EventHub.
func NewEventHub() *EventHub {
	return &EventHub{
		subscribers: make(map[chan model.GeofenceEvent]struct{}),
	}
}

// Subscribe registers a new subscriber channel. Buffer size prevents slow consumers from immediately dropping events.
func (h *EventHub) Subscribe() chan model.GeofenceEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan model.GeofenceEvent, 32)
	h.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe unregisters and closes a subscriber channel.
func (h *EventHub) Unsubscribe(ch chan model.GeofenceEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscribers[ch]; ok {
		delete(h.subscribers, ch)
		close(ch)
	}
}

// SubscriberCount returns the current count of active subscribers.
func (h *EventHub) SubscriberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers)
}

// Broadcast sends a geofence change event to all active subscribers without blocking.
func (h *EventHub) Broadcast(event model.GeofenceEvent) {
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers {
		select {
		case ch <- event:
		default:
			// Non-blocking drop if consumer buffer is full to preserve real-time latency
		}
	}
}
