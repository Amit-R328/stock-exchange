package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Amit-R328/stock-exchange/internal/models"
)

type Exchange struct {
	Stocks       map[string]*models.Stock
	Traders      map[string]*models.Trader
	BuyOrders    map[string][]*models.Order
	SellOrders   map[string][]*models.Order
	Transactions []models.Transaction
	mu           sync.RWMutex
}

var exchange *Exchange

func main() {
	exchange = &Exchange{
		Stocks:       make(map[string]*models.Stock),
		Traders:      make(map[string]*models.Trader),
		BuyOrders:    make(map[string][]*models.Order),
		SellOrders:   make(map[string][]*models.Order),
		Transactions: make([]models.Transaction, 0),
	}

	if err := loadConfig("../../config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	go priceUpdater()

	// API routes
	http.HandleFunc("/api/v1/stocks", handleGetStocks)
	http.HandleFunc("/api/v1/stocks/", handleGetStock)
	http.HandleFunc("/api/v1/orders", handleOrders)
	http.HandleFunc("/api/v1/orders/", handleCancelOrder)
	http.HandleFunc("/api/v1/traders", handleGetTraders)
	http.HandleFunc("/api/v1/traders/", handleGetTrader)

	log.Println("Server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("- GET  /api/v1/stocks")
	log.Println("- GET  /api/v1/stocks/{id}")
	log.Println("- POST /api/v1/orders")
	log.Println("- DELETE /api/v1/orders/{id}")
	log.Println("- GET  /api/v1/traders")
	log.Println("- GET  /api/v1/traders/{id}")
	log.Println("- GET  /api/v1/traders/{id}/transactions")

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

	var config models.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}

	for _, stock := range config.Shares {
		s := stock
		exchange.Stocks[s.ID] = &s
		exchange.BuyOrders[s.ID] = make([]*models.Order, 0)
		exchange.SellOrders[s.ID] = make([]*models.Order, 0)

		initialOrder := &models.Order{
			ID:        fmt.Sprintf("init-%s", s.ID),
			TraderID:  "exchange",
			StockID:   s.ID,
			Type:      models.Sell,
			Price:     s.CurrentPrice,
			Quantity:  s.Amount,
			Status:    models.Open,
			CreatedAt: time.Now(),
		}
		exchange.SellOrders[s.ID] = append(exchange.SellOrders[s.ID], initialOrder)
	}

	for _, trader := range config.Traders {
		t := trader
		t.Holdings = make(map[string]int)
		t.InitialMoney = t.Money // Store initial money
		exchange.Traders[t.ID] = &t
	}

	return nil
}

func priceUpdater() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		exchange.mu.Lock() // Thread safety
		for _, stock := range exchange.Stocks {
			change := (rand.Float64() - 0.5) * 0.04
			oldPrice := stock.CurrentPrice
			stock.CurrentPrice = stock.CurrentPrice * (1 + change)
			// Round to 2 decimal places
			stock.CurrentPrice = float64(int(stock.CurrentPrice*100+0.5)) / 100
			if oldPrice != stock.CurrentPrice {
				log.Printf("Updated %s: $%.2f -> $%.2f", stock.Name, oldPrice, stock.CurrentPrice)
			}
		}
		exchange.mu.Unlock()
	}
}

// Get all stocks
func handleGetStocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	exchange.mu.RLock()
	defer exchange.mu.RUnlock()

	stocks := make([]models.Stock, 0, len(exchange.Stocks))
	for _, stock := range exchange.Stocks {
		stocks = append(stocks, *stock)
	}

	json.NewEncoder(w).Encode(stocks)
}

// Get specific stock with details
func handleGetStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Extract stock ID from URL
	stockID := r.URL.Path[len("/api/v1/stocks/"):]

	exchange.mu.RLock()
	defer exchange.mu.RUnlock()

	stock, exists := exchange.Stocks[stockID]
	if !exists {
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}

	// Get open orders for this stock
	openOrders := make([]models.Order, 0)
	for _, order := range exchange.BuyOrders[stockID] {
		if order.Status == models.Open {
			openOrders = append(openOrders, *order)
		}
	}
	for _, order := range exchange.SellOrders[stockID] {
		if order.Status == models.Open {
			openOrders = append(openOrders, *order)
		}
	}

	// Get last 10 transactions
	transactions := getLastTransactions(stockID, 10)

	response := map[string]interface{}{
		"id":               stock.ID,
		"name":             stock.Name,
		"currentPrice":     stock.CurrentPrice,
		"amount":           stock.Amount,
		"openOrders":       openOrders,
		"lastTransactions": transactions,
	}

	json.NewEncoder(w).Encode(response)
}

