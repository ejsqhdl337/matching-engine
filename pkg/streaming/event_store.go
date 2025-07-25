package streaming

import (
	"bufio"
	"encoding/gob"
	"os"
	"sync"
)

// Consideration for event storage:
// - We are using a simple file-based storage for events. This is a good starting point, but it has its limitations.
// - For a production system, we would want to use a more robust storage solution, such as a distributed database or a log-based message broker like Kafka.
// - We are using gob for encoding, which is simple to use but not very efficient. For a high-performance system, we would want to use a more efficient encoding, such as protobuf or flatbuffers.
// - We are using a mutex to protect the file from concurrent writes. This is a simple solution, but it can be a bottleneck. For a high-performance system, we would want to use a lock-free data structure or a single-writer design.
type EventStore struct {
	file     *os.File
	writer   *bufio.Writer
	encoder  *gob.Encoder
	mutex    sync.Mutex
	enabled  bool
}

func NewEventStore(filePath string, enabled bool) (*EventStore, error) {
	if !enabled {
		return &EventStore{enabled: false}, nil
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(file)
	encoder := gob.NewEncoder(writer)

	return &EventStore{
		file:    file,
		writer:  writer,
		encoder: encoder,
		enabled: true,
	}, nil
}

func (s *EventStore) Store(event *Event) error {
	if !s.enabled {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.encoder.Encode(event)
}

func (s *EventStore) Close() error {
	if !s.enabled {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.file.Close()
}
