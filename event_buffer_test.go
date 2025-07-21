package main

import (
	"testing"
)

func TestEventBus(t *testing.T) {
	bus := NewEventBus(4)

	bus.Publish(Event{Data: "1"})
	bus.Publish(Event{Data: "2"})
	bus.Publish(Event{Data: "3"})

	sub := bus.Subscribe()

	bus.Publish(Event{Data: "4"})
	bus.Publish(Event{Data: "5"})

	e, ok := sub.Poll()
	if !ok || e.Data != "4" {
		t.Errorf("poll failed, expected 4, got %s", e.Data)
	}
	e, ok = sub.Poll()
	if !ok || e.Data != "5" {
		t.Errorf("poll failed, expected 5, got %s", e.Data)
	}
	_, ok = sub.Poll()
	if ok {
		t.Error("poll should have failed")
	}
}
