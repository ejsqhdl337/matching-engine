package main

import (
	"testing"
)

func TestEventBus(t *testing.T) {
	t.Run("should publish and poll events", func(t *testing.T) {
		eb := NewEventBus(10)
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
	})

	t.Run("should return not ok when polling empty buffer", func(t *testing.T) {
		eb := NewEventBus(10)
		sub := eb.Subscribe()

		_, ok := sub.Poll()
		if ok {
			t.Error("Expected no events, but got one")
		}
	})

	t.Run("should handle multiple subscribers", func(t *testing.T) {
		eb := NewEventBus(10)
		sub1 := eb.Subscribe()
		sub2 := eb.Subscribe()

		eb.Publish(Event{Data: "event1"})

		event1, ok1 := sub1.Poll()
		if !ok1 || event1.Data != "event1" {
			t.Errorf("Subscriber 1 expected event1, got %v", event1)
		}

		event2, ok2 := sub2.Poll()
		if !ok2 || event2.Data != "event1" {
			t.Errorf("Subscriber 2 expected event1, got %v", event2)
		}
	})

	t.Run("should handle buffer full condition", func(t *testing.T) {
		eb := NewEventBus(2)
		sub := eb.Subscribe()

		eb.Publish(Event{Data: "event1"})
		eb.Publish(Event{Data: "event2"})
		eb.Publish(Event{Data: "event3"}) // This should overwrite event1

		// The subscription starts after the events have been published,
		// so it will only see the events currently in the buffer.
		event, ok := sub.Poll()
		if !ok || event.Data != "event2" {
			t.Errorf("Expected event2, got %v", event)
		}

		event, ok = sub.Poll()
		if !ok || event.Data != "event3" {
			t.Errorf("Expected event3, got %v", event)
		}

		_, ok = sub.Poll()
		if ok {
			t.Error("Expected no more events")
		}
	})
}
