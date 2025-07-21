package main

import (
	"strings"
	"testing"
)

func TestMatchingEngine(t *testing.T) {
	me := NewMatchingEngine()
	subscription := me.eventBus.Subscribe()

	// Place a limit buy order
	buyOrder1 := &Order{ID: 1, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
	me.PlaceOrder(buyOrder1)
	if me.orderBook.BestBid().Price != 100*PricePrecision {
		t.Errorf("Expected best bid to be %d, got %d", 100*PricePrecision, me.orderBook.BestBid().Price)
	}

	// Place a limit sell order that matches
	sellOrder1 := &Order{ID: 2, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
	me.PlaceOrder(sellOrder1)

	event, ok := subscription.Poll()
	if !ok {
		t.Fatal("Expected an event, but got none")
	}
	if !strings.HasPrefix(event.Data, "TRADE:") {
		t.Errorf("Expected a trade event, but got %s", event.Data)
	}
	if me.orderBook.BestBid().Quantity != 5 {
		t.Errorf("Expected best bid quantity to be 5, got %d", me.orderBook.BestBid().Quantity)
	}

	// Place a market buy order
	buyOrder2 := &Order{ID: 3, Type: "market", Side: "buy", Price: 0, Quantity: 10}
	sellOrder2 := &BookOrder{ID: 4, Side: "sell", Price: 99 * PricePrecision, Quantity: 10}
	me.orderBook.AddOrder(sellOrder2)
	me.PlaceOrder(buyOrder2)

	event, ok = subscription.Poll()
	if !ok {
		t.Fatal("Expected an event, but got none")
	}
	if !strings.HasPrefix(event.Data, "TRADE:") {
		t.Errorf("Expected a trade event, but got %s", event.Data)
	}
	if me.orderBook.BestAsk() != nil {
		t.Errorf("Expected order book to be empty, but got %v", me.orderBook.BestAsk())
	}

	// Test stop-loss order
	slOrder := &Order{ID: 5, Type: "stop-loss", Side: "sell", Price: 98 * PricePrecision, Quantity: 5}
	me.PlaceOrder(slOrder)
	if me.sellStopOrders.Len() != 1 {
		t.Errorf("Expected 1 stop-loss order, got %d", me.sellStopOrders.Len())
	}

	// Place a trade that should trigger the stop-loss
	buyOrder3 := &BookOrder{ID: 6, Side: "buy", Price: 98 * PricePrecision, Quantity: 5}
	sellOrder3 := &Order{ID: 7, Type: "limit", Side: "sell", Price: 98 * PricePrecision, Quantity: 5}
	me.orderBook.AddOrder(buyOrder3)
	me.PlaceOrder(sellOrder3)

	event, ok = subscription.Poll()
	if !ok {
		t.Fatal("Expected an event, but got none")
	}
	if !strings.HasPrefix(event.Data, "TRADE:") {
		t.Errorf("Expected a trade event, but got %s", event.Data)
	}

	event, ok = subscription.Poll()
	if !ok {
		t.Fatal("Expected an event, but got none")
	}
	if !strings.HasPrefix(event.Data, "TRADE:") {
		t.Errorf("Expected a trade event, but got %s", event.Data)
	}

	if me.sellStopOrders.Len() != 0 {
		t.Errorf("Expected 0 stop-loss orders, got %d", me.sellStopOrders.Len())
	}
}
