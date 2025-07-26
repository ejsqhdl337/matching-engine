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

	bus := streaming.NewEventBus(1024, store)
	server := streaming.NewServer(bus)

	log.Println("Event streaming server started on :8081")
	if err := server.ListenAndServe(8081); err != nil {
		log.Fatal(err)
	}
}
