# Order Matching Engine

A small, in-memory order matching engine (the core of any exchange/trading
platform) with a REST API on top. Built as a portfolio project to show Go
specifically for what it's good at: safe, concurrent processing of a stream
of events — not "yet another CRUD API".

## What it actually does

You submit limit orders — "I want to BUY 5 units at price 100" or "SELL 3
units at price 105" — and the engine matches them against each other using
**price-time priority**, the same rule real exchanges use:

- A buy order matches the *cheapest* resting sell order whose price it can
  afford.
- A sell order matches the *most expensive* resting buy order that's willing
  to pay its price.
- If several resting orders share the same price, the **oldest one wins**
  (first come, first served).
- An order can be **partially filled**: if you want to buy 10 but only 3 are
  available at a matchable price, 3 get traded immediately and 7 keep
  resting in the book, waiting for the next matching sell order.

Every match produces a `Trade` record (who bought, who sold, at what price,
how much).

## Why this is the interesting part, concurrency-wise

Multiple people submitting orders at the same time is not an edge case for
an exchange — it's the normal case. `internal/engine/orderbook.go` guards
all book state with a single `sync.Mutex`, so two goroutines calling
`Submit` at once can never corrupt the order book.

`internal/engine/orderbook_test.go` has `TestConcurrentSubmit`, which fires
1000 orders from 50 goroutines simultaneously with no coordination between
them, then checks the numbers add up (nothing lost, nothing double-counted).
Run it with the race detector on:

```bash
go test ./... -race -v
```

If the locking were missing or wrong, this is the test (and flag) that would
catch it — that's the whole point of writing it this way instead of just
"it seems to work when I click around".

## Project layout

```
order-matching-engine/
├── main.go                        # starts the HTTP server
├── internal/
│   ├── engine/
│   │   ├── order.go                # Order and Trade types
│   │   ├── heap.go                 # price-time-priority heaps (buy/sell)
│   │   ├── orderbook.go            # the matching logic + the mutex
│   │   └── orderbook_test.go       # unit tests + the concurrency test
│   └── api/
│       └── handlers.go             # REST handlers (net/http, no framework)
└── go.mod
```

No third-party dependencies — only the Go standard library. That's a
deliberate choice, not a limitation: it means `go run .` works the instant
Go is installed, with nothing else to fetch or configure.

## Running it

```bash
go run .
```

You should see:

```
order matching engine listening on http://localhost:8080
```

## Trying it out

In a second terminal (or with a REST client like Postman/Insomnia/the VS
Code "REST Client" extension):

```bash
# Rest a sell order: someone wants to sell 5 units at price 100
curl -X POST localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"side":"SELL","price":100,"quantity":5}'

# A buy order for 3 units at 100 matches it immediately (partial fill)
curl -X POST localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"side":"BUY","price":100,"quantity":3}'

# See what's still resting in the book (2 units of the sell order left)
curl localhost:8080/orderbook

# See the full trade history
curl localhost:8080/trades
```

## Testing

```bash
go vet ./...          # static checks
go test ./... -race -v # unit tests + the concurrency test, with the race detector
```

## Go concepts this project deliberately exercises

If you're new to Go (like I was when I wrote this), these are the parts
worth understanding line by line rather than skimming:

- **Goroutines & `sync.WaitGroup`** (`orderbook_test.go`) — how you spin up
  many lightweight concurrent workers and wait for all of them to finish.
- **`sync.Mutex`** (`orderbook.go`) — the lock that makes `Submit` safe to
  call from many goroutines at once.
- **`container/heap`** (`heap.go`) — Go's generic-ish way of maintaining a
  priority queue; here it's what makes "find the best resting order" fast.
- **Structs, methods and pointer receivers** — `Order`, `Trade`, `OrderBook`
  and how methods like `(ob *OrderBook) Submit(...)` mutate shared state.
- **`net/http` without a framework** — Go 1.22+'s `ServeMux` supports
  method-aware routes (`"POST /orders"`) directly, which is enough for an
  API this size and is worth understanding before reaching for something
  like Gin or Echo on a bigger project.
- **Unexported fields and JSON tags** (`order.go`) — `sequence` is lowercase
  on purpose, so `encoding/json` never serializes it into API responses.

## Honest limitations (a next-step list, not hidden gaps)

- **In-memory only** — restart the process and the book is empty. A real
  version would append every order/trade to a database or an event log.
- **No authentication** — anyone can submit orders as anyone.
- **No persistence/replay** — no way to reconstruct book state after a
  crash. The natural fix is writing every accepted order to a durable log
  (a database table, or Kafka/RabbitMQ) before acknowledging it.
- **No Docker** yet — trivial to add (a single-stage `Dockerfile` building
  a static Go binary), planned as a follow-up.
