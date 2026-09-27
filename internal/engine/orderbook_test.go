package engine

import (
	"sync"
	"testing"
)

func TestConcurrentSubmit(t *testing.T) {
	book := NewOrderBook()

	const workers = 50
	const ordersPerWorker = 20

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < ordersPerWorker; i++ {
				side := Buy
				if (worker+i)%2 == 0 {
					side = Sell
				}

				book.Submit(side, 100, 1)
			}
		}(w)
	}

	wg.Wait()

	buys, sells, trades := book.Snapshot()

	var restingQty int64
	for _, o := range buys {
		restingQty += o.Remaining
	}
	for _, o := range sells {
		restingQty += o.Remaining
	}

	var tradedQty int64
	for _, tr := range trades {
		tradedQty += tr.Quantity
	}

	submittedQty := int64(workers * ordersPerWorker)

	if got, want := restingQty+2*tradedQty, submittedQty; got != want {
		t.Fatalf("quantity accounting mismatch: submitted=%d resting=%d traded=%d (resting+2*traded=%d)",
			submittedQty, restingQty, tradedQty, got)
	}
}

func TestPriceTimePriority(t *testing.T) {
	book := NewOrderBook()

	first, _ := book.Submit(Sell, 100, 5)
	book.Submit(Sell, 100, 5)

	_, trades := book.Submit(Buy, 100, 5)

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].SellOrderID != first.ID {
		t.Fatalf("expected the first resting order (id=%d) to match first, got id=%d",
			first.ID, trades[0].SellOrderID)
	}
}

func TestPartialFill(t *testing.T) {
	book := NewOrderBook()

	book.Submit(Sell, 100, 3)
	_, trades := book.Submit(Buy, 100, 10)

	if len(trades) != 1 || trades[0].Quantity != 3 {
		t.Fatalf("expected a single trade of quantity 3, got %+v", trades)
	}

	buys, sells, _ := book.Snapshot()
	if len(sells) != 0 {
		t.Fatalf("expected the sell side to be empty, got %d resting order(s)", len(sells))
	}
	if len(buys) != 1 || buys[0].Remaining != 7 {
		t.Fatalf("expected 7 remaining on the buy side, got %+v", buys)
	}
}

func TestNoCrossNoTrade(t *testing.T) {
	book := NewOrderBook()

	book.Submit(Sell, 110, 5)             // wants at least 110
	_, trades := book.Submit(Buy, 100, 5) // only willing to pay 100

	if len(trades) != 0 {
		t.Fatalf("expected no trade when prices don't cross, got %+v", trades)
	}

	buys, sells, _ := book.Snapshot()
	if len(buys) != 1 || len(sells) != 1 {
		t.Fatalf("expected both orders resting, got buys=%d sells=%d", len(buys), len(sells))
	}
}
