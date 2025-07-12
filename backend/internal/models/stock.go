package models

import (
	"sync"
	"time"
)

type Stock struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CurrentPrice float64 `json:"currentPrice"`
	Amount       int     `json:"amount"`
	mu           sync.RWMutex
}

func (s *Stock) GetPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentPrice
}

func (s *Stock) SetPrice(price float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentPrice = price
}

// PriceQuote represents a historical price point
type PriceQuote struct {
	Timestamp time.Time `json:"timestamp"`
	Price     float64   `json:"price"`
	Volume    int       `json:"volume"`
}

// PerformanceData represents trader performance at a point in time
type PerformanceData struct {
	Date           time.Time `json:"date"`
	PortfolioValue float64   `json:"portfolioValue"`
	ProfitLoss     float64   `json:"profitLoss"`
	CashBalance    float64   `json:"cashBalance"`
}

// PortfolioData represents current portfolio distribution
type PortfolioData struct {
	Holdings    []PortfolioHolding `json:"holdings"`
	TotalValue  float64            `json:"totalValue"`
	CashBalance float64            `json:"cashBalance"`
}

// PortfolioHolding represents a single holding in the portfolio
type PortfolioHolding struct {
	StockID    string  `json:"stockId"`
	StockName  string  `json:"stockName"`
	Quantity   int     `json:"quantity"`
	Value      float64 `json:"value"`
	Percentage float64 `json:"percentage"`
}

// ActivityLog represents trading activity data
type ActivityLog struct {
	Period     string  `json:"period"`
	BuyOrders  int     `json:"buyOrders"`
	SellOrders int     `json:"sellOrders"`
	Volume     int     `json:"volume"`
	Value      float64 `json:"value"`
}
