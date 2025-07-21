package main

import (
	"sync"
	"testing"
)

func TestEventBuffer(t *testing.T) {
	var wg sync.WaitGroup
	var loggedEvents []Event
	var processedEvents []Event

	eventLogger := func(e Event) {
		loggedEvents = append(loggedEvents, e)
		wg.Done()
	}

	matchingEngine := func(events []Event) {
		processedEvents = append(processedEvents, events...)
	}

	eventBuffer := &EventBuffer{
		handler:   eventLogger,
		processor: matchingEngine,
	}

	// Test concurrent event handling
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			eventBuffer.AddEvent(Event{Data: "event"})
		}(i)
	}

	wg.Wait()

	if len(loggedEvents) != 10 {
		t.Errorf("Expected 10 logged events, got %d", len(loggedEvents))
	}

	eventBuffer.ProcessEvents()

	if len(processedEvents) != 10 {
		t.Errorf("Expected 10 processed events, got %d", len(processedEvents))
	}
}
