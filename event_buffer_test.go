package main

import (
	"testing"
)

func TestEventBus(t *testing.T) {
	eb := NewEventBus(10)
	if eb == nil {
		t.Fatal("NewEventBus should not return nil")
	}

	sub := eb.Subscribe()

	eb.Publish(Event{Data: "event1"})
	eb.Publish(Event{Data: "event2"})

	event, ok := sub.Poll()
	if !ok || event.Data != "event1" {
		t.Errorf("Expected event1, got %v", event)
	}

	event, ok = sub.Poll()
	if !ok || event.Data != "event2" {
		t.Errorf("Expected event2, got %v", event)
	}

	_, ok = sub.Poll()
	if ok {
		t.Error("Expected no more events")
	}
}
