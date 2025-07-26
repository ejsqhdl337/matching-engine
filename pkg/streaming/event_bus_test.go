package streaming

import (
	"os"
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	store, err := NewEventStore("test_events.log", true)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_events.log")
	defer store.Close()

	bus := NewEventBus(10, store)

	// Test Add and Poll
	go func() {
		time.Sleep(100 * time.Millisecond)
		bus.Add([]byte("test event 1"))
		bus.Add([]byte("test event 2"))
	}()

	events, err := bus.Poll(2)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if string(events[0].Payload) != "test event 1" {
		t.Errorf("expected 'test event 1', got '%s'", string(events[0].Payload))
	}

	if string(events[1].Payload) != "test event 2" {
		t.Errorf("expected 'test event 2', got '%s'", string(events[1].Payload))
	}

	// Test AddBatch
	go func() {
		time.Sleep(100 * time.Millisecond)
		bus.AddBatch([][]byte{[]byte("test event 3"), []byte("test event 4")})
	}()

	events, err = bus.Poll(2)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if string(events[0].Payload) != "test event 3" {
		t.Errorf("expected 'test event 3', got '%s'", string(events[0].Payload))
	}

	if string(events[1].Payload) != "test event 4" {
		t.Errorf("expected 'test event 4', got '%s'", string(events[1].Payload))
	}
}
