package main

import (
	"testing"
)

type Order struct {
	ID        int
	OrdererID int
	Type      string // "market", "limit", "stop-loss", "post-only", "aon", "fok", "ioc"
	Side      string // "buy", "sell"
	Price     float64
	Quantity  int
}

type MatchingEngine struct {
	Bids []*Order
	Asks []*Order
}

func (me *MatchingEngine) PlaceOrder(order *Order) {
	// Basic implementation for testing purposes
	if order.Side == "buy" {
		me.Bids = append(me.Bids, order)
	} else {
		me.Asks = append(me.Asks, order)
	}
}

func TestMatchingEngine(t *testing.T) {
	t.Run("MarketOrder", func(t *testing.T) {
		me := &MatchingEngine{}
		order := &Order{ID: 1, OrdererID: 1, Type: "market", Side: "buy", Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})

	t.Run("LimitOrder", func(t *testing.T) {
		me := &MatchingEngine{}
		order := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})

	t.Run("StopLossOrder", func(t *testing.T) {
		// Stop-loss orders are more complex and require a trigger price.
		// This test will be a placeholder for now.
	})

	t.Run("PostOnlyOrder", func(t *testing.T) {
		me := &MatchingEngine{}
		// This order should be rejected if it matches immediately.
		// For now, we'll just place it.
		order := &Order{ID: 1, OrdererID: 1, Type: "post-only", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})

	t.Run("AllOrNoneOrder", func(t *testing.T) {
		me := &MatchingEngine{}
		order := &Order{ID: 1, OrdererID: 1, Type: "aon", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})

	t.Run("FillOrKillOrder", func(t *testing.T) {
		me := &MatchingEngine{}
		order := &Order{ID: 1, OrdererID: 1, Type: "fok", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})

	t.Run("ImmediateOrCancelOrder", func(t *testing.T) {
		me := &MatchingEngine{}
		order := &Order{ID: 1, OrdererID: 1, Type: "ioc", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})
}
