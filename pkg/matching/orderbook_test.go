package matching

import (
	"testing"
)

func TestOrderBook_AddAndRemoveOrders(t *testing.T) {
	config := &OrderBookConfig{MinTickSize: 1}
	ob := NewOrderBook(config)

	t.Run("should add buy and sell orders", func(t *testing.T) {
		buyOrder := &BookOrder{ID: 1, Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		ob.AddOrder(buyOrder)
		if ob.BestBid().Price != 100*PricePrecision {
			t.Errorf("Expected best bid to be %d, got %d", 100*PricePrecision, ob.BestBid().Price)
		}

		sellOrder := &BookOrder{ID: 2, Side: "sell", Price: 101 * PricePrecision, Quantity: 5}
		ob.AddOrder(sellOrder)
		if ob.BestAsk().Price != 101*PricePrecision {
			t.Errorf("Expected best ask to be %d, got %d", 101*PricePrecision, ob.BestAsk().Price)
		}
	})

	t.Run("should remove an order", func(t *testing.T) {
		buyOrder := &BookOrder{ID: 3, Side: "buy", Price: 99 * PricePrecision, Quantity: 10}
		ob.AddOrder(buyOrder)
		ob.RemoveOrder(3)
		if ob.BestBid().Price == 99*PricePrecision {
			t.Error("Expected order to be removed")
		}
	})
}

func TestOrderBook_BestBidAndAsk(t *testing.T) {
	config := &OrderBookConfig{MinTickSize: 1}
	ob := NewOrderBook(config)

	t.Run("should return correct best bid and ask", func(t *testing.T) {
		buyOrder1 := &BookOrder{ID: 1, Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		ob.AddOrder(buyOrder1)
		buyOrder2 := &BookOrder{ID: 2, Side: "buy", Price: 101 * PricePrecision, Quantity: 5}
		ob.AddOrder(buyOrder2)
		if ob.BestBid().Price != 101*PricePrecision {
			t.Errorf("Expected best bid to be %d, got %d", 101*PricePrecision, ob.BestBid().Price)
		}

		sellOrder1 := &BookOrder{ID: 3, Side: "sell", Price: 102 * PricePrecision, Quantity: 8}
		ob.AddOrder(sellOrder1)
		sellOrder2 := &BookOrder{ID: 4, Side: "sell", Price: 101*PricePrecision - 5000, Quantity: 12}
		ob.AddOrder(sellOrder2)
		if ob.BestAsk().Price != 101*PricePrecision-5000 {
			t.Errorf("Expected best ask to be %d, got %d", 101*PricePrecision-5000, ob.BestAsk().Price)
		}
	})

	t.Run("should return nil for empty order book", func(t *testing.T) {
		config := &OrderBookConfig{MinTickSize: 1}
		ob := NewOrderBook(config)
		if ob.BestBid() != nil {
			t.Error("Expected best bid to be nil")
		}
		if ob.BestAsk() != nil {
			t.Error("Expected best ask to be nil")
		}
	})
}

func TestOrderBook_RoundPrice(t *testing.T) {
	config := &OrderBookConfig{MinTickSize: 100}
	ob := NewOrderBook(config)

	t.Run("should round price down to nearest tick size", func(t *testing.T) {
		price := int64(12345)
		roundedPrice := ob.roundPrice(price)
		if roundedPrice != 12300 {
			t.Errorf("Expected price to be rounded to 12300, got %d", roundedPrice)
		}
	})
}
