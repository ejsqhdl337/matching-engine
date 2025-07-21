package main

import (
	"sync"
)

type Event struct {
	Data string
}

type EventBuffer struct {
	buffer    []Event
	lock      sync.Mutex
	handler   func(Event)
	processor func([]Event)
}

func (eb *EventBuffer) AddEvent(e Event) {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	eb.buffer = append(eb.buffer, e)
	if eb.handler != nil {
		eb.handler(e)
	}
}

func (eb *EventBuffer) ProcessEvents() {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	if eb.processor != nil {
		eb.processor(eb.buffer)
	}
	eb.buffer = nil
}
