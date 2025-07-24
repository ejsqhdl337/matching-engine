package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	outputBuffer := NewRingBuffer(1024)
	me := NewMatchingEngine(outputBuffer)

	go me.Run()

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		var orders []*Order
		if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		me.PlaceOrders(orders)
		w.WriteHeader(http.StatusAccepted)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		for {
			event, ok := outputBuffer.Pop()
			if !ok {
				continue
			}
			if err := conn.WriteJSON(event); err != nil {
				log.Println(err)
				return
			}
		}
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