// Get specific trader details
func handleGetTrader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := r.URL.Path[len("/api/v1/traders/"):]

	// Check if asking for transactions
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	if len(path) > 13 && path[len(path)-13:] == "/transactions" {
		traderID := path[:len(path)-13]
		handleGetTraderTransactions(w, r, traderID)
		return
	}

	traderID := path

	exchange.mu.RLock()
	defer exchange.mu.RUnlock()

	trader, exists := exchange.Traders[traderID]
	if !exists {
		http.Error(w, "Trader not found", http.StatusNotFound)
		return
	}

	// Get open orders
	openOrders := make([]models.Order, 0)
	for _, orders := range exchange.BuyOrders {
		for _, order := range orders {
			if order.TraderID == traderID && order.Status == models.Open {
				openOrders = append(openOrders, *order)
			}
		}
	}
	for _, orders := range exchange.SellOrders {
		for _, order := range orders {
			if order.TraderID == traderID && order.Status == models.Open {
				openOrders = append(openOrders, *order)
			}
		}
	}

	response := map[string]interface{}{
		"id":           trader.ID,
		"name":         trader.Name,
		"money":        trader.Money,
		"initialMoney": trader.InitialMoney,
		"holdings":     trader.Holdings,
		"openOrders":   openOrders,
	}

	json.NewEncoder(w).Encode(response)
}

// Get trader transactions
func handleGetTraderTransactions(w http.ResponseWriter, r *http.Request, traderID string) {
	exchange.mu.RLock()
	defer exchange.mu.RUnlock()

	trader, exists := exchange.Traders[traderID]
	if !exists {
		http.Error(w, "Trader not found", http.StatusNotFound)
		return
	}

	// Get last 8 transactions
	transactions := make([]models.Transaction, 0)
	count := 0
	for i := len(exchange.Transactions) - 1; i >= 0 && count < 8; i-- {
		tx := exchange.Transactions[i]
		if tx.BuyerID == traderID || tx.SellerID == traderID {
			transactions = append(transactions, tx)
			count++
		}
	}

	// Calculate profit/loss
	currentValue := trader.Money
	for stockID, quantity := range trader.Holdings {
		if stock, exists := exchange.Stocks[stockID]; exists {
			currentValue += stock.CurrentPrice * float64(quantity)
		}
	}
	profitLoss := currentValue - trader.InitialMoney

	response := map[string]interface{}{
		"transactions": transactions,
		"profitLoss":   profitLoss,
	}

	json.NewEncoder(w).Encode(response)
}

