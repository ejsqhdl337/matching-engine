package main

import (
	"testing"
)

func BenchmarkMatchingEngine_PlaceLimitOrder(b *testing.B) {
	me := NewMatchingEngine()
	for i := 0; i < b.N; i++ {
		order := &Order{ID: i, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(order)
	}
}

func BenchmarkMatchingEngine_PlaceMarketOrder(b *testing.B) {
	me := NewMatchingEngine()
	// Pre-fill the order book
	for i := 0; i < 1000; i++ {
		order := &BookOrder{ID: i, Side: "sell", Price: int64(100 + i), Quantity: 10}
		me.orderBook.AddOrder(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order := &Order{ID: 1000 + i, Type: "market", Side: "buy", Price: 0, Quantity: 1}
		me.PlaceOrder(order)
	}
}

func BenchmarkMatchingEngine_PlaceAndMatchOrder(b *testing.B) {
	me := NewMatchingEngine()
	for i := 0; i < b.N; i++ {
		buyOrder := &Order{ID: i, Type: "limit", Side: "buy", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(buyOrder)
		sellOrder := &Order{ID: i + b.N, Type: "limit", Side: "sell", Price: 100 * PricePrecision, Quantity: 10}
		me.PlaceOrder(sellOrder)
	}
}
