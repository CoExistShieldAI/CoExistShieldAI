package geofence

import (
	"testing"
	"time"

	"ai.coexistshield/backend/internal/model"
)

func TestEventHub_BroadcastAndSubscribe(t *testing.T) {
	hub := NewEventHub()
	if hub.SubscriberCount() != 0 {
		t.Fatalf("Expected 0 subscribers, got %d", hub.SubscriberCount())
	}

	ch := hub.Subscribe()
	if hub.SubscriberCount() != 1 {
		t.Fatalf("Expected 1 subscriber, got %d", hub.SubscriberCount())
	}

	ev := model.GeofenceEvent{
		Type:       "CREATED",
		Action:     "created",
		GeofenceID: "geo_test_123",
		Tier:       "TIER_3_DANGER",
		Message:    "High risk alert zone",
	}

	hub.Broadcast(ev)

	select {
	case received := <-ch:
		if received.GeofenceID != "geo_test_123" {
			t.Errorf("Expected geofenceId geo_test_123, got %s", received.GeofenceID)
		}
		if received.Type != "CREATED" {
			t.Errorf("Expected Type CREATED, got %s", received.Type)
		}
		if received.Timestamp == "" {
			t.Error("Expected Timestamp to be populated")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for event on subscriber channel")
	}

	hub.Unsubscribe(ch)
	if hub.SubscriberCount() != 0 {
		t.Fatalf("Expected 0 subscribers after unsubscribe, got %d", hub.SubscriberCount())
	}
}
