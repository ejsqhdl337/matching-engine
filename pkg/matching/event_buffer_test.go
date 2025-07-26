package matching

import (
	"sync"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	t.Run("should push and pop items", func(t *testing.T) {
		rb := NewRingBuffer(2)

		ok := rb.Push(Event{Data: "event1"})
		if !ok {
			t.Fatal("push failed")
		}

		event, ok := rb.Pop()
		if !ok || event.Data != "event1" {
			t.Errorf("expected event1, got %v", event)
		}
	})

	t.Run("should return not ok when popping empty buffer", func(t *testing.T) {
		rb := NewRingBuffer(2)

		_, ok := rb.Pop()
		if ok {
			t.Error("expected no events, but got one")
		}
	})

	t.Run("should return not ok when pushing to full buffer", func(t *testing.T) {
		rb := NewRingBuffer(2)

		rb.Push(Event{Data: "event1"})
		rb.Push(Event{Data: "event2"})

		ok := rb.Push(Event{Data: "event3"})
		if ok {
			t.Error("expected push to fail, but it succeeded")
		}
	})

	t.Run("should handle concurrent push and pop", func(t *testing.T) {
		rb := NewRingBuffer(1024)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				for !rb.Push(Event{Data: "event"}) {
				}
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				for {
					_, ok := rb.Pop()
					if ok {
						break
					}
				}
			}
		}()

		wg.Wait()

		if rb.Size() != 0 {
			t.Errorf("expected buffer to be empty, but size is %d", rb.Size())
		}
	})
}
