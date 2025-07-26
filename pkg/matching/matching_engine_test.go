package matching

import (
	"testing"
)

func TestMatchingEngine_PlaceLimitOrder(t *testing.T) {
	outputBuffer := NewRingBuffer(1024)
	me := NewMatchingEngine(outputBuffer)

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

		event, ok := outputBuffer.Pop()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if trade, ok := event.Data.(Trade); !ok {
			t.Errorf("Expected a trade event, but got %v", event.Data)
		} else if trade.TakerOrderID != 2 {
			t.Errorf("Expected taker order ID to be 2, got %d", trade.TakerOrderID)
		}
		if me.orderBook.BestBid().Quantity != 5 {
			t.Errorf("Expected best bid quantity to be 5, got %d", me.orderBook.BestBid().Quantity)
		}
	})
}

func TestMatchingEngine_PartialFill(t *testing.T) {
	outputBuffer := NewRingBuffer(1024)
	me := NewMatchingEngine(outputBuffer)

	t.Run("should partially fill a limit order", func(t *testing.T) {
		buyOrder := &Order{ID: 1, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(buyOrder)

		sellOrder := &Order{ID: 2, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder)

		event, ok := outputBuffer.Pop()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if _, ok := event.Data.(Trade); !ok {
			t.Errorf("Expected a trade event, but got %v", event.Data)
		}
		if me.orderBook.BestBid().Quantity != 5 {
			t.Errorf("Expected best bid quantity to be 5, got %d", me.orderBook.BestBid().Quantity)
		}
	})
}

func TestMatchingEngine_MultipleFills(t *testing.T) {
	outputBuffer := NewRingBuffer(1024)
	me := NewMatchingEngine(outputBuffer)

	t.Run("should fill an order with multiple trades", func(t *testing.T) {
		sellOrder1 := &Order{ID: 1, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder1)
		sellOrder2 := &Order{ID: 2, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder2)

		buyOrder := &Order{ID: 3, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(buyOrder)

		for outputBuffer.Size() < 2 {
		}

		event, ok := outputBuffer.Pop()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if _, ok := event.Data.(Trade); !ok {
			t.Errorf("Expected a trade event, but got %v", event.Data)
		}

		event, ok = outputBuffer.Pop()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if _, ok := event.Data.(Trade); !ok {
			t.Errorf("Expected a trade event, but got %v", event.Data)
		}

		if me.orderBook.BestAsk() != nil {
			t.Errorf("Expected order book to be empty, but got %v", me.orderBook.BestAsk())
		}
	})
}

func TestMatchingEngine_PlaceMarketOrder(t *testing.T) {
	outputBuffer := NewRingBuffer(1024)
	me := NewMatchingEngine(outputBuffer)

	t.Run("should match a market buy order", func(t *testing.T) {
		sellOrder := &BookOrder{ID: 1, Side: "sell", Price: 99 * PricePrecision, Quantity: 10}
		me.orderBook.AddOrder(sellOrder)

		buyOrder := &Order{ID: 2, Type: "market", Side: "buy", Price: 0, Quantity: 10}
		me.PlaceOrder(buyOrder)

		event, ok := outputBuffer.Pop()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if _, ok := event.Data.(Trade); !ok {
			t.Errorf("Expected a trade event, but got %v", event.Data)
		}
		if me.orderBook.BestAsk() != nil {
			t.Errorf("Expected order book to be empty, but got %v", me.orderBook.BestAsk())
		}
	})
}

func TestMatchingEngine_PlaceStopLossOrder(t *testing.T) {
	outputBuffer := NewRingBuffer(1024)
	me := NewMatchingEngine(outputBuffer)

	t.Run("should place a stop-loss order", func(t *testing.T) {
		slOrder := &Order{ID: 1, Type: "stop-loss", Side: "sell", Price: 98 * PricePrecision, Quantity: 5}
		me.PlaceOrder(slOrder)

		if me.sellStopOrders.Len() != 1 {
			t.Errorf("Expected 1 stop-loss order, got %d", me.sellStopOrders.Len())
		}
	})

	t.Run("should trigger a stop-loss order", func(t *testing.T) {
		buyOrder := &BookOrder{ID: 2, Side: "buy", Price: 98 * PricePrecision, Quantity: 5}
		me.orderBook.AddOrder(buyOrder)
		sellOrder := &Order{ID: 3, Type: "limit", Side: "sell", Price: 98 * PricePrecision, Quantity: 5}
		me.PlaceOrder(sellOrder)

		for outputBuffer.Size() < 2 {
		}

		event, ok := outputBuffer.Pop()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if _, ok := event.Data.(Trade); !ok {
			t.Errorf("Expected a trade event, but got %v", event.Data)
		}

		event, ok = outputBuffer.Pop()
		if !ok {
			t.Fatal("Expected an event, but got none")
		}
		if _, ok := event.Data.(Trade); !ok {
			t.Errorf("Expected a trade event, but got %v", event.Data)
		}

		if me.sellStopOrders.Len() != 0 {
			t.Errorf("Expected 0 stop-loss orders, got %d", me.sellStopOrders.Len())
		}
	})
}

func TestMatchingEngine_PlaceOrders(t *testing.T) {
	outputBuffer := NewRingBuffer(1024)
	me := NewMatchingEngine(outputBuffer)

	t.Run("should place multiple orders", func(t *testing.T) {
		orders := []*Order{
			{ID: 1, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10},
			{ID: 2, Type: "limit", Side: "buy", Price: 101 * PricePrecision, Quantity: 5},
		}
		me.PlaceOrders(orders)

		if me.inputBuffer.Size() != 2 {
			t.Errorf("Expected 2 orders in the input buffer, got %d", me.inputBuffer.Size())
		}
	})
}
