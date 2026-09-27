package api

import (
	"encoding/json"
	"net/http"

	"orderengine/internal/engine"
)

type server struct {
	book *engine.OrderBook
}

func RegisterRoutes(mux *http.ServeMux, book *engine.OrderBook) {
	s := &server{book: book}

	mux.HandleFunc("POST /orders", s.createOrder)
	mux.HandleFunc("GET /orderbook", s.getOrderBook)
	mux.HandleFunc("GET /trades", s.getTrades)
}

type createOrderRequest struct {
	Side     engine.Side `json:"side"`
	Price    int64       `json:"price"`
	Quantity int64       `json:"quantity"`
}

type createOrderResponse struct {
	Order  engine.Order   `json:"order"`
	Trades []engine.Trade `json:"trades"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func (s *server) createOrder(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if req.Side != engine.Buy && req.Side != engine.Sell {
		writeError(w, http.StatusBadRequest, `side must be "BUY" or "SELL"`)
		return
	}
	if req.Price <= 0 {
		writeError(w, http.StatusBadRequest, "price must be greater than 0")
		return
	}
	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	order, trades := s.book.Submit(req.Side, req.Price, req.Quantity)
	if trades == nil {
		trades = []engine.Trade{}
	}

	writeJSON(w, http.StatusCreated, createOrderResponse{Order: *order, Trades: trades})
}

func (s *server) getOrderBook(w http.ResponseWriter, r *http.Request) {
	buys, sells, _ := s.book.Snapshot()
	if buys == nil {
		buys = []engine.Order{}
	}
	if sells == nil {
		sells = []engine.Order{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"buyOrders": buys, "sellOrders": sells})
}

func (s *server) getTrades(w http.ResponseWriter, r *http.Request) {
	_, _, trades := s.book.Snapshot()
	if trades == nil {
		trades = []engine.Trade{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"trades": trades})
}
