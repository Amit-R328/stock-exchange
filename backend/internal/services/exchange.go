package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"stock-exchange/internal/models"
	"strconv"
	"sync"
	"time"
)

type Subscription struct {
	ch   chan Update
	done chan struct{}
}

type Exchange struct {
	stocks        map[string]*models.Stock
	traders       map[string]*models.Trader
	buyOrders     map[string][]*models.Order
	sellOrders    map[string][]*models.Order
	transactions  []models.Transaction
	subscriptions map[*Subscription]bool
	mu            sync.RWMutex
}

func NewExchange() *Exchange {
	return &Exchange{
		stocks:        make(map[string]*models.Stock),
		traders:       make(map[string]*models.Trader),
		buyOrders:     make(map[string][]*models.Order),
		sellOrders:    make(map[string][]*models.Order),
		transactions:  make([]models.Transaction, 0),
		subscriptions: make(map[*Subscription]bool),
	}
}

func (e *Exchange) LoadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var config struct {
		Shares  []models.Stock  `json:"shares"` // Changed from "stocks"
		Traders []models.Trader `json:"traders"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	// Load stocks
	for i := range config.Shares {
		stock := &config.Shares[i] // Get pointer to original struct
		e.stocks[stock.ID] = stock
		e.buyOrders[stock.ID] = make([]*models.Order, 0)
		e.sellOrders[stock.ID] = make([]*models.Order, 0)

		// Create initial sell orders from exchange
		initialOrder := &models.Order{
			ID:        fmt.Sprintf("init-%s-%d", stock.ID, time.Now().Unix()),
			TraderID:  "exchange",
			StockID:   stock.ID,
			Type:      models.Sell,
			Price:     stock.CurrentPrice,
			Quantity:  stock.Amount,
			Status:    models.Open,
			CreatedAt: time.Now(),
		}
		e.sellOrders[stock.ID] = append(e.sellOrders[stock.ID], initialOrder)
	}

	// Load traders
	for i := range config.Traders {
		trader := &config.Traders[i] // Get pointer to original struct
		t := models.NewTrader(trader.ID, trader.Name, trader.Money)
		e.traders[t.ID] = t
	}

	log.Println("Exchange loaded successfully")
	return nil
}

func (e *Exchange) PlaceOrder(order *models.Order) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate order
	if err := e.validateOrder(order); err != nil {
		return err
	}

	// Add order to appropriate book
	if order.Type == models.Buy {
		e.buyOrders[order.StockID] = append(e.buyOrders[order.StockID], order)
	} else {
		e.sellOrders[order.StockID] = append(e.sellOrders[order.StockID], order)
	}

	// Try to match orders
	e.matchOrders(order.StockID)

	return nil
}

func (e *Exchange) validateOrder(order *models.Order) error {
	// Basic validation
	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if order.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	trader, exists := e.traders[order.TraderID]
	if !exists && order.TraderID != "exchange" {
		return fmt.Errorf("trader not found")
	}

	if order.Type == models.Buy && trader != nil {
		requiredMoney := order.Price * float64(order.Quantity)
		if trader.Money < requiredMoney {
			return fmt.Errorf("insufficient funds")
		}
	}

	if order.Type == models.Sell && trader != nil {
		holdings := trader.Holdings[order.StockID]

		// Calculate shares already committed in pending sell orders
		pendingSellQuantity := 0
		for _, sellOrder := range e.sellOrders[order.StockID] {
			if sellOrder.TraderID == order.TraderID && sellOrder.Status == models.Open {
				pendingSellQuantity += sellOrder.Quantity
			}
		}

		// Check if trader has enough available shares (holdings - pending sells)
		availableShares := holdings - pendingSellQuantity
		if availableShares < order.Quantity {
			return fmt.Errorf("insufficient holdings: have %d shares, %d already pending sale, only %d available",
				holdings, pendingSellQuantity, availableShares)
		}
	}

	return nil
}

func (e *Exchange) matchOrders(stockID string) {
	buyOrders := e.buyOrders[stockID]
	sellOrders := e.sellOrders[stockID]

	for _, buyOrder := range buyOrders {
		if buyOrder.Status != models.Open {
			continue
		}

		for _, sellOrder := range sellOrders {
			if sellOrder.Status != models.Open {
				continue
			}

			if sellOrder.TraderID == buyOrder.TraderID {
				continue // Can't trade with yourself
			}

			if buyOrder.Price >= sellOrder.Price {
				// Execute trade at the buyer's price (since buyer is willing to pay more)
				quantity := min(buyOrder.Quantity, sellOrder.Quantity)
				executionPrice := buyOrder.Price
				e.executeTrade(buyOrder, sellOrder, quantity, executionPrice)

				if buyOrder.Quantity == 0 {
					break
				}
			}
		}
	}

	// Clean up filled orders
	e.cleanupOrders(stockID)
}

func (e *Exchange) executeTrade(buyOrder, sellOrder *models.Order, quantity int, price float64) {
	// Update orders
	buyOrder.Quantity -= quantity
	sellOrder.Quantity -= quantity

	if buyOrder.Quantity == 0 {
		buyOrder.Status = models.Filled
	}
	if sellOrder.Quantity == 0 {
		sellOrder.Status = models.Filled
	}

	// Create transaction
	transaction := models.Transaction{
		ID:         fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		BuyerID:    buyOrder.TraderID,
		SellerID:   sellOrder.TraderID,
		StockID:    buyOrder.StockID,
		Price:      price,
		Quantity:   quantity,
		ExecutedAt: time.Now(),
	}
	e.transactions = append(e.transactions, transaction)

	// Update stock price
	if stock, exists := e.stocks[buyOrder.StockID]; exists {
		stock.SetPrice(price)
	}

	// Update trader holdings and money
	if buyOrder.TraderID != "exchange" {
		buyer := e.traders[buyOrder.TraderID]
		buyer.Money -= price * float64(quantity)
		buyer.Holdings[buyOrder.StockID] += quantity
	} else {
		// If exchange is buying (someone selling back), increase available stock amount
		if stock, exists := e.stocks[sellOrder.StockID]; exists {
			stock.Amount += quantity
			log.Printf("Exchange bought %d shares of %s, total available: %d", quantity, stock.ID, stock.Amount)
		}
	}

	if sellOrder.TraderID != "exchange" {
		seller := e.traders[sellOrder.TraderID]
		seller.Money += price * float64(quantity)
		seller.Holdings[sellOrder.StockID] -= quantity
	} else {
		// If exchange is selling (initial stock supply), reduce available stock amount
		if stock, exists := e.stocks[buyOrder.StockID]; exists {
			stock.Amount -= quantity
			log.Printf("Exchange sold %d shares of %s, remaining: %d", quantity, stock.ID, stock.Amount)
		}
	}

	log.Printf("Trade executed: %s bought %d shares of %s from %s at %.2f",
		buyOrder.TraderID, quantity, buyOrder.StockID, sellOrder.TraderID, price)
}

func (e *Exchange) cleanupOrders(stockID string) {
	// Remove filled buy orders
	activeBuyOrders := make([]*models.Order, 0)
	for _, order := range e.buyOrders[stockID] {
		if order.Status == models.Open {
			activeBuyOrders = append(activeBuyOrders, order)
		}
	}
	e.buyOrders[stockID] = activeBuyOrders

	// Remove filled sell orders
	activeSellOrders := make([]*models.Order, 0)
	for _, order := range e.sellOrders[stockID] {
		if order.Status == models.Open {
			activeSellOrders = append(activeSellOrders, order)
		}
	}
	e.sellOrders[stockID] = activeSellOrders
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Exchange) GetAllStocks() []*models.Stock {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stocks := make([]*models.Stock, 0, len(e.stocks))
	for _, stock := range e.stocks {
		stocks = append(stocks, stock)
	}

	// Sort stocks by ID to ensure consistent order (numeric sorting)
	sort.Slice(stocks, func(i, j int) bool {
		// Convert string IDs to integers for proper numeric sorting
		idI, _ := strconv.Atoi(stocks[i].ID)
		idJ, _ := strconv.Atoi(stocks[j].ID)
		return idI < idJ
	})

	return stocks
}

func (e *Exchange) GetStock(stockID string) (*models.Stock, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stock, exists := e.stocks[stockID]
	return stock, exists
}

func (e *Exchange) GetOpenOrders(stockID string) []models.Order {
	e.mu.RLock()
	defer e.mu.RUnlock()

	orders := make([]models.Order, 0) // Initialize with empty slice instead of nil

	// Add buy orders
	for _, order := range e.buyOrders[stockID] {
		if order.Status == models.Open {
			orders = append(orders, *order)
		}
	}

	// Add sell orders
	for _, order := range e.sellOrders[stockID] {
		if order.Status == models.Open {
			orders = append(orders, *order)
		}
	}

	return orders
}

func (e *Exchange) GetLastTransactions(stockID string, limit int) []models.Transaction {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var stockTransactions []models.Transaction

	// Filter transactions for this stock
	for i := len(e.transactions) - 1; i >= 0 && len(stockTransactions) < limit; i-- {
		if e.transactions[i].StockID == stockID {
			stockTransactions = append(stockTransactions, e.transactions[i])
		}
	}

	return stockTransactions
}

func (e *Exchange) HasConflictingOrder(traderID, stockID string, orderType models.OrderType) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Check opposite order type
	if orderType == models.Buy {
		for _, order := range e.sellOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				return true
			}
		}
	} else {
		for _, order := range e.buyOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				return true
			}
		}
	}

	return false
}

func (e *Exchange) CancelOrder(orderID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Search in all order books
	for stockID := range e.stocks {
		// Check buy orders
		for i, order := range e.buyOrders[stockID] {
			if order.ID == orderID && order.Status == models.Open {
				order.Status = models.Cancelled
				e.buyOrders[stockID] = append(e.buyOrders[stockID][:i], e.buyOrders[stockID][i+1:]...)
				return nil
			}
		}

		// Check sell orders
		for i, order := range e.sellOrders[stockID] {
			if order.ID == orderID && order.Status == models.Open {
				order.Status = models.Cancelled
				e.sellOrders[stockID] = append(e.sellOrders[stockID][:i], e.sellOrders[stockID][i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("order not found or already closed")
}

func (e *Exchange) GetAllTraders() []TraderInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	traders := make([]TraderInfo, 0, len(e.traders))
	for _, trader := range e.traders {
		traders = append(traders, TraderInfo{
			ID:   trader.ID,
			Name: trader.Name,
		})
	}

	// Sort traders by ID to ensure consistent order (numeric sorting)
	sort.Slice(traders, func(i, j int) bool {
		// Convert string IDs to integers for proper numeric sorting
		idI, _ := strconv.Atoi(traders[i].ID)
		idJ, _ := strconv.Atoi(traders[j].ID)
		return idI < idJ
	})

	return traders
}

func (e *Exchange) GetTrader(traderID string) (*models.Trader, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	trader, exists := e.traders[traderID]
	return trader, exists
}

func (e *Exchange) GetTraderOpenOrders(traderID string) []models.Order {
	e.mu.RLock()
	defer e.mu.RUnlock()

	orders := make([]models.Order, 0) // Initialize with empty slice instead of nil

	// Check all stocks
	for stockID := range e.stocks {
		// Buy orders
		for _, order := range e.buyOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				orders = append(orders, *order)
			}
		}

		// Sell orders
		for _, order := range e.sellOrders[stockID] {
			if order.TraderID == traderID && order.Status == models.Open {
				orders = append(orders, *order)
			}
		}
	}

	return orders
}

func (e *Exchange) GetTraderTransactions(traderID string, limit int) []models.Transaction {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var traderTransactions []models.Transaction

	// Filter transactions for this trader
	for i := len(e.transactions) - 1; i >= 0 && len(traderTransactions) < limit; i-- {
		tx := e.transactions[i]
		if tx.BuyerID == traderID || tx.SellerID == traderID {
			traderTransactions = append(traderTransactions, tx)
		}
	}

	return traderTransactions
}

func (e *Exchange) CalculateProfitLoss(traderID string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	trader, exists := e.traders[traderID]
	if !exists {
		return 0
	}

	// Calculate current portfolio value
	portfolioValue := trader.Money

	for stockID, quantity := range trader.Holdings {
		if stock, exists := e.stocks[stockID]; exists {
			portfolioValue += stock.CurrentPrice * float64(quantity)
		}
	}

	return portfolioValue - trader.InitialMoney // Use actual initial money
}

// WebSocket support
type Update struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (e *Exchange) Subscribe() *Subscription {
	sub := &Subscription{
		ch:   make(chan Update, 100),
		done: make(chan struct{}),
	}

	e.mu.Lock()
	e.subscriptions[sub] = true
	e.mu.Unlock()

	// Send periodic updates
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		defer close(sub.ch) // Close channel when goroutine exits

		for {
			select {
			case <-sub.done:
				return
			case <-ticker.C:
				select {
				case sub.ch <- Update{Type: "stocks", Data: e.GetAllStocks()}:
				case <-sub.done:
					return
				default:
					// Channel full, skip update
				}
			}
		}
	}()

	return sub
}

func (e *Exchange) Unsubscribe(sub *Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.subscriptions[sub]; exists {
		delete(e.subscriptions, sub)
		// Close the done channel to stop the goroutine
		select {
		case <-sub.done:
			// Already closed
		default:
			close(sub.done)
		}
		// Don't close the channel here, let the goroutine finish naturally
	}
}

// GetChannel returns the update channel for a subscription
func (sub *Subscription) GetChannel() chan Update {
	return sub.ch
}

// Helper type for GetAllTraders
type TraderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetStockHistory returns historical price data for a stock
func (e *Exchange) GetStockHistory(stockID string, days int) []models.PriceQuote {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// For now, generate sample historical data
	// In a real implementation, this would query a database
	history := make([]models.PriceQuote, days)

	stock, exists := e.stocks[stockID]
	if !exists {
		return history
	}

	currentPrice := stock.GetPrice()

	for i := 0; i < days; i++ {
		// Generate realistic price movement
		change := (0.5 - rand.Float64()) * 10                        // Random change between -5 and +5
		priceVariation := currentPrice * (0.95 + rand.Float64()*0.1) // ±5% variation

		history[i] = models.PriceQuote{
			Timestamp: time.Now().AddDate(0, 0, -(days - i - 1)),
			Price:     priceVariation + change,
			Volume:    50 + rand.Intn(200), // Random volume between 50-250
		}
	}

	return history
}

// GetTraderPerformance returns historical performance data for a trader
func (e *Exchange) GetTraderPerformance(traderID string, days int) []models.PerformanceData {
	e.mu.RLock()
	defer e.mu.RUnlock()

	performance := make([]models.PerformanceData, days)

	trader, exists := e.traders[traderID]
	if !exists {
		return performance
	}

	// Calculate current values
	profitLoss := e.CalculateProfitLoss(traderID)
	portfolioValue := trader.Money + profitLoss

	for i := 0; i < days; i++ {
		// Generate realistic performance progression
		dayProgress := float64(i) / float64(days)
		valueChange := (0.5 - rand.Float64()) * 2000 // Daily variation

		performance[i] = models.PerformanceData{
			Date:           time.Now().AddDate(0, 0, -(days - i - 1)),
			PortfolioValue: portfolioValue*(0.8+dayProgress*0.4) + valueChange,
			ProfitLoss:     profitLoss*dayProgress + valueChange*0.5,
			CashBalance:    trader.Money,
		}
	}

	return performance
}

// GetTraderPortfolio returns current portfolio distribution
func (e *Exchange) GetTraderPortfolio(traderID string) models.PortfolioData {
	e.mu.RLock()
	defer e.mu.RUnlock()

	trader, exists := e.traders[traderID]
	if !exists {
		return models.PortfolioData{}
	}

	holdings := make([]models.PortfolioHolding, 0)
	totalValue := trader.Money

	// Calculate holdings from transactions
	stockHoldings := make(map[string]int)
	for _, tx := range e.transactions {
		if tx.BuyerID == traderID {
			stockHoldings[tx.StockID] += tx.Quantity
		} else if tx.SellerID == traderID {
			stockHoldings[tx.StockID] -= tx.Quantity
		}
	}

	// Convert to holdings with current values
	for stockID, quantity := range stockHoldings {
		if quantity > 0 {
			stock, exists := e.stocks[stockID]
			if exists {
				value := float64(quantity) * stock.GetPrice()
				totalValue += value

				holdings = append(holdings, models.PortfolioHolding{
					StockID:   stockID,
					StockName: stock.Name,
					Quantity:  quantity,
					Value:     value,
				})
			}
		}
	}

	// Calculate percentages
	for i := range holdings {
		holdings[i].Percentage = (holdings[i].Value / totalValue) * 100
	}

	return models.PortfolioData{
		Holdings:    holdings,
		TotalValue:  totalValue,
		CashBalance: trader.Money,
	}
}

// GetTraderActivity returns trading activity over periods
func (e *Exchange) GetTraderActivity(traderID string, months int) []models.ActivityLog {
	e.mu.RLock()
	defer e.mu.RUnlock()

	activity := make([]models.ActivityLog, months)
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for i := 0; i < months; i++ {
		month := time.Now().AddDate(0, -(months - i - 1), 0)
		monthName := monthNames[month.Month()-1]

		buyOrders := 0
		sellOrders := 0
		volume := 0
		value := 0.0

		// Count orders for this trader in this month
		// In a real implementation, this would filter by actual dates
		for _, tx := range e.transactions {
			if tx.BuyerID == traderID || tx.SellerID == traderID {
				if tx.BuyerID == traderID {
					buyOrders++
				} else {
					sellOrders++
				}
				volume += tx.Quantity
				value += tx.Price * float64(tx.Quantity)
			}
		}

		// Add some randomness for demo purposes
		buyOrders += 5 + rand.Intn(20)
		sellOrders += 3 + rand.Intn(15)

		activity[i] = models.ActivityLog{
			Period:     monthName,
			BuyOrders:  buyOrders,
			SellOrders: sellOrders,
			Volume:     volume + 100 + rand.Intn(500),
			Value:      value + 5000 + rand.Float64()*20000,
		}
	}

	return activity
}

// RegisterTrader adds a new trader to the exchange
func (e *Exchange) RegisterTrader(traderID, name string, initialMoney float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if trader already exists
	if _, exists := e.traders[traderID]; exists {
		log.Printf("⚠️  Trader %s already exists", traderID)
		return
	}

	// Create new trader
	trader := models.NewTrader(traderID, name, initialMoney)
	e.traders[traderID] = trader
	log.Printf("✅ Registered new trader: %s (%s) with $%.2f", name, traderID, initialMoney)
}
