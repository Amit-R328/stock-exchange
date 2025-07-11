package services

import (
	"encoding/json"
	"fmt"
	"log"
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

func NewExchange() *Exchange {
	return &Exchange{
		Stocks:       make(map[string]*models.Stock),
		Traders:      make(map[string]*models.Trader),
		BuyOrders:    make(map[string][]*models.Order),
		SellOrders:   make(map[string][]*models.Order),
		Transactions: make([]models.Transaction, 0),
	}
}

func (e *Exchange) LoadConfig(filename string) error {
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
		e.Stocks[s.ID] = &s
		e.BuyOrders[s.ID] = make([]*models.Order, 0)
		e.SellOrders[s.ID] = make([]*models.Order, 0)

		// Create initial sell orders from exchange
		initialOrder := &models.Order{
			ID:        fmt.Sprintf("init-%s-%d", s.ID, time.Now().Unix()),
			TraderID:  "exchange",
			StockID:   s.ID,
			Type:      models.Sell,
			Price:     s.CurrentPrice,
			Quantity:  s.Amount,
			Status:    models.Open,
			CreatedAt: time.Now(),
		}
		e.SellOrders[s.ID] = append(e.SellOrders[s.ID], initialOrder)
		log.Printf("Created initial sell order for %s: %d shares at $%.2f", s.Name, s.Amount, s.CurrentPrice)
	}

	for _, trader := range config.Traders {
		t := trader
		t.Holdings = make(map[string]int)
		t.InitialMoney = t.Money
		e.Traders[t.ID] = &t
	}

	return nil
}

func (e *Exchange) GetAllStocks() []models.Stock {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stocks := make([]models.Stock, 0, len(e.Stocks))
	for _, stock := range e.Stocks {
		stocks = append(stocks, *stock)
	}
	return stocks
}

func (e *Exchange) GetStock(stockID string) (*models.Stock, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stock, exists := e.Stocks[stockID]
	return stock, exists
}

func (e *Exchange) GetOpenOrders(stockID string) []models.Order {
	e.mu.RLock()
	defer e.mu.RUnlock()

	orders := make([]models.Order, 0)

	// Buy orders
	for _, order := range e.BuyOrders[stockID] {
		if order.Status == models.Open {
			orders = append(orders, *order)
		}
	}

	// Sell orders
	for _, order := range e.SellOrders[stockID] {
		if order.Status == models.Open {
			orders = append(orders, *order)
		}
	}

	return orders
}

func (e *Exchange) GetLastTransactions(stockID string, limit int) []models.Transaction {
	e.mu.RLock()
	defer e.mu.RUnlock()

	transactions := make([]models.Transaction, 0)
	count := 0

	for i := len(e.Transactions) - 1; i >= 0 && count < limit; i-- {
		if e.Transactions[i].StockID == stockID {
			transactions = append(transactions, e.Transactions[i])
			count++
		}
	}

	return transactions
}

// New method: Get all traders
func (e *Exchange) GetAllTraders() []models.Trader {
	e.mu.RLock()
	defer e.mu.RUnlock()

	traders := make([]models.Trader, 0, len(e.Traders))
	for _, trader := range e.Traders {
		traders = append(traders, *trader)
	}
	return traders
}

// New method: Get specific trader
func (e *Exchange) GetTrader(traderID string) (*models.Trader, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	trader, exists := e.Traders[traderID]
	return trader, exists
}

// New method: Get trader's open orders
func (e *Exchange) GetTraderOpenOrders(traderID string) []models.Order {
	e.mu.RLock()
	defer e.mu.RUnlock()

	orders := make([]models.Order, 0)

	// Check all stocks for this trader's orders
	for _, buyOrders := range e.BuyOrders {
		for _, order := range buyOrders {
			if order.TraderID == traderID && order.Status == models.Open {
				orders = append(orders, *order)
			}
		}
	}

	for _, sellOrders := range e.SellOrders {
		for _, order := range sellOrders {
			if order.TraderID == traderID && order.Status == models.Open {
				orders = append(orders, *order)
			}
		}
	}

	return orders
}

// New method: Get trader's transactions
func (e *Exchange) GetTraderTransactions(traderID string, limit int) []models.Transaction {
	e.mu.RLock()
	defer e.mu.RUnlock()

	transactions := make([]models.Transaction, 0)
	count := 0

	// Go through transactions in reverse order (newest first)
	for i := len(e.Transactions) - 1; i >= 0 && count < limit; i-- {
		tx := e.Transactions[i]
		if tx.BuyerID == traderID || tx.SellerID == traderID {
			transactions = append(transactions, tx)
			count++
		}
	}

	return transactions
}

