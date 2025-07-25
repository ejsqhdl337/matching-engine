package streaming

import (
	"sync"
)

// Consideration for event bus design:
// - We are using a simple in-memory event bus. This is a good starting point, but it has its limitations.
// - For a production system, we would want to use a more robust event bus, such as a distributed message broker like Kafka or RabbitMQ.
// - We are using a simple ring buffer for storing events in memory. This is a good choice for performance, but it has a fixed size.
// - For a more flexible solution, we could use a dynamic data structure, such as a slice or a linked list.
// - We are using a mutex to protect the ring buffer from concurrent access. This is a simple solution, but it can be a bottleneck.
// - For a high-performance system, we would want to use a lock-free data structure or a single-writer design.
type EventBus struct {
	buffer     []*Event
	capacity   int
	head       int
	tail       int
	mutex      sync.Mutex
	sequencer  *EventSequencer
	store      *EventStore
	cond       *sync.Cond
}

func NewEventBus(capacity int, store *EventStore) *EventBus {
	bus := &EventBus{
		buffer:    make([]*Event, capacity),
		capacity:  capacity,
		sequencer: NewEventSequencer(),
		store:     store,
	}
	bus.cond = sync.NewCond(&bus.mutex)
	return bus
}

func (b *EventBus) Add(payload []byte) error {
	return b.AddBatch([][]byte{payload})
}

func (b *EventBus) AddBatch(payloads [][]byte) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, payload := range payloads {
		event := &Event{
			Timestamp: b.sequencer.Next(),
			Payload:   payload,
		}

		if err := b.store.Store(event); err != nil {
			return err
		}

		b.buffer[b.tail] = event
		b.tail = (b.tail + 1) % b.capacity
		if b.tail == b.head {
			b.head = (b.head + 1) % b.capacity
		}
	}

	b.cond.Broadcast()
	return nil
}

func (b *EventBus) Poll(maxEvents int) ([]*Event, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for b.head == b.tail {
		b.cond.Wait()
	}

	var events []*Event
	if b.head < b.tail {
		end := b.tail
		if end-b.head > maxEvents {
			end = b.head + maxEvents
		}
		events = b.buffer[b.head:end]
		b.head = end
	} else {
		end := b.capacity
		if end-b.head > maxEvents {
			end = b.head + maxEvents
		}
		events = b.buffer[b.head:end]
		b.head = end % b.capacity

		if len(events) < maxEvents {
			remaining := maxEvents - len(events)
			end := b.tail
			if end > remaining {
				end = remaining
			}
			events = append(events, b.buffer[0:end]...)
			b.head = end
		}
	}

	return events, nil
}
