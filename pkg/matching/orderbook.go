package matching

import (
	"container/heap"
)

const PricePrecision = 10000

type BookOrder struct {
	ID       int
	Side     string
	Price    int64
	Quantity int
}

// An Item is something we manage in a priority queue.
type Item struct {
	value    *BookOrder // The value of the item; arbitrary.
	priority int64      // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	if pq[i].priority == pq[j].priority {
		// When priorities are equal, the older order gets priority
		return pq[i].value.ID < pq[j].value.ID
	}
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

type OrderBookConfig struct {
	MinTickSize int64
}

type OrderBook struct {
	orders map[int]*Item
	bids   *PriorityQueue
	asks   *PriorityQueue
	config *OrderBookConfig
}

func NewOrderBook(config *OrderBookConfig) *OrderBook {
	bids := &PriorityQueue{}
	asks := &PriorityQueue{}
	heap.Init(bids)
	heap.Init(asks)
	return &OrderBook{
		orders: make(map[int]*Item),
		bids:   bids,
		asks:   asks,
		config: config,
	}
}

func (ob *OrderBook) AddOrder(order *BookOrder) {
	order.Price = ob.roundPrice(order.Price)
	item := &Item{
		value:    order,
		priority: order.Price,
	}
	ob.orders[order.ID] = item
	if order.Side == "buy" {
		heap.Push(ob.bids, item)
	} else {
		// For asks, we want the lowest price to have the highest priority.
		item.priority = -order.Price
		heap.Push(ob.asks, item)
	}
}

func (ob *OrderBook) roundPrice(price int64) int64 {
	return (price / ob.config.MinTickSize) * ob.config.MinTickSize
}

func (ob *OrderBook) RemoveOrder(orderID int) {
	item, ok := ob.orders[orderID]
	if !ok {
		return
	}
	delete(ob.orders, orderID)

	var pq *PriorityQueue
	if item.value.Side == "buy" {
		pq = ob.bids
	} else {
		pq = ob.asks
	}
	heap.Remove(pq, item.index)
}

func (ob *OrderBook) BestBid() *BookOrder {
	if ob.bids.Len() == 0 {
		return nil
	}
	return (*ob.bids)[0].value
}

func (ob *OrderBook) BestAsk() *BookOrder {
	if ob.asks.Len() == 0 {
		return nil
	}
	return (*ob.asks)[0].value
}
