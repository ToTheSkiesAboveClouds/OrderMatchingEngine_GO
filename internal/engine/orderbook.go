package engine

import (
	"container/heap"
	"sync"
	"time"
)

type OrderBook struct {
	mu    sync.Mutex
	buys  buyHeap
	sells sellHeap

	trades []Trade

	nextOrderID uint64
	nextTradeID uint64
	sequence    uint64
}

func NewOrderBook() *OrderBook {
	ob := &OrderBook{}
	heap.Init(&ob.buys)
	heap.Init(&ob.sells)
	return ob
}

func (ob *OrderBook) Submit(side Side, price int64, quantity int64) (*Order, []Trade) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.nextOrderID++
	ob.sequence++
	order := &Order{
		ID:        ob.nextOrderID,
		Side:      side,
		Price:     price,
		Quantity:  quantity,
		Remaining: quantity,
		CreatedAt: time.Now(),
		sequence:  ob.sequence,
	}

	var newTrades []Trade

	if side == Buy {
		for order.Remaining > 0 && ob.sells.Len() > 0 && ob.sells[0].Price <= order.Price {
			best := ob.sells[0]
			newTrades = append(newTrades, ob.match(order, best, best.Price))
			if best.Remaining == 0 {
				heap.Pop(&ob.sells)
			}
		}
		if order.Remaining > 0 {
			heap.Push(&ob.buys, order)
		}
	} else {
		for order.Remaining > 0 && ob.buys.Len() > 0 && ob.buys[0].Price >= order.Price {
			best := ob.buys[0]
			newTrades = append(newTrades, ob.match(best, order, best.Price))
			if best.Remaining == 0 {
				heap.Pop(&ob.buys)
			}
		}
		if order.Remaining > 0 {
			heap.Push(&ob.sells, order)
		}
	}

	ob.trades = append(ob.trades, newTrades...)
	return order, newTrades
}
func (ob *OrderBook) match(buyOrder, sellOrder *Order, price int64) Trade {
	qty := buyOrder.Remaining
	if sellOrder.Remaining < qty {
		qty = sellOrder.Remaining
	}

	buyOrder.Remaining -= qty
	sellOrder.Remaining -= qty

	ob.nextTradeID++
	return Trade{
		ID:          ob.nextTradeID,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		Price:       price,
		Quantity:    qty,
		ExecutedAt:  time.Now(),
	}
}

func (ob *OrderBook) Snapshot() (buys []Order, sells []Order, trades []Trade) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	for _, o := range ob.buys {
		buys = append(buys, *o)
	}
	for _, o := range ob.sells {
		sells = append(sells, *o)
	}
	trades = append(trades, ob.trades...)
	return
}
