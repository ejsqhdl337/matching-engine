package main

import (
	"testing"
)

func TestEventBus(t *testing.T) {
	bus := NewEventBus(1024)
	sub1 := bus.Subscribe()
	sub2 := bus.Subscribe()

	bus.Publish(Event{Data: "1"})
	bus.Publish(Event{Data: "2"})

	e, ok := sub1.Poll()
	if !ok || e.Data != "1" {
		t.Errorf("sub1 poll failed, expected 1, got %s", e.Data)
	}
	e, ok = sub1.Poll()
	if !ok || e.Data != "2" {
		t.Errorf("sub1 poll failed, expected 2, got %s", e.Data)
	}
	_, ok = sub1.Poll()
	if ok {
		t.Error("sub1 poll should have failed")
	}

	e, ok = sub2.Poll()
	if !ok || e.Data != "1" {
		t.Errorf("sub2 poll failed, expected 1, got %s", e.Data)
	}
	e, ok = sub2.Poll()
	if !ok || e.Data != "2" {
		t.Errorf("sub2 poll failed, expected 2, got %s", e.Data)
	}
	_, ok = sub2.Poll()
	if ok {
		t.Error("sub2 poll should have failed")
	}

	bus.Publish(Event{Data: "3"})

	e, ok = sub1.Poll()
	if !ok || e.Data != "3" {
		t.Errorf("sub1 poll failed, expected 3, got %s", e.Data)
	}

	e, ok = sub2.Poll()
	if !ok || e.Data != "3" {
		t.Errorf("sub2 poll failed, expected 3, got %s", e.Data)
	}
}
