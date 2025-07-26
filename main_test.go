package main

import (
	"bytes"
	"encoding/json"
	"log"
	"matching_engine/pkg/matching"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestOrdersHandler(t *testing.T) {
	outputBuffer := matching.NewRingBuffer(1024)
	me := matching.NewMatchingEngine(outputBuffer)

	orders := []*matching.Order{
		{ID: 1, Type: "limit", Side: "buy", Price: 100 * matching.PricePrecision, Quantity: 10},
	}
	body, _ := json.Marshal(orders)
	req, err := http.NewRequest("POST", "/orders", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var orders []*matching.Order
		if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		me.PlaceOrders(orders)
		w.WriteHeader(http.StatusAccepted)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}

	// Allow some time for the order to be processed
	time.Sleep(10 * time.Millisecond)

	if me.GetInputBufferSize() != 1 {
		t.Errorf("Expected 1 order in the input buffer, got %d", me.GetInputBufferSize())
	}
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
