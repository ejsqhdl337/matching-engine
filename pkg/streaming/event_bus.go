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
	buffers      map[string]*topicBuffer
	capacity     int
	mutex        sync.Mutex
	sequencer    *EventSequencer
	store        *EventStore
	topicManager *TopicManager
	cond         *sync.Cond
}

type topicBuffer struct {
	buffer   []*Event
	head     int
	tail     int
	capacity int
}

func NewEventBus(capacity int, store *EventStore, topicManager *TopicManager) *EventBus {
	bus := &EventBus{
		buffers:      make(map[string]*topicBuffer),
		capacity:     capacity,
		sequencer:    NewEventSequencer(),
		store:        store,
		topicManager: topicManager,
	}
	bus.cond = sync.NewCond(&bus.mutex)
	return bus
}

func (b *EventBus) Add(topicName string, payload []byte) error {
	return b.AddBatch(topicName, [][]byte{payload})
}

func (b *EventBus) AddBatch(topicName string, payloads [][]byte) error {
	topic, err := b.topicManager.GetTopic(topicName)
	if err != nil {
		return err
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, payload := range payloads {
		if err := topic.Validate(payload); err != nil {
			return err
		}

		event := &Event{
			Topic:     topicName,
			Timestamp: b.sequencer.Next(),
			Payload:   payload,
		}

		if err := b.store.Store(event); err != nil {
			return err
		}

		buffer, ok := b.buffers[topicName]
		if !ok {
			buffer = &topicBuffer{
				buffer:   make([]*Event, b.capacity),
				capacity: b.capacity,
			}
			b.buffers[topicName] = buffer
		}

		buffer.buffer[buffer.tail] = event
		buffer.tail = (buffer.tail + 1) % buffer.capacity
		if buffer.tail == buffer.head {
			buffer.head = (buffer.head + 1) % buffer.capacity
		}
	}

	b.cond.Broadcast()
	return nil
}

func (b *EventBus) Poll(topicName string, maxEvents int) ([]*Event, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	buffer, ok := b.buffers[topicName]
	if !ok {
		return nil, nil
	}

	for buffer.head == buffer.tail {
		b.cond.Wait()
	}

	var events []*Event
	if buffer.head < buffer.tail {
		end := buffer.tail
		if end-buffer.head > maxEvents {
			end = buffer.head + maxEvents
		}
		events = buffer.buffer[buffer.head:end]
		buffer.head = end
	} else {
		end := buffer.capacity
		if end-buffer.head > maxEvents {
			end = buffer.head + maxEvents
		}
		events = buffer.buffer[buffer.head:end]
		buffer.head = end % buffer.capacity

		if len(events) < maxEvents {
			remaining := maxEvents - len(events)
			end := buffer.tail
			if end > remaining {
				end = remaining
			}
			events = append(events, buffer.buffer[0:end]...)
			buffer.head = end
		}
	}

	return events, nil
}
