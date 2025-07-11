package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Amit-R328/stock-exchange/internal/models"
	"github.com/Amit-R328/stock-exchange/internal/services"
)

type Handlers struct {
	exchange *services.Exchange
}

func NewHandlers(exchange *services.Exchange) *Handlers {
	return &Handlers{
		exchange: exchange,
	}
}

// Get all stocks
func (h *Handlers) GetAllStocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stocks := h.exchange.GetAllStocks()
	json.NewEncoder(w).Encode(stocks)
}

// Get specific stock with details
func (h *Handlers) GetStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stockID := r.URL.Path[len("/api/v1/stocks/"):]

	stock, exists := h.exchange.GetStock(stockID)
	if !exists {
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}

	openOrders := h.exchange.GetOpenOrders(stockID)
	transactions := h.exchange.GetLastTransactions(stockID, 10)

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

// Place an order
func (h *Handlers) PlaceOrder(w http.ResponseWriter, r *http.Request) {
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

	// Create order with unique ID
	order := &models.Order{
		ID:        fmt.Sprintf("order-%d", time.Now().UnixNano()),
		TraderID:  orderReq.TraderID,
		StockID:   orderReq.StockID,
		Type:      orderReq.Type,
		Price:     orderReq.Price,
		Quantity:  orderReq.Quantity,
		Status:    models.Open,
		CreatedAt: time.Now(),
	}

	if err := h.exchange.PlaceOrder(order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// Cancel an order
func (h *Handlers) CancelOrder(w http.ResponseWriter, r *http.Request) {
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

	if err := h.exchange.CancelOrder(orderID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order cancelled successfully"})
}

// Get all traders (names only)
func (h *Handlers) GetAllTraders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type TraderInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	traders := h.exchange.GetAllTraders()
	traderInfos := make([]TraderInfo, 0, len(traders))

	for _, trader := range traders {
		traderInfos = append(traderInfos, TraderInfo{
			ID:   trader.ID,
			Name: trader.Name,
		})
	}

	json.NewEncoder(w).Encode(traderInfos)
}

// Get trader details
func (h *Handlers) GetTrader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := r.URL.Path[len("/api/v1/traders/"):]

	// Check if asking for transactions
	if len(path) > 13 && path[len(path)-13:] == "/transactions" {
		traderID := path[:len(path)-13]
		h.GetTraderTransactions(w, r, traderID)
		return
	}

	traderID := path

	trader, exists := h.exchange.GetTrader(traderID)
	if !exists {
		http.Error(w, "Trader not found", http.StatusNotFound)
		return
	}

	openOrders := h.exchange.GetTraderOpenOrders(traderID)

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
func (h *Handlers) GetTraderTransactions(w http.ResponseWriter, r *http.Request, traderID string) {
	_, exists := h.exchange.GetTrader(traderID)
	if !exists {
		http.Error(w, "Trader not found", http.StatusNotFound)
		return
	}

	transactions := h.exchange.GetTraderTransactions(traderID, 8)
	profitLoss := h.exchange.CalculateProfitLoss(traderID)

	response := map[string]interface{}{
		"transactions": transactions,
		"profitLoss":   profitLoss,
	}

	json.NewEncoder(w).Encode(response)
}
