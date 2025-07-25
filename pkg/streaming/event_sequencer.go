package streaming

import (
	"time"
)

// Consideration for event sequencing:
// - We are using a simple atomic counter for sequencing events. This is a good starting point, but it has its limitations.
// - For a distributed system, we would want to use a more robust sequencing mechanism, such as a distributed consensus algorithm like Paxos or Raft.
// - We are using the arrival timestamp for sequencing. This is a good choice for most use cases, but it can be problematic if the clocks on the machines are not synchronized.
// - For a high-performance system, we would want to use a more efficient sequencing mechanism, such as a hardware-based timestamp counter.
type EventSequencer struct {
	sequence uint64
}

func NewEventSequencer() *EventSequencer {
	return &EventSequencer{}
}

func (s *EventSequencer) Next() int64 {
	// We are using the current time in nanoseconds as the sequence number.
	// This is not strictly monotonic, but it is good enough for most use cases.
	// For a more robust solution, we would want to use a combination of a timestamp and a counter.
	return time.Now().UnixNano()
}
