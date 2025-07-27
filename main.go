package main

import (
	"log"
	"matching_engine/pkg/streaming"
)

func main() {
	store, err := streaming.NewEventStore("events.log", true)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	topics := []*streaming.Topic{
		{
			Name: "order",
			Schema: map[string]interface{}{
				"user_id":  "string",
				"price":    "float64",
				"amount":   "float64",
				"side":     "string",
				"type":     "string",
				"id":       "float64",
				"quantity": "float64",
			},
		},
		{
			Name: "snapshot",
			Schema: map[string]interface{}{
				"snapshot_id": "string",
				"data":        "string",
			},
		},
	}

	topicManager := streaming.NewTopicManager(topics)

	bus := streaming.NewEventBus(1024, store, topicManager)
	server := streaming.NewServer(bus)

	go startMatchingEngine()

	log.Println("Event streaming server started on :8081")
	if err := server.ListenAndServe(8081); err != nil {
		log.Fatal(err)
	}
}
