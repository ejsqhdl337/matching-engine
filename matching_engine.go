package main

import (
	"sort"
)

type Order struct {
	ID        int
	OrdererID int
	Type      string // "market", "limit", "stop-loss", "post-only", "aon", "fok", "ioc"
	Side      string // "buy", "sell"
	Price     float64
	Quantity  int
}

type Trade struct {
	TakerOrderID int
	MakerOrderID int
	Price        float64
	Quantity     int
}

type MatchingEngine struct {
	Bids   []*Order
	Asks   []*Order
	Trades []Trade
}

func (me *MatchingEngine) PlaceOrder(order *Order) {
	if order.Type == "post-only" {
		if order.Side == "buy" {
			if len(me.Asks) > 0 && me.Asks[0].Price <= order.Price {
				return // Reject order
			}
		} else { // order.Side == "sell"
			if len(me.Bids) > 0 && me.Bids[0].Price >= order.Price {
				return // Reject order
			}
		}
	}


		if order.Type == "fok" || order.Type == "aon" {
		if order.Side == "buy" {
			var totalQuantity int
			for _, ask := range me.Asks {
				if ask.Price <= order.Price {
					totalQuantity += ask.Quantity
				}
			}
			if totalQuantity < order.Quantity {
				return // Reject order
			}
		} else { // order.Side == "sell"
			var totalQuantity int
			for _, bid := range me.Bids {
				if bid.Price >= order.Price {
					totalQuantity += bid.Quantity
				}
			}
			if totalQuantity < order.Quantity {
				return // Reject order
			}
		}
	}

	if order.Side == "buy" {
		if len(me.Asks) > 0 && me.Asks[0].Price <= order.Price {
			// Match with existing sell orders
			for i := 0; i < len(me.Asks); i++ {
				if me.Asks[i].Price <= order.Price {
					if order.Quantity >= me.Asks[i].Quantity {
						trade := Trade{
							TakerOrderID: order.ID,
							MakerOrderID: me.Asks[i].ID,
							Price:        me.Asks[i].Price,
							Quantity:     me.Asks[i].Quantity,
						}
						me.Trades = append(me.Trades, trade)
						order.Quantity -= me.Asks[i].Quantity
						me.Asks = append(me.Asks[:i], me.Asks[i+1:]...)
						i--
					} else {
						trade := Trade{
							TakerOrderID: order.ID,
							MakerOrderID: me.Asks[i].ID,
							Price:        me.Asks[i].Price,
							Quantity:     order.Quantity,
						}
						me.Trades = append(me.Trades, trade)
						me.Asks[i].Quantity -= order.Quantity
						order.Quantity = 0
						break
					}
				}
			}
		}
		if order.Quantity > 0 {
			if order.Type != "ioc" {
				me.Bids = append(me.Bids, order)
				sort.Slice(me.Bids, func(i, j int) bool {
					return me.Bids[i].Price > me.Bids[j].Price
				})
			}
		}
	} else { // order.Side == "sell"
		if len(me.Bids) > 0 && me.Bids[0].Price >= order.Price {
			// Match with existing buy orders
			for i := 0; i < len(me.Bids); i++ {
				if me.Bids[i].Price >= order.Price {
					if order.Quantity >= me.Bids[i].Quantity {
						trade := Trade{
							TakerOrderID: order.ID,
							MakerOrderID: me.Bids[i].ID,
							Price:        me.Bids[i].Price,
							Quantity:     me.Bids[i].Quantity,
						}
						me.Trades = append(me.Trades, trade)
						order.Quantity -= me.Bids[i].Quantity
						me.Bids = append(me.Bids[:i], me.Bids[i+1:]...)
						i--
					} else {
						trade := Trade{
							TakerOrderID: order.ID,
							MakerOrderID: me.Bids[i].ID,
							Price:        me.Bids[i].Price,
							Quantity:     order.Quantity,
						}
						me.Trades = append(me.Trades, trade)
						me.Bids[i].Quantity -= order.Quantity
						order.Quantity = 0
						break
					}
				}
			}
		}
		if order.Quantity > 0 {
			if order.Type != "ioc" {
				me.Asks = append(me.Asks, order)
				sort.Slice(me.Asks, func(i, j int) bool {
					return me.Asks[i].Price < me.Asks[j].Price
				})
			}
		}
	}
}

func newMatchingEngine() *MatchingEngine {
	return &MatchingEngine{}
}