// Cancel order
func handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Path[len("/api/v1/orders/"):]

	exchange.mu.Lock()
	defer exchange.mu.Unlock()

	// Find and cancel the order
	found := false
	for stockID, orders := range exchange.BuyOrders {
		for i, order := range orders {
			if order.ID == orderID && order.Status == models.Open {
				order.Status = models.Cancelled
				// Remove from active orders
				exchange.BuyOrders[stockID] = append(orders[:i], orders[i+1:]...)
				found = true
				break
			}
		}
	}

	if !found {
		for stockID, orders := range exchange.SellOrders {
			for i, order := range orders {
				if order.ID == orderID && order.Status == models.Open {
					order.Status = models.Cancelled
					exchange.SellOrders[stockID] = append(orders[:i], orders[i+1:]...)
					found = true
					break
				}
			}
		}
	}

	if !found {
		http.Error(w, "Order not found or already closed", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order cancelled"})
}

// Enhanced order handling with validation
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
		TraderID string           `json:"traderId"`
		StockID  string           `json:"stockId"`
		Type     models.OrderType `json:"type"`
		Price    float64          `json:"price"`
		Quantity int              `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	exchange.mu.Lock()
	defer exchange.mu.Unlock()

	// Validation
	trader, exists := exchange.Traders[orderReq.TraderID]
	if !exists {
		http.Error(w, "Trader not found", http.StatusBadRequest)
		return
	}

	// Check for conflicting orders
	if hasConflictingOrder(orderReq.TraderID, orderReq.StockID, orderReq.Type) {
		http.Error(w, "Cannot have both buy and sell orders for the same stock", http.StatusBadRequest)
		return
	}

	// Validate funds/holdings
	if orderReq.Type == models.Buy {
		if trader.Money < orderReq.Price*float64(orderReq.Quantity) {
			http.Error(w, "Insufficient funds", http.StatusBadRequest)
			return
		}
	} else {
		if trader.Holdings[orderReq.StockID] < orderReq.Quantity {
			http.Error(w, "Insufficient holdings", http.StatusBadRequest)
			return
		}
	}

	// Create order
	order := &models.Order{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		TraderID:  orderReq.TraderID,
		StockID:   orderReq.StockID,
		Type:      orderReq.Type,
		Price:     orderReq.Price,
		Quantity:  orderReq.Quantity,
		Status:    models.Open,
		CreatedAt: time.Now(),
	}

	// Add to order book
	if order.Type == models.Buy {
		exchange.BuyOrders[order.StockID] = append(exchange.BuyOrders[order.StockID], order)
	} else {
		exchange.SellOrders[order.StockID] = append(exchange.SellOrders[order.StockID], order)
	}

	matchOrders(order.StockID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// Check for conflicting orders
func hasConflictingOrder(traderID, stockID string, orderType models.OrderType) bool {
	if orderType == models.Buy {
		for _, order := range exchange.SellOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				return true
			}
		}
	} else {
		for _, order := range exchange.BuyOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				return true
			}
		}
	}
	return false
}

// Get last N transactions for a stock
func getLastTransactions(stockID string, limit int) []models.Transaction {
	transactions := make([]models.Transaction, 0)
	count := 0
	for i := len(exchange.Transactions) - 1; i >= 0 && count < limit; i-- {
		if exchange.Transactions[i].StockID == stockID {
			transactions = append(transactions, exchange.Transactions[i])
			count++
		}
	}
	return transactions
}

func handleGetTraders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	exchange.mu.RLock()
	defer exchange.mu.RUnlock()

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

func matchOrders(stockID string) {
	buyOrders := exchange.BuyOrders[stockID]
	sellOrders := exchange.SellOrders[stockID]

	for _, buyOrder := range buyOrders {
		if buyOrder.Status != models.Open {
			continue
		}
		for _, sellOrder := range sellOrders {
			if sellOrder.Status != models.Open {
				continue
			}
			if buyOrder.Quantity > 0 && sellOrder.Quantity > 0 &&
				buyOrder.Price >= sellOrder.Price &&
				buyOrder.TraderID != sellOrder.TraderID {
				quantity := min(buyOrder.Quantity, sellOrder.Quantity)
				executeTransaction(buyOrder, sellOrder, quantity, sellOrder.Price)
			}
		}
	}

	// Clean up filled orders
	cleanupOrders(stockID)
}

func cleanupOrders(stockID string) {
	// Remove filled buy orders
	activeBuyOrders := make([]*models.Order, 0)
	for _, order := range exchange.BuyOrders[stockID] {
		if order.Status == models.Open {
			activeBuyOrders = append(activeBuyOrders, order)
		}
	}
	exchange.BuyOrders[stockID] = activeBuyOrders

	// Remove filled sell orders
	activeSellOrders := make([]*models.Order, 0)
	for _, order := range exchange.SellOrders[stockID] {
		if order.Status == models.Open {
			activeSellOrders = append(activeSellOrders, order)
		}
	}
	exchange.SellOrders[stockID] = activeSellOrders
}

func executeTransaction(buyOrder, sellOrder *models.Order, quantity int, price float64) {
	buyOrder.Quantity -= quantity
	sellOrder.Quantity -= quantity

	if buyOrder.Quantity == 0 {
		buyOrder.Status = models.Filled
	}
	if sellOrder.Quantity == 0 {
		sellOrder.Status = models.Filled
	}

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
	tx := models.Transaction{
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
