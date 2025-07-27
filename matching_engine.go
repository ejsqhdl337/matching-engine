package main

import (
	"context"
	"encoding/json"
	"log"
	"matching_engine/pkg/matching"
	"matching_engine/pkg/streaming/proto"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

func startMatchingEngine() {
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := proto.NewEventServiceClient(conn)

	me := matching.NewMatchingEngine(nil)

	go func() {
		stream, err := client.Poll(context.Background(), &proto.PollRequest{Topic: "order", MaxEvents: 100})
		if err != nil {
			log.Fatalf("failed to poll: %v", err)
		}
		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Fatalf("failed to receive: %v", err)
			}
			for _, event := range resp.Events {
				var order matching.Order
				if err := json.Unmarshal(event.Payload, &order); err != nil {
					log.Printf("failed to unmarshal order: %v", err)
					continue
				}
				me.PlaceOrder(&order)
			}
		}
	}()

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		var orders []*matching.Order
		if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var payloads [][]byte
		for _, order := range orders {
			payload, err := json.Marshal(order)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			payloads = append(payloads, payload)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if _, err := client.Add(ctx, &proto.AddRequest{Topic: "order", Payloads: payloads}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})

	log.Println("Matching engine server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
