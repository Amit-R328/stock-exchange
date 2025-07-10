package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)

type Stock struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CurrentPrice float64 `json:"currentPrice"`
	Amount       int     `json:"amount"`
}

type Trader struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Money    float64        `json:"money"`
	Holdings map[string]int `json:"holdings"`
}

type OrderType string

const (
	Buy  OrderType = "buy"
	Sell OrderType = "sell"
)

type Order struct {
	ID        string    `json:"id"`
	TraderID  string    `json:"traderId"`
	StockID   string    `json:"stockId"`
	Type      OrderType `json:"type"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
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

type Exchange struct {
	Stocks       map[string]*Stock
	Traders      map[string]*Trader
	BuyOrders    map[string][]*Order // stockID -> orders
	SellOrders   map[string][]*Order
	Transactions []Transaction
}

var exchange *Exchange

func main() {

	// Initialize exchange with order books
	exchange = &Exchange{
		Stocks:       make(map[string]*Stock),
		Traders:      make(map[string]*Trader),
		BuyOrders:    make(map[string][]*Order),
		SellOrders:   make(map[string][]*Order),
		Transactions: make([]Transaction, 0),
	}

	// Load configuration
	if err := loadConfig("../../config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	go priceUpdater()

	// Setup routes
	http.HandleFunc("/api/v1/stocks", handleGetStocks)
	http.HandleFunc("/api/v1/orders", handleOrders)
	http.HandleFunc("/api/v1/traders", handleGetTraders)

	log.Println("Server starting on :8080")
	log.Printf("Loaded %d stocks and %d traders", len(exchange.Stocks), len(exchange.Traders))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}

	// Load stocks
	for _, stock := range config.Shares {
		s := stock
		exchange.Stocks[s.ID] = &s
		exchange.BuyOrders[s.ID] = make([]*Order, 0)
		exchange.SellOrders[s.ID] = make([]*Order, 0)

		// Create initial sell orders from exchange
		initialOrder := &Order{
			ID:        fmt.Sprintf("init-%s", s.ID),
			TraderID:  "exchange",
			StockID:   s.ID,
			Type:      Sell,
			Price:     s.CurrentPrice,
			Quantity:  s.Amount,
			CreatedAt: time.Now(),
		}
		exchange.SellOrders[s.ID] = append(exchange.SellOrders[s.ID], initialOrder)
		log.Printf("Created initial sell order for %s: %d shares at $%.2f", s.Name, s.Amount, s.CurrentPrice)
	}

	// Load traders
	for _, trader := range config.Traders {
		t := trader
		t.Holdings = make(map[string]int)
		exchange.Traders[t.ID] = &t
	}

	return nil
}

func priceUpdater() {
	// Simple price updater - runs every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		for _, stock := range exchange.Stocks {
			// Random change between -2% and +2%
			change := (rand.Float64() - 0.5) * 0.04
			stock.CurrentPrice = stock.CurrentPrice * (1 + change)
			log.Printf("Updated %s price to $%.2f", stock.Name, stock.CurrentPrice)
		}
	}
}

func handleGetStocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for development

	stocks := make([]Stock, 0, len(exchange.Stocks))
	for _, stock := range exchange.Stocks {
		stocks = append(stocks, *stock)
	}

	json.NewEncoder(w).Encode(stocks)
}

func handleGetTraders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Just return names for privacy
	type TraderInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	traders := make([]TraderInfo, 0, len(exchange.Traders))
	for _, trader := range exchange.Traders {
		traders = append(traders, TraderInfo{ID: trader.ID, Name: trader.Name})
	}

	json.NewEncoder(w).Encode(traders)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var orderReq struct {
		TraderID string    `json:"traderId"`
		StockID  string    `json:"stockId"`
		Type     OrderType `json:"type"`
		Price    float64   `json:"price"`
		Quantity int       `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation
	trader, exists := exchange.Traders[orderReq.TraderID]
	if !exists {
		http.Error(w, "Trader not found", http.StatusBadRequest)
		return
	}

	if orderReq.Type == Buy && trader.Money < orderReq.Price*float64(orderReq.Quantity) {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	// Create order
	order := &Order{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		TraderID:  orderReq.TraderID,
		StockID:   orderReq.StockID,
		Type:      orderReq.Type,
		Price:     orderReq.Price,
		Quantity:  orderReq.Quantity,
		CreatedAt: time.Now(),
	}

	// Add to order book
	if order.Type == Buy {
		exchange.BuyOrders[order.StockID] = append(exchange.BuyOrders[order.StockID], order)
	} else {
		exchange.SellOrders[order.StockID] = append(exchange.SellOrders[order.StockID], order)
	}

	matchOrders(order.StockID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func matchOrders(stockID string) {
	buyOrders := exchange.BuyOrders[stockID]
	sellOrders := exchange.SellOrders[stockID]

	for _, buyOrder := range buyOrders {
		for _, sellOrder := range sellOrders {
			if buyOrder.Quantity > 0 && sellOrder.Quantity > 0 &&
				buyOrder.Price >= sellOrder.Price &&
				buyOrder.TraderID != sellOrder.TraderID {
				// Execute trade
				quantity := min(buyOrder.Quantity, sellOrder.Quantity)
				executeTransaction(buyOrder, sellOrder, quantity, sellOrder.Price)
			}
		}
	}
}

func executeTransaction(buyOrder, sellOrder *Order, quantity int, price float64) {
	// Update quantities
	buyOrder.Quantity -= quantity
	sellOrder.Quantity -= quantity

	// Update trader balances and holdings
	if buyOrder.TraderID != "exchange" {
		buyer := exchange.Traders[buyOrder.TraderID]
		buyer.Money -= price * float64(quantity)
		buyer.Holdings[buyOrder.StockID] += quantity
	}

	if sellOrder.TraderID != "exchange" {
		seller := exchange.Traders[sellOrder.TraderID]
		seller.Money += price * float64(quantity)
		seller.Holdings[sellOrder.StockID] -= quantity
	}

	// Create transaction record
	tx := Transaction{
		ID:         fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		BuyerID:    buyOrder.TraderID,
		SellerID:   sellOrder.TraderID,
		StockID:    buyOrder.StockID,
		Price:      price,
		Quantity:   quantity,
		ExecutedAt: time.Now(),
	}
	exchange.Transactions = append(exchange.Transactions, tx)

	// Update stock price
	if stock, exists := exchange.Stocks[buyOrder.StockID]; exists {
		stock.CurrentPrice = price
	}

	log.Printf("Transaction: %s bought %d shares of %s from %s at $%.2f",
		buyOrder.TraderID, quantity, buyOrder.StockID, sellOrder.TraderID, price)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
