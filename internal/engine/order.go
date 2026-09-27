package engine

import "time"

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

type Order struct {
	ID        uint64    `json:"id"`
	Side      Side      `json:"side"`
	Price     int64     `json:"price"`
	Quantity  int64     `json:"quantity"`
	Remaining int64     `json:"remaining"`
	CreatedAt time.Time `json:"createdAt"`

	sequence uint64
}

type Trade struct {
	ID          uint64    `json:"id"`
	BuyOrderID  uint64    `json:"buyOrderId"`
	SellOrderID uint64    `json:"sellOrderId"`
	Price       int64     `json:"price"`
	Quantity    int64     `json:"quantity"`
	ExecutedAt  time.Time `json:"executedAt"`
}
