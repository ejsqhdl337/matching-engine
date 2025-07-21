package main

import (
	"testing"
)

func TestOrderBook(t *testing.T) {
	config := &OrderBookConfig{MinTickSize: 1}
	ob := NewOrderBook(config)

	// Add a buy order
	buyOrder1 := &BookOrder{ID: 1, Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
	ob.AddOrder(buyOrder1)
	if ob.BestBid().Price != 100*PricePrecision {
		t.Errorf("Expected best bid to be %d, got %d", 100*PricePrecision, ob.BestBid().Price)
	}

	// Add another buy order
	buyOrder2 := &BookOrder{ID: 2, Side: "buy", Price: 101 * PricePrecision, Quantity: 5}
	ob.AddOrder(buyOrder2)
	if ob.BestBid().Price != 101*PricePrecision {
		t.Errorf("Expected best bid to be %d, got %d", 101*PricePrecision, ob.BestBid().Price)
	}

	// Add a sell order
	sellOrder1 := &BookOrder{ID: 3, Side: "sell", Price: 102 * PricePrecision, Quantity: 8}
	ob.AddOrder(sellOrder1)
	if ob.BestAsk().Price != 102*PricePrecision {
		t.Errorf("Expected best ask to be %d, got %d", 102*PricePrecision, ob.BestAsk().Price)
	}

	// Add another sell order
	sellOrder2 := &BookOrder{ID: 4, Side: "sell", Price: 101*PricePrecision + 5000, Quantity: 12}
	ob.AddOrder(sellOrder2)
	if ob.BestAsk().Price != 101*PricePrecision+5000 {
		t.Errorf("Expected best ask to be %d, got %d", 101*PricePrecision+5000, ob.BestAsk().Price)
	}

	// Remove an order
	ob.RemoveOrder(2)
	if ob.BestBid().Price != 100*PricePrecision {
		t.Errorf("Expected best bid to be %d after removal, got %d", 100*PricePrecision, ob.BestBid().Price)
	}
}
