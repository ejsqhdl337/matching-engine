package main

import (
	"container/heap"
	"fmt"
)

type Order struct {
	ID        int
	OrdererID int
	Type      string // "market", "limit", "stop-loss", "post-only", "aon", "fok", "ioc"
	Side      string // "buy", "sell"
	Price     int64
	Quantity  int
}

type Trade struct {
	TakerOrderID int
	MakerOrderID int
	Price        int64
	Quantity     int
}

// A StopLossOrder is something we manage in a priority queue.
type StopLossOrder struct {
	value    *Order // The value of the item; arbitrary.
	priority int64  // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A StopLossQueue implements heap.Interface and holds StopLossOrders.
type StopLossQueue []*StopLossOrder

func (pq StopLossQueue) Len() int { return len(pq) }

func (pq StopLossQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority > pq[j].priority
}

func (pq StopLossQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *StopLossQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*StopLossOrder)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *StopLossQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

type MatchingEngine struct {
	orderBook      *OrderBook
	buyStopOrders  *StopLossQueue
	sellStopOrders *StopLossQueue
	inputBuffer    *RingBuffer
	outputBuffer   *RingBuffer
}

func NewMatchingEngine(outputBuffer *RingBuffer) *MatchingEngine {
	buyStopOrders := &StopLossQueue{}
	sellStopOrders := &StopLossQueue{}
	heap.Init(buyStopOrders)
	heap.Init(sellStopOrders)
	return &MatchingEngine{
		orderBook:      NewOrderBook(&OrderBookConfig{MinTickSize: 1}),
		buyStopOrders:  buyStopOrders,
		sellStopOrders: sellStopOrders,
		inputBuffer:    NewRingBuffer(1024),
		outputBuffer:   outputBuffer,
	}
}

func (me *MatchingEngine) Run() {
	for {
		_, ok := me.inputBuffer.Pop()
		if !ok {
			continue
		}
		// me.PlaceOrder(event.Order)
	}
}

func (me *MatchingEngine) PlaceOrders(orders []*Order) {
	for _, order := range orders {
		for !me.inputBuffer.Push(Event{Order: order}) {
			// Keep trying until the push is successful
		}
	}
}

func (me *MatchingEngine) PlaceOrder(order *Order) {
	if order.Type == "stop-loss" {
		item := &StopLossOrder{
			value:    order,
			priority: order.Price,
		}
		if order.Side == "buy" {
			heap.Push(me.buyStopOrders, item)
		} else {
			item.priority = -order.Price
			heap.Push(me.sellStopOrders, item)
		}
		return
	}

	if order.Type == "market" {
		me.matchMarketOrder(order)
	} else {
		me.matchLimitOrder(order)
	}
	me.triggerStopLossOrders(order.Price)
}

func (me *MatchingEngine) triggerStopLossOrders(currentPrice int64) {
	// Trigger sell stop-loss orders
	for me.sellStopOrders.Len() > 0 && -(*me.sellStopOrders)[0].priority <= currentPrice {
		slOrder := heap.Pop(me.sellStopOrders).(*StopLossOrder).value
		marketOrder := &Order{
			ID:        slOrder.ID,
			OrdererID: slOrder.OrdererID,
			Type:      "market",
			Side:      slOrder.Side,
			Quantity:  slOrder.Quantity,
		}
		me.matchMarketOrder(marketOrder)
	}

	// Trigger buy stop-loss orders
	for me.buyStopOrders.Len() > 0 && (*me.buyStopOrders)[0].priority >= currentPrice {
		slOrder := heap.Pop(me.buyStopOrders).(*StopLossOrder).value
		marketOrder := &Order{
			ID:        slOrder.ID,
			OrdererID: slOrder.OrdererID,
			Type:      "market",
			Side:      slOrder.Side,
			Quantity:  slOrder.Quantity,
		}
		me.matchMarketOrder(marketOrder)
	}
}

func (me *MatchingEngine) matchMarketOrder(order *Order) {
	if order.Side == "buy" {
		for order.Quantity > 0 && me.orderBook.BestAsk() != nil {
			bestAsk := me.orderBook.BestAsk()
			if order.Quantity >= bestAsk.Quantity {
				me.executeTrade(order, bestAsk, bestAsk.Price)
				order.Quantity -= bestAsk.Quantity
				me.orderBook.RemoveOrder(bestAsk.ID)
			} else {
				me.executeTrade(order, bestAsk, bestAsk.Price)
				bestAsk.Quantity -= order.Quantity
				order.Quantity = 0
			}
		}
	} else { // order.Side == "sell"
		for order.Quantity > 0 && me.orderBook.BestBid() != nil {
			bestBid := me.orderBook.BestBid()
			if order.Quantity >= bestBid.Quantity {
				me.executeTrade(order, bestBid, bestBid.Price)
				order.Quantity -= bestBid.Quantity
				me.orderBook.RemoveOrder(bestBid.ID)
			} else {
				me.executeTrade(order, bestBid, bestBid.Price)
				bestBid.Quantity -= order.Quantity
				order.Quantity = 0
			}
		}
	}
}

func (me *MatchingEngine) matchLimitOrder(order *Order) {
	if order.Side == "buy" {
		for order.Quantity > 0 && me.orderBook.BestAsk() != nil && order.Price >= me.orderBook.BestAsk().Price {
			bestAsk := me.orderBook.BestAsk()
			if order.Quantity >= bestAsk.Quantity {
				me.executeTrade(order, bestAsk, bestAsk.Price)
				order.Quantity -= bestAsk.Quantity
				me.orderBook.RemoveOrder(bestAsk.ID)
			} else {
				me.executeTrade(order, bestAsk, bestAsk.Price)
				bestAsk.Quantity -= order.Quantity
				order.Quantity = 0
			}
		}
	} else { // order.Side == "sell"
		for order.Quantity > 0 && me.orderBook.BestBid() != nil && order.Price <= me.orderBook.BestBid().Price {
			bestBid := me.orderBook.BestBid()
			if order.Quantity >= bestBid.Quantity {
				me.executeTrade(order, bestBid, bestBid.Price)
				order.Quantity -= bestBid.Quantity
				me.orderBook.RemoveOrder(bestBid.ID)
			} else {
				me.executeTrade(order, bestBid, bestBid.Price)
				bestBid.Quantity -= order.Quantity
				order.Quantity = 0
			}
		}
	}

	if order.Quantity > 0 {
		bookOrder := &BookOrder{
			ID:       order.ID,
			Side:     order.Side,
			Price:    order.Price,
			Quantity: order.Quantity,
		}
		me.orderBook.AddOrder(bookOrder)
	}
}

func (me *MatchingEngine) executeTrade(takerOrder *Order, makerOrder *BookOrder, price int64) {
	trade := Trade{
		TakerOrderID: takerOrder.ID,
		MakerOrderID: makerOrder.ID,
		Price:        price,
		Quantity:     takerOrder.Quantity,
	}
	if takerOrder.Quantity > makerOrder.Quantity {
		trade.Quantity = makerOrder.Quantity
	}

	me.outputBuffer.Push(Event{Data: fmt.Sprintf("TRADE: %v", trade)})
	me.triggerStopLossOrders(price)
}

func (me *MatchingEngine) TakeSnapshot() {
	snapshot := make([]string, 0)
	for _, item := range *me.orderBook.bids {
		order := item.value
		snapshot = append(snapshot, fmt.Sprintf("BID: %d, %d, %d", order.ID, order.Price, order.Quantity))
	}
	for _, item := range *me.orderBook.asks {
		order := item.value
		snapshot = append(snapshot, fmt.Sprintf("ASK: %d, %d, %d", order.ID, order.Price, order.Quantity))
	}
	me.outputBuffer.Push(Event{Data: fmt.Sprintf("SNAPSHOT: %v", snapshot)})
}
