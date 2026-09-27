package engine

type buyHeap []*Order

func (h buyHeap) Len() int { return len(h) }

func (h buyHeap) Less(i, j int) bool {
	if h[i].Price != h[j].Price {
		return h[i].Price > h[j].Price
	}
	return h[i].sequence < h[j].sequence
}

func (h buyHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *buyHeap) Push(x any) { *h = append(*h, x.(*Order)) }

func (h *buyHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}

type sellHeap []*Order

func (h sellHeap) Len() int { return len(h) }

func (h sellHeap) Less(i, j int) bool {
	if h[i].Price != h[j].Price {
		return h[i].Price < h[j].Price
	}
	return h[i].sequence < h[j].sequence
}

func (h sellHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *sellHeap) Push(x any) { *h = append(*h, x.(*Order)) }

func (h *sellHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}
