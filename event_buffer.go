package main

import (
	"sync/atomic"
)

type Event struct {
	Data string
}

type RingBuffer struct {
	buffer []Event
	head   int64
	tail   int64
	size   int64
	mask   int64
}

func NewRingBuffer(size int64) *RingBuffer {
	if size <= 0 {
		size = 1024
	}
	// ensure size is a power of 2
	if (size & (size - 1)) != 0 {
		var power int64 = 1
		for power < size {
			power *= 2
		}
		size = power
	}
	return &RingBuffer{
		buffer: make([]Event, size),
		size:   size,
		mask:   size - 1,
	}
}

func (rb *RingBuffer) Enqueue(e Event) {
	tail := atomic.LoadInt64(&rb.tail)
	nextTail := tail + 1
	if nextTail-atomic.LoadInt64(&rb.head) > rb.size {
		atomic.AddInt64(&rb.head, 1)
	}
	rb.buffer[tail&rb.mask] = e
	atomic.StoreInt64(&rb.tail, nextTail)
}

func (rb *RingBuffer) Dequeue() (Event, bool) {
	head := atomic.LoadInt64(&rb.head)
	tail := atomic.LoadInt64(&rb.tail)
	if head >= tail {
		return Event{}, false
	}
	e := rb.buffer[head&rb.mask]
	atomic.AddInt64(&rb.head, 1)
	return e, true
}

type EventBus struct {
	buffer *RingBuffer
}

func NewEventBus(size int64) *EventBus {
	return &EventBus{
		buffer: NewRingBuffer(size),
	}
}

func (eb *EventBus) Publish(e Event) {
	eb.buffer.Enqueue(e)
}

func (eb *EventBus) Subscribe() *Subscription {
	return &Subscription{
		buffer: eb.buffer,
		cursor: atomic.LoadInt64(&eb.buffer.head),
	}
}

type Subscription struct {
	buffer *RingBuffer
	cursor int64
}

func (s *Subscription) Poll() (Event, bool) {
	tail := atomic.LoadInt64(&s.buffer.tail)
	head := atomic.LoadInt64(&s.buffer.head)
	if s.cursor < head {
		s.cursor = head
	}
	if s.cursor >= tail {
		return Event{}, false
	}
	e := s.buffer.buffer[s.cursor&s.buffer.mask]
	s.cursor++
	return e, true
}
