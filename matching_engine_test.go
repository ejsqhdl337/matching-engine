package main

import (
	"strings"
	"testing"
)

func TestMatchingEngine_PlaceLimitOrder(t *testing.T) {
	me := NewMatchingEngine()
	subscription := me.eventBus.Subscribe()

	t.Run("should place a limit buy order", func(t *testing.T) {
		buyOrder := &Order{ID: 1, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(buyOrder)
		if me.orderBook.BestBid().Price != 100*PricePrecision {
			t.Errorf("Expected best bid to be %d, got %d", 100*PricePrecision, me.orderBook.BestBid().Price)
		}
	})

	t.Run("should match a limit sell order", func(t *testing.T) {
		sellOrder := &Order{ID: 2, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder)

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
	})
}

func TestMatchingEngine_PartialFill(t *testing.T) {
	me := NewMatchingEngine()
	subscription := me.eventBus.Subscribe()

	t.Run("should partially fill a limit order", func(t *testing.T) {
		buyOrder := &Order{ID: 1, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(buyOrder)

		sellOrder := &Order{ID: 2, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder)

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
	})
}

func TestMatchingEngine_MultipleFills(t *testing.T) {
	me := NewMatchingEngine()
	subscription := me.eventBus.Subscribe()

	t.Run("should fill an order with multiple trades", func(t *testing.T) {
		sellOrder1 := &Order{ID: 1, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder1)
		sellOrder2 := &Order{ID: 2, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder2)

		buyOrder := &Order{ID: 3, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(buyOrder)

		event, ok := subscription.Poll()
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

		if me.orderBook.BestAsk() != nil {
			t.Errorf("Expected order book to be empty, but got %v", me.orderBook.BestAsk())
		}
	})
}

func TestMatchingEngine_PlaceMarketOrder(t *testing.T) {
	me := NewMatchingEngine()
	subscription := me.eventBus.Subscribe()

	t.Run("should match a market buy order", func(t *testing.T) {
		sellOrder := &BookOrder{ID: 1, Side: "sell", Price: 99 * PricePrecision, Quantity: 10}
		me.orderBook.AddOrder(sellOrder)

		buyOrder := &Order{ID: 2, Type: "market", Side: "buy", Price: 0, Quantity: 10}
		me.PlaceOrder(buyOrder)

		event, ok := subscription.Poll()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if !strings.HasPrefix(event.Data, "TRADE:") {
			t.Errorf("Expected a trade event, but got %s", event.Data)
		}
		if me.orderBook.BestAsk() != nil {
			t.Errorf("Expected order book to be empty, but got %v", me.orderBook.BestAsk())
		}
	})
}

func TestMatchingEngine_PlaceStopLossOrder(t *testing.T) {
	me := NewMatchingEngine()
	subscription := me.eventBus.Subscribe()

	t.Run("should place a stop-loss order", func(t *testing.T) {
		slOrder := &Order{ID: 1, Type: "stop-loss", Side: "sell", Price: 98 * PricePrecision, Quantity: 5}
		me.PlaceOrder(slOrder)
		if me.sellStopOrders.Len() != 1 {
			t.Errorf("Expected 1 stop-loss order, got %d", me.sellStopOrders.Len())
		}
	})

	t.Run("should trigger a stop-loss order", func(t *testing.T) {
		buyOrder := &BookOrder{ID: 2, Side: "buy", Price: 98 * PricePrecision, Quantity: 5}
		sellOrder := &Order{ID: 3, Type: "limit", Side: "sell", Price: 98 * PricePrecision, Quantity: 5}
		me.orderBook.AddOrder(buyOrder)
		me.PlaceOrder(sellOrder)

		event, ok := subscription.Poll()
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
	})
}
