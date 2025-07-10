package models

import "time"

type Stock struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CurrentPrice float64 `json:"currentPrice"`
	Amount       int     `json:"amount"`
}

type Trader struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Money        float64        `json:"money"`
	InitialMoney float64        `json:"initialMoney"`
	Holdings     map[string]int `json:"holdings"`
}

type OrderType string
type OrderStatus string

const (
	Buy  OrderType = "buy"
	Sell OrderType = "sell"

	Open      OrderStatus = "open"
	Filled    OrderStatus = "filled"
	Cancelled OrderStatus = "cancelled"
)

type Order struct {
	ID        string      `json:"id"`
	TraderID  string      `json:"traderId"`
	StockID   string      `json:"stockId"`
	Type      OrderType   `json:"type"`
	Price     float64     `json:"price"`
	Quantity  int         `json:"quantity"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
}

type Transaction struct {
	ID         string    `json:"id"`
	BuyerID    string    `json:"buyerId"`
	SellerID   string    `json:"sellerId"`
	StockID    string    `json:"stockId"`
	Price      float64   `json:"price"`
	Quantity   int       `json:"quantity"`
	ExecutedAt time.Time `json:"executedAt"`
}

type Config struct {
	Shares  []Stock  `json:"shares"`
	Traders []Trader `json:"traders"`
}
