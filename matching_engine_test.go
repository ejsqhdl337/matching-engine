package main

import (
	"testing"
)

func TestMarketOrder(t *testing.T) {
	t.Run("should place a market order", func(t *testing.T) {
		me := newMatchingEngine()
		order := &Order{ID: 1, OrdererID: 1, Type: "market", Side: "buy", Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})

	t.Run("should partially fill a market order", func(t *testing.T) {
		me := newMatchingEngine()
		sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 5}
		me.PlaceOrder(sellOrder)

		marketOrder := &Order{ID: 2, OrdererID: 2, Type: "market", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(marketOrder)

		if len(me.Trades) != 1 {
			t.Errorf("Expected 1 trade, got %d", len(me.Trades))
		}

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}

		if me.Bids[0].Quantity != 5 {
			t.Errorf("Expected remaining bid quantity to be 5, got %d", me.Bids[0].Quantity)
		}
	})
}

func TestLimitOrder(t *testing.T) {
	t.Run("should place a limit order", func(t *testing.T) {
		me := newMatchingEngine()
		order := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(order)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})
}

func TestStopLossOrder(t *testing.T) {
	t.Run("should handle stop-loss orders", func(t *testing.T) {
		// Stop-loss orders are more complex and require a trigger price.
		// This test will be a placeholder for now.
	})
}

func TestAONOrder(t *testing.T) {
	t.Run("should execute an AON order if liquidity is sufficient", func(t *testing.T) {
		me := newMatchingEngine()
		sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 10}
		me.PlaceOrder(sellOrder)

		aonOrder := &Order{ID: 2, OrdererID: 2, Type: "aon", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(aonOrder)

		if len(me.Trades) != 1 {
			t.Errorf("Expected 1 trade, got %d", len(me.Trades))
		}
	})

	t.Run("should reject an AON order if liquidity is insufficient", func(t *testing.T) {
		me := newMatchingEngine()
		sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 5}
		me.PlaceOrder(sellOrder)

		aonOrder := &Order{ID: 2, OrdererID: 2, Type: "aon", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(aonOrder)

		if len(me.Trades) != 0 {
			t.Errorf("Expected 0 trades, got %d", len(me.Trades))
		}
	})
}

func TestMatchingLogic(t *testing.T) {
	me := newMatchingEngine()
	sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 10}
	me.PlaceOrder(sellOrder)

	buyOrder := &Order{ID: 2, OrdererID: 2, Type: "limit", Side: "buy", Price: 100.0, Quantity: 10}
	me.PlaceOrder(buyOrder)

	if len(me.Trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(me.Trades))
	}

	if len(me.Bids) != 0 {
		t.Errorf("Expected 0 bid orders, got %d", len(me.Bids))
	}

	if len(me.Asks) != 0 {
		t.Errorf("Expected 0 ask orders, got %d", len(me.Asks))
	}
}

func TestOrderBookSorting(t *testing.T) {
	me := newMatchingEngine()
	buyOrder1 := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "buy", Price: 100.0, Quantity: 10}
	buyOrder2 := &Order{ID: 2, OrdererID: 2, Type: "limit", Side: "buy", Price: 101.0, Quantity: 10}
	me.PlaceOrder(buyOrder1)
	me.PlaceOrder(buyOrder2)

	if me.Bids[0].Price != 101.0 {
		t.Errorf("Expected highest bid to be 101.0, got %f", me.Bids[0].Price)
	}

	sellOrder1 := &Order{ID: 3, OrdererID: 3, Type: "limit", Side: "sell", Price: 103.0, Quantity: 10}
	sellOrder2 := &Order{ID: 4, OrdererID: 4, Type: "limit", Side: "sell", Price: 102.0, Quantity: 10}
	me.PlaceOrder(sellOrder1)
	me.PlaceOrder(sellOrder2)

	if me.Asks[0].Price != 102.0 {
		t.Errorf("Expected lowest ask to be 102.0, got %f", me.Asks[0].Price)
	}
}

func TestPostOnlyOrder(t *testing.T) {
	t.Run("should place a post-only order if it does not cross the spread", func(t *testing.T) {
		me := newMatchingEngine()
		sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 101.0, Quantity: 10}
		me.PlaceOrder(sellOrder)

		postOnlyOrder := &Order{ID: 2, OrdererID: 2, Type: "post-only", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(postOnlyOrder)

		if len(me.Bids) != 1 {
			t.Errorf("Expected 1 bid order, got %d", len(me.Bids))
		}
	})

	t.Run("should reject a post-only order if it crosses the spread", func(t *testing.T) {
		me := newMatchingEngine()
		sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 10}
		me.PlaceOrder(sellOrder)

		postOnlyOrder := &Order{ID: 2, OrdererID: 2, Type: "post-only", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(postOnlyOrder)

		if len(me.Bids) != 0 {
			t.Errorf("Expected 0 bid orders, got %d", len(me.Bids))
		}
	})
}

func TestFOKOrder(t *testing.T) {
	t.Run("should execute a FOK order if liquidity is sufficient", func(t *testing.T) {
		me := newMatchingEngine()
		sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 10}
		me.PlaceOrder(sellOrder)

		fokOrder := &Order{ID: 2, OrdererID: 2, Type: "fok", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(fokOrder)

		if len(me.Trades) != 1 {
			t.Errorf("Expected 1 trade, got %d", len(me.Trades))
		}
	})

	t.Run("should reject a FOK order if liquidity is insufficient", func(t *testing.T) {
		me := newMatchingEngine()
		sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 5}
		me.PlaceOrder(sellOrder)

		fokOrder := &Order{ID: 2, OrdererID: 2, Type: "fok", Side: "buy", Price: 100.0, Quantity: 10}
		me.PlaceOrder(fokOrder)

		if len(me.Trades) != 0 {
			t.Errorf("Expected 0 trades, got %d", len(me.Trades))
		}
	})
}

func TestIOCOrder(t *testing.T) {
	me := newMatchingEngine()
	sellOrder := &Order{ID: 1, OrdererID: 1, Type: "limit", Side: "sell", Price: 100.0, Quantity: 5}
	me.PlaceOrder(sellOrder)

	iocOrder := &Order{ID: 2, OrdererID: 2, Type: "ioc", Side: "buy", Price: 100.0, Quantity: 10}
	me.PlaceOrder(iocOrder)

	if len(me.Trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(me.Trades))
	}

	if len(me.Bids) != 0 {
		t.Errorf("Expected 0 bid orders, got %d", len(me.Bids))
	}

	if len(me.Asks) != 0 {
		t.Errorf("Expected 0 ask orders, got %d", len(me.Asks))
	}
}
