package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Amit-R328/stock-exchange/internal/models"
	"github.com/Amit-R328/stock-exchange/internal/services"
)

var exchange *services.Exchange

func main() {
	// Initialize exchange using the service
	exchange = services.NewExchange()

	// Load configuration
	if err := exchange.LoadConfig("config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Start price updater
	priceUpdater := services.NewPriceUpdater(exchange, 10*time.Second)
	priceUpdater.Start()

	// Setup routes
	setupRoutes()

	log.Println("Server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("- GET    /api/v1/stocks")
	log.Println("- GET    /api/v1/stocks/{id}")
	log.Println("- POST   /api/v1/orders")
	log.Println("- DELETE /api/v1/orders/{id}")
	log.Println("- GET    /api/v1/traders")
	log.Println("- GET    /api/v1/traders/{id}")
	log.Println("- GET    /api/v1/traders/{id}/transactions")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func setupRoutes() {
	http.HandleFunc("/api/v1/stocks", handleGetStocks)
	http.HandleFunc("/api/v1/stocks/", handleGetStock)
	http.HandleFunc("/api/v1/orders", handleOrders)
	http.HandleFunc("/api/v1/orders/", handleCancelOrder)
	http.HandleFunc("/api/v1/traders", handleGetTraders)
	http.HandleFunc("/api/v1/traders/", handleGetTrader)
}

func handleGetStocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stocks := exchange.GetAllStocks()
	json.NewEncoder(w).Encode(stocks)
}

func handleGetStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stockID := r.URL.Path[len("/api/v1/stocks/"):]

	stock, exists := exchange.GetStock(stockID)
	if !exists {
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}

	openOrders := exchange.GetOpenOrders(stockID)
	transactions := exchange.GetLastTransactions(stockID, 10)

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

	// Use exchange service to place order
	if err := exchange.PlaceOrder(order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

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

	if err := exchange.CancelOrder(orderID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order cancelled"})
}

func handleGetTraders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// TODO: Move this to exchange service
	type TraderInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	traders := make([]TraderInfo, 0)

	json.NewEncoder(w).Encode(traders)
}

func handleGetTrader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := r.URL.Path[len("/api/v1/traders/"):]

	// Check if asking for transactions
	if len(path) > 13 && path[len(path)-13:] == "/transactions" {
		traderID := path[:len(path)-13]
		handleGetTraderTransactions(w, r, traderID)
		return
	}

	// TODO: Implement trader details
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func handleGetTraderTransactions(w http.ResponseWriter, r *http.Request, traderID string) {
	// TODO: Implement trader transactions
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}
