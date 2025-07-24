package main

import (
	"sync/atomic"
)

type Event struct {
	Order *Order
	Data  string
}

// A CacheLinePad is used to pad structs to avoid false sharing.
// Common cache line size is 64 bytes.
type CacheLinePad [64]byte

// RingBuffer represents a single-producer, single-consumer (SPSC) lock-free queue.
type RingBuffer struct {
	data []Event // The underlying fixed-size buffer

	// head and tail indices are padded to prevent false sharing.
	// We use uint64 for indices for Go's atomic operations.
	readIdx  atomic.Uint64 // Reader's index (read by consumer, written by consumer)
	_        CacheLinePad  // Padding to separate readIdx and writeIdxCached

	writeIdxCached uint64 // Consumer's cached copy of the producer's writeIdx
	_              CacheLinePad  // Padding to separate writeIdxCached and writeIdx

	writeIdx atomic.Uint64 // Writer's index (read by producer, written by producer)
	_        CacheLinePad  // Padding to separate writeIdx and readIdxCached

	readIdxCached uint64 // Producer's cached copy of the consumer's readIdx
	mask          uint64
	// No padding needed after the last field
}

// New creates a new RingBuffer with a given capacity.
// Capacity must be a power of two.
func NewRingBuffer(capacity uint64) *RingBuffer {
	if capacity == 0 {
		panic("capacity cannot be zero")
	}
	if (capacity & (capacity - 1)) != 0 {
		panic("capacity must be a power of two for this implementation")
	}
	return &RingBuffer{
		data: make([]Event, capacity),
		mask: capacity - 1,
	}
}

// Push attempts to add an item to the ring buffer.
// Returns true if successful, false if the buffer is full.
func (rb *RingBuffer) Push(item Event) bool {
	writeIdx := rb.writeIdx.Load()
	nextWriteIdx := writeIdx + 1

	// Check if buffer is full using cached read index.
	// If the cached index says it's full, then load the actual read index and re-check.
	if nextWriteIdx-rb.readIdxCached == uint64(len(rb.data)) {
		rb.readIdxCached = rb.readIdx.Load() // Atomic Load
		if nextWriteIdx-rb.readIdxCached == uint64(len(rb.data)) {
			return false // Still full
		}
	}

	rb.data[writeIdx&rb.mask] = item
	rb.writeIdx.Store(nextWriteIdx) // Atomic Store (memory_order_release equivalent)
	return true
}

// Pop attempts to retrieve an item from the ring buffer.
// Returns the item and true if successful, or default(T) and false if the buffer is empty.
func (rb *RingBuffer) Pop() (Event, bool) {
	readIdx := rb.readIdx.Load() // Atomic Load (memory_order_relaxed equivalent in C++ context)

	// Check if buffer is empty using cached write index.
	// If the cached index says it's empty, then load the actual write index and re-check.
	if readIdx == rb.writeIdxCached {
		rb.writeIdxCached = rb.writeIdx.Load() // Atomic Load (memory_order_acquire equivalent)
		if readIdx == rb.writeIdxCached {
			var zero Event // Return zero value for type T
			return zero, false // Still empty
		}
	}

	item := rb.data[readIdx&rb.mask]
	var zero Event // Zero out the slot to allow GC if item is a pointer type
	rb.data[readIdx&rb.mask] = zero

	nextReadIdx := readIdx + 1

	rb.readIdx.Store(nextReadIdx) // Atomic Store (memory_order_release equivalent)
	return item, true
}

// Size returns the approximate number of items in the buffer.
// Note: This is an approximation in a concurrent context without stronger synchronization.
func (rb *RingBuffer) Size() uint64 {
	// Atomically load both indices to get a more consistent view, though still subject to race.
	w := rb.writeIdx.Load()
	r := rb.readIdx.Load()
	return w - r
}
