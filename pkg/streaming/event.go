package streaming

// Consideration for efficient event data structure:
// The Event struct is the fundamental unit of data in our streaming service.
// Its design is critical for performance and memory efficiency.
// - We use a protobuf for the payload, which is a good choice for cross-language compatibility and efficiency.
// - The timestamp is a Unix timestamp in nanoseconds, which provides high resolution for sequencing.
// - We could consider using a more compact data structure for the payload, such as Cap'n Proto or FlatBuffers,
//   if we need to squeeze out every last drop of performance.
// - For now, protobuf is a good starting point.
type Event struct {
	Topic     string
	Timestamp int64
	Payload   []byte
}
