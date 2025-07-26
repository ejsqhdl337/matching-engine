package main

import (
	"encoding/json"
	"log"
	"matching_engine/pkg/matching"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

var (
	trades = make([]*matching.Trade, 0)
)

func main() {
	outputBuffer := matching.NewRingBuffer(1024)
	me := matching.NewMatchingEngine(outputBuffer)
	hub := newHub()

	go me.Run()
	go hub.run()

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		var orders []*matching.Order
		if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		me.PlaceOrders(orders)
		w.WriteHeader(http.StatusAccepted)
	})

	go func() {
		for {
			event, ok := outputBuffer.Pop()
			if !ok {
				continue
			}

			if trade, ok := event.Data.(matching.Trade); ok {
				trades = append(trades, &trade)
			}

			jsonEvent, err := json.Marshal(event)
			if err != nil {
				log.Println(err)
				continue
			}
			hub.broadcast <- jsonEvent
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		client := &Client{conn: conn, send: make(chan []byte, 256)}
		hub.register <- client

		go client.writePump()

		defer func() {
			hub.unregister <- client
			conn.Close()
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println(err)
				continue
			}

			switch msg.Type {
			case "place_order":
				var order matching.Order
				if err := json.Unmarshal(msg.Data, &order); err != nil {
					log.Println(err)
					continue
				}
				me.PlaceOrders([]*matching.Order{&order})
			case "get_recent_trades":
				jsonTrades, err := json.Marshal(trades)
				if err != nil {
					log.Println(err)
					continue
				}
				client.send <- jsonTrades
			case "get_price":
				if len(trades) == 0 {
					client.send <- []byte("0")
					continue
				}
				price := trades[len(trades)-1].Price
				jsonPrice, err := json.Marshal(price)
				if err != nil {
					log.Println(err)
					continue
				}
				client.send <- jsonPrice
			}
		}
	})

	http.HandleFunc("/trades", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(trades); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/price", func(w http.ResponseWriter, r *http.Request) {
		if len(trades) == 0 {
			if err := json.NewEncoder(w).Encode(0); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		price := trades[len(trades)-1].Price
		if err := json.NewEncoder(w).Encode(price); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
