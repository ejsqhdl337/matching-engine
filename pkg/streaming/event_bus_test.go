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

	topics := []*Topic{
		{
			Name: "test",
			Schema: map[string]interface{}{
				"message": "string",
			},
		},
	}
	topicManager := NewTopicManager(topics)

	bus := NewEventBus(10, store, topicManager)

	// Test Add and Poll
	go func() {
		time.Sleep(100 * time.Millisecond)
		bus.Add("test", []byte(`{"message": "test event 1"}`))
		bus.Add("test", []byte(`{"message": "test event 2"}`))
	}()

	time.Sleep(200 * time.Millisecond)

	events, err := bus.Poll("test", 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if string(events[0].Payload) != `{"message": "test event 1"}` {
		t.Errorf("expected '{\"message\": \"test event 1\"}', got '%s'", string(events[0].Payload))
	}

	if string(events[1].Payload) != `{"message": "test event 2"}` {
		t.Errorf("expected '{\"message\": \"test event 2\"}', got '%s'", string(events[1].Payload))
	}

	// Test AddBatch
	go func() {
		time.Sleep(100 * time.Millisecond)
		bus.AddBatch("test", [][]byte{[]byte(`{"message": "test event 3"}`), []byte(`{"message": "test event 4"}`)})
	}()

	events, err = bus.Poll("test", 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if string(events[0].Payload) != `{"message": "test event 3"}` {
		t.Errorf("expected '{\"message\": \"test event 3\"}', got '%s'", string(events[0].Payload))
	}

	if string(events[1].Payload) != `{"message": "test event 4"}` {
		t.Errorf("expected '{\"message\": \"test event 4\"}', got '%s'", string(events[1].Payload))
	}
}
