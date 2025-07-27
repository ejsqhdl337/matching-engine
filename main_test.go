package main

import (
	"log"
	"matching_engine/pkg/matching"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func TestWebsocketHandler(t *testing.T) {
	outputBuffer := matching.NewRingBuffer(1024)
	me := matching.NewMatchingEngine(outputBuffer)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s: %v", wsURL, err)
	}
	defer ws.Close()

	order := &matching.Order{ID: 1, Type: "limit", Side: "buy", Price: 100 * matching.PricePrecision, Quantity: 10}
	me.GetOrderBook().AddOrder(&matching.BookOrder{ID: order.ID, Side: order.Side, Price: order.Price, Quantity: order.Quantity})
	me.TakeSnapshot()

	var event matching.Event
	if err := ws.ReadJSON(&event); err != nil {
		t.Fatalf("could not read json from ws: %v", err)
	}

	if _, ok := event.Data.(string); !ok || !strings.HasPrefix(event.Data.(string), "SNAPSHOT:") {
		t.Errorf("Expected a snapshot event, but got %s", event.Data)
	}
}