// New method: Calculate profit/loss for a trader
func (e *Exchange) CalculateProfitLoss(traderID string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	trader, exists := e.Traders[traderID]
	if !exists {
		return 0
	}

	// Current value = cash + portfolio value
	currentValue := trader.Money

	// Add value of all holdings
	for stockID, quantity := range trader.Holdings {
		if stock, exists := e.Stocks[stockID]; exists {
			currentValue += stock.CurrentPrice * float64(quantity)
		}
	}

	// Profit/Loss = current value - initial money
	return currentValue - trader.InitialMoney
}

func (e *Exchange) PlaceOrder(order *models.Order) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate order
	if err := e.validateOrder(order); err != nil {
		return err
	}

	// Add to order book
	if order.Type == models.Buy {
		e.BuyOrders[order.StockID] = append(e.BuyOrders[order.StockID], order)
	} else {
		e.SellOrders[order.StockID] = append(e.SellOrders[order.StockID], order)
	}

	// Try to match orders
	e.matchOrders(order.StockID)

	return nil
}

func (e *Exchange) validateOrder(order *models.Order) error {
	trader, exists := e.Traders[order.TraderID]
	if !exists {
		return fmt.Errorf("trader not found")
	}

	// Check for conflicting orders
	if e.hasConflictingOrder(order.TraderID, order.StockID, order.Type) {
		return fmt.Errorf("cannot have both buy and sell orders for the same stock")
	}

	// Validate funds/holdings
	if order.Type == models.Buy {
		if trader.Money < order.Price*float64(order.Quantity) {
			return fmt.Errorf("insufficient funds")
		}
	} else {
		// Check actual holdings minus pending sell orders
		holdings := trader.Holdings[order.StockID]
		pendingSells := 0

		for _, sellOrder := range e.SellOrders[order.StockID] {
			if sellOrder.TraderID == order.TraderID && sellOrder.Status == models.Open {
				pendingSells += sellOrder.Quantity
			}
		}

		availableShares := holdings - pendingSells
		if availableShares < order.Quantity {
			return fmt.Errorf("insufficient holdings")
		}
	}

	return nil
}

func (e *Exchange) hasConflictingOrder(traderID, stockID string, orderType models.OrderType) bool {
	if orderType == models.Buy {
		for _, order := range e.SellOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				return true
			}
		}
	} else {
		for _, order := range e.BuyOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				return true
			}
		}
	}
	return false
}

func (e *Exchange) matchOrders(stockID string) {
	buyOrders := e.BuyOrders[stockID]
	sellOrders := e.SellOrders[stockID]

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
				e.executeTransaction(buyOrder, sellOrder, quantity, sellOrder.Price)
			}
		}
	}

	// Clean up filled orders
	e.cleanupOrders(stockID)
}

func (e *Exchange) executeTransaction(buyOrder, sellOrder *models.Order, quantity int, price float64) {
	// Update order quantities
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
		buyer := e.Traders[buyOrder.TraderID]
		buyer.Money -= price * float64(quantity)
		buyer.Holdings[buyOrder.StockID] += quantity
	}

	if sellOrder.TraderID != "exchange" {
		seller := e.Traders[sellOrder.TraderID]
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
	e.Transactions = append(e.Transactions, tx)

	// Update stock price
	if stock, exists := e.Stocks[buyOrder.StockID]; exists {
		stock.CurrentPrice = price
	}

	log.Printf("Transaction: %s bought %d shares of %s from %s at $%.2f",
		buyOrder.TraderID, quantity, buyOrder.StockID, sellOrder.TraderID, price)
}

func (e *Exchange) cleanupOrders(stockID string) {
	// Remove filled buy orders
	activeBuyOrders := make([]*models.Order, 0)
	for _, order := range e.BuyOrders[stockID] {
		if order.Status == models.Open {
			activeBuyOrders = append(activeBuyOrders, order)
		}
	}
	e.BuyOrders[stockID] = activeBuyOrders

	// Remove filled sell orders
	activeSellOrders := make([]*models.Order, 0)
	for _, order := range e.SellOrders[stockID] {
		if order.Status == models.Open {
			activeSellOrders = append(activeSellOrders, order)
		}
	}
	e.SellOrders[stockID] = activeSellOrders
}

func (e *Exchange) CancelOrder(orderID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Search in buy orders
	for stockID, orders := range e.BuyOrders {
		for i, order := range orders {
			if order.ID == orderID && order.Status == models.Open {
				order.Status = models.Cancelled
				e.BuyOrders[stockID] = append(orders[:i], orders[i+1:]...)
				return nil
			}
		}
	}

	// Search in sell orders
	for stockID, orders := range e.SellOrders {
		for i, order := range orders {
			if order.ID == orderID && order.Status == models.Open {
				order.Status = models.Cancelled
				e.SellOrders[stockID] = append(orders[:i], orders[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("order not found or already closed")
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
