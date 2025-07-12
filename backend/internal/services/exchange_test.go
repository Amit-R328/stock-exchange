package services

import (
	"fmt"
	"stock-exchange/internal/models"
	"sync"
	"testing"
	"time"
)

// Test helper functions
func createTestExchange() *Exchange {
	exchange := NewExchange()

	// Add test stocks
	stock1 := &models.Stock{
		ID:           "1",
		Name:         "Apple Inc.",
		CurrentPrice: 150.0,
		Amount:       1000,
	}

	stock2 := &models.Stock{
		ID:           "2",
		Name:         "Microsoft Corp.",
		CurrentPrice: 300.0,
		Amount:       500,
	}

	exchange.stocks["1"] = stock1
	exchange.stocks["2"] = stock2
	exchange.buyOrders["1"] = make([]*models.Order, 0)
	exchange.sellOrders["1"] = make([]*models.Order, 0)
	exchange.buyOrders["2"] = make([]*models.Order, 0)
	exchange.sellOrders["2"] = make([]*models.Order, 0)

	// Add test traders
	trader1 := models.NewTrader("trader1", "John Doe", 10000.0)
	trader2 := models.NewTrader("trader2", "Jane Smith", 15000.0)

	exchange.traders["trader1"] = trader1
	exchange.traders["trader2"] = trader2

	return exchange
}

func createTestOrder(id, traderID, stockID string, orderType models.OrderType, price float64, quantity int) *models.Order {
	return &models.Order{
		ID:        id,
		TraderID:  traderID,
		StockID:   stockID,
		Type:      orderType,
		Price:     price,
		Quantity:  quantity,
		Status:    models.Open,
		CreatedAt: time.Now(),
	}
}

// Test NewExchange
func TestNewExchange(t *testing.T) {
	exchange := NewExchange()

	if exchange == nil {
		t.Fatal("NewExchange() returned nil")
	}

	if exchange.stocks == nil {
		t.Error("stocks map not initialized")
	}

	if exchange.traders == nil {
		t.Error("traders map not initialized")
	}

	if exchange.buyOrders == nil {
		t.Error("buyOrders map not initialized")
	}

	if exchange.sellOrders == nil {
		t.Error("sellOrders map not initialized")
	}

	if exchange.transactions == nil {
		t.Error("transactions slice not initialized")
	}

	if exchange.subscriptions == nil {
		t.Error("subscriptions map not initialized")
	}
}

// Test PlaceOrder - Valid Orders
func TestPlaceOrder_ValidOrders(t *testing.T) {
	exchange := createTestExchange()

	// Test valid buy order
	buyOrder := createTestOrder("buy1", "trader1", "1", models.Buy, 150.0, 10)
	err := exchange.PlaceOrder(buyOrder)
	if err != nil {
		t.Errorf("PlaceOrder() failed for valid buy order: %v", err)
	}

	// Check if order was added to buy orders
	if len(exchange.buyOrders["1"]) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(exchange.buyOrders["1"]))
	}

	// Give trader some shares first
	exchange.traders["trader1"].Holdings["1"] = 20

	// Test valid sell order
	sellOrder := createTestOrder("sell1", "trader1", "1", models.Sell, 155.0, 5)
	err = exchange.PlaceOrder(sellOrder)
	if err != nil {
		t.Errorf("PlaceOrder() failed for valid sell order: %v", err)
	}

	// Check if order was added to sell orders
	if len(exchange.sellOrders["1"]) != 1 {
		t.Errorf("Expected 1 sell order, got %d", len(exchange.sellOrders["1"]))
	}
}

// Test PlaceOrder - Invalid Orders
func TestPlaceOrder_InvalidOrders(t *testing.T) {
	exchange := createTestExchange()

	tests := []struct {
		name          string
		order         *models.Order
		expectedError string
	}{
		{
			name:          "Zero quantity",
			order:         createTestOrder("invalid1", "trader1", "1", models.Buy, 150.0, 0),
			expectedError: "quantity must be greater than 0",
		},
		{
			name:          "Negative quantity",
			order:         createTestOrder("invalid2", "trader1", "1", models.Buy, 150.0, -5),
			expectedError: "quantity must be greater than 0",
		},
		{
			name:          "Zero price",
			order:         createTestOrder("invalid3", "trader1", "1", models.Buy, 0.0, 10),
			expectedError: "price must be greater than 0",
		},
		{
			name:          "Negative price",
			order:         createTestOrder("invalid4", "trader1", "1", models.Buy, -100.0, 10),
			expectedError: "price must be greater than 0",
		},
		{
			name:          "Nonexistent trader",
			order:         createTestOrder("invalid5", "nonexistent", "1", models.Buy, 150.0, 10),
			expectedError: "trader not found",
		},
		{
			name:          "Insufficient funds",
			order:         createTestOrder("invalid6", "trader1", "1", models.Buy, 1000.0, 100),
			expectedError: "insufficient funds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exchange.PlaceOrder(tt.order)
			if err == nil {
				t.Errorf("Expected error '%s', got nil", tt.expectedError)
			} else if err.Error() != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// Test PlaceOrder - Insufficient Holdings
func TestPlaceOrder_InsufficientHoldings(t *testing.T) {
	exchange := createTestExchange()

	// Try to sell shares the trader doesn't have
	sellOrder := createTestOrder("sell1", "trader1", "1", models.Sell, 150.0, 10)
	err := exchange.PlaceOrder(sellOrder)

	if err == nil {
		t.Error("Expected error for insufficient holdings, got nil")
	}

	expectedError := "insufficient holdings: have 0 shares, 0 already pending sale, only 0 available"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Test Order Matching
func TestOrderMatching(t *testing.T) {
	exchange := createTestExchange()

	// Give trader2 some shares to sell
	exchange.traders["trader2"].Holdings["1"] = 50

	// Place a sell order
	sellOrder := createTestOrder("sell1", "trader2", "1", models.Sell, 140.0, 10)
	err := exchange.PlaceOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to place sell order: %v", err)
	}

	// Place a matching buy order (higher price)
	buyOrder := createTestOrder("buy1", "trader1", "1", models.Buy, 150.0, 10)
	err = exchange.PlaceOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}

	// Check if trade was executed
	if len(exchange.transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(exchange.transactions))
	}

	// Check if orders were filled
	if buyOrder.Status != models.Filled {
		t.Errorf("Expected buy order to be filled, got status: %v", buyOrder.Status)
	}

	if sellOrder.Status != models.Filled {
		t.Errorf("Expected sell order to be filled, got status: %v", sellOrder.Status)
	}

	// Check trader balances
	buyer := exchange.traders["trader1"]
	seller := exchange.traders["trader2"]

	expectedBuyerMoney := 10000.0 - (150.0 * 10) // Initial money - purchase cost
	if buyer.Money != expectedBuyerMoney {
		t.Errorf("Expected buyer money: %.2f, got: %.2f", expectedBuyerMoney, buyer.Money)
	}

	expectedSellerMoney := 15000.0 + (150.0 * 10) // Initial money + sale proceeds
	if seller.Money != expectedSellerMoney {
		t.Errorf("Expected seller money: %.2f, got: %.2f", expectedSellerMoney, seller.Money)
	}

	// Check holdings
	if buyer.Holdings["1"] != 10 {
		t.Errorf("Expected buyer to have 10 shares, got: %d", buyer.Holdings["1"])
	}

	if seller.Holdings["1"] != 40 { // 50 - 10 sold
		t.Errorf("Expected seller to have 40 shares, got: %d", seller.Holdings["1"])
	}
}

// Test Partial Order Matching
func TestPartialOrderMatching(t *testing.T) {
	exchange := createTestExchange()

	// Give trader2 some shares
	exchange.traders["trader2"].Holdings["1"] = 50

	// Place a large sell order
	sellOrder := createTestOrder("sell1", "trader2", "1", models.Sell, 140.0, 20)
	err := exchange.PlaceOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to place sell order: %v", err)
	}

	// Place a smaller buy order
	buyOrder := createTestOrder("buy1", "trader1", "1", models.Buy, 150.0, 10)
	err = exchange.PlaceOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}

	// Check partial execution
	if buyOrder.Status != models.Filled {
		t.Errorf("Expected buy order to be filled, got status: %v", buyOrder.Status)
	}

	if sellOrder.Status != models.Open {
		t.Errorf("Expected sell order to remain open, got status: %v", sellOrder.Status)
	}

	if sellOrder.Quantity != 10 { // 20 - 10 executed
		t.Errorf("Expected sell order quantity to be 10, got: %d", sellOrder.Quantity)
	}
}

// Test Self-Trading Prevention
func TestSelfTradingPrevention(t *testing.T) {
	exchange := createTestExchange()

	// Give trader1 some shares
	exchange.traders["trader1"].Holdings["1"] = 50

	// Place a sell order
	sellOrder := createTestOrder("sell1", "trader1", "1", models.Sell, 140.0, 10)
	err := exchange.PlaceOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to place sell order: %v", err)
	}

	// Place a buy order from the same trader
	buyOrder := createTestOrder("buy1", "trader1", "1", models.Buy, 150.0, 10)
	err = exchange.PlaceOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}

	// Check that no trade was executed
	if len(exchange.transactions) != 0 {
		t.Errorf("Expected 0 transactions (self-trading), got %d", len(exchange.transactions))
	}

	// Both orders should remain open
	if buyOrder.Status != models.Open {
		t.Errorf("Expected buy order to remain open, got status: %v", buyOrder.Status)
	}

	if sellOrder.Status != models.Open {
		t.Errorf("Expected sell order to remain open, got status: %v", sellOrder.Status)
	}
}

// Test Price Execution Logic
func TestPriceExecution(t *testing.T) {
	exchange := createTestExchange()

	// Give trader2 some shares
	exchange.traders["trader2"].Holdings["1"] = 50

	// Place a sell order at 140
	sellOrder := createTestOrder("sell1", "trader2", "1", models.Sell, 140.0, 10)
	err := exchange.PlaceOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to place sell order: %v", err)
	}

	// Place a buy order at 150 (higher than sell price)
	buyOrder := createTestOrder("buy1", "trader1", "1", models.Buy, 150.0, 10)
	err = exchange.PlaceOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}

	// Check that trade executed at buyer's price (150)
	if len(exchange.transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(exchange.transactions))
	}

	transaction := exchange.transactions[0]
	if transaction.Price != 150.0 {
		t.Errorf("Expected execution price 150.0, got %.2f", transaction.Price)
	}

	// Check stock price was updated
	stock := exchange.stocks["1"]
	if stock.CurrentPrice != 150.0 {
		t.Errorf("Expected stock price to be updated to 150.0, got %.2f", stock.CurrentPrice)
	}
}

// Test Order Cancellation
func TestCancelOrder(t *testing.T) {
	exchange := createTestExchange()

	// Place a buy order
	buyOrder := createTestOrder("buy1", "trader1", "1", models.Buy, 150.0, 10)
	err := exchange.PlaceOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}

	// Cancel the order
	err = exchange.CancelOrder("buy1")
	if err != nil {
		t.Errorf("Failed to cancel order: %v", err)
	}

	// Check that order was removed from order book
	if len(exchange.buyOrders["1"]) != 0 {
		t.Errorf("Expected 0 buy orders after cancellation, got %d", len(exchange.buyOrders["1"]))
	}

	// Try to cancel non-existent order
	err = exchange.CancelOrder("nonexistent")
	if err == nil {
		t.Error("Expected error when cancelling non-existent order")
	}
}

// Test GetAllStocks
func TestGetAllStocks(t *testing.T) {
	exchange := createTestExchange()

	stocks := exchange.GetAllStocks()

	if len(stocks) != 2 {
		t.Errorf("Expected 2 stocks, got %d", len(stocks))
	}

	// Check that stocks are sorted by ID
	if stocks[0].ID != "1" || stocks[1].ID != "2" {
		t.Error("Stocks not sorted correctly by ID")
	}
}

// Test GetStock
func TestGetStock(t *testing.T) {
	exchange := createTestExchange()

	// Test existing stock
	stock, exists := exchange.GetStock("1")
	if !exists {
		t.Error("Expected stock to exist")
	}
	if stock.ID != "1" {
		t.Errorf("Expected stock ID '1', got '%s'", stock.ID)
	}

	// Test non-existent stock
	_, exists = exchange.GetStock("999")
	if exists {
		t.Error("Expected stock to not exist")
	}
}

// Test Concurrent Order Placement (Race Conditions)
func TestConcurrentOrderPlacement(t *testing.T) {
	exchange := createTestExchange()

	// Give both traders some shares and money
	exchange.traders["trader1"].Holdings["1"] = 100
	exchange.traders["trader2"].Holdings["1"] = 100

	var wg sync.WaitGroup
	errors := make(chan error, 20)

	// Place multiple concurrent orders
	for i := 0; i < 10; i++ {
		wg.Add(2)

		go func(i int) {
			defer wg.Done()
			buyOrder := createTestOrder(
				fmt.Sprintf("buy%d", i),
				"trader1",
				"1",
				models.Buy,
				150.0,
				5,
			)
			if err := exchange.PlaceOrder(buyOrder); err != nil {
				errors <- err
			}
		}(i)

		go func(i int) {
			defer wg.Done()
			sellOrder := createTestOrder(
				fmt.Sprintf("sell%d", i),
				"trader2",
				"1",
				models.Sell,
				145.0,
				5,
			)
			if err := exchange.PlaceOrder(sellOrder); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent order placement error: %v", err)
	}

	// Verify that some trades occurred
	if len(exchange.transactions) == 0 {
		t.Error("Expected some transactions from concurrent orders")
	}
}

// Test WebSocket Subscription
func TestWebSocketSubscription(t *testing.T) {
	exchange := createTestExchange()

	// Subscribe to updates
	sub := exchange.Subscribe()
	if sub == nil {
		t.Fatal("Subscribe() returned nil")
	}

	// Check that subscription was added
	exchange.mu.RLock()
	if len(exchange.subscriptions) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(exchange.subscriptions))
	}
	exchange.mu.RUnlock()

	// Wait for at least one update
	select {
	case update := <-sub.GetChannel():
		if update.Type != "stocks" {
			t.Errorf("Expected update type 'stocks', got '%s'", update.Type)
		}
	case <-time.After(3 * time.Second):
		t.Error("Expected to receive an update within 3 seconds")
	}

	// Unsubscribe
	exchange.Unsubscribe(sub)

	// Check that subscription was removed
	exchange.mu.RLock()
	if len(exchange.subscriptions) != 0 {
		t.Errorf("Expected 0 subscriptions after unsubscribe, got %d", len(exchange.subscriptions))
	}
	exchange.mu.RUnlock()
}

// Test Profit/Loss Calculation
func TestCalculateProfitLoss(t *testing.T) {
	exchange := createTestExchange()

	// Give trader some shares and change stock price
	trader := exchange.traders["trader1"]
	trader.Holdings["1"] = 10 // 10 shares at current price 150.0

	// Calculate P&L (should be 0 initially since portfolio value = initial money)
	initialPL := exchange.CalculateProfitLoss("trader1")
	expectedPL := (10000.0 + (10 * 150.0)) - 10000.0 // current portfolio - initial money
	if initialPL != expectedPL {
		t.Errorf("Expected P&L %.2f, got %.2f", expectedPL, initialPL)
	}

	// Change stock price and recalculate
	exchange.stocks["1"].SetPrice(200.0)
	newPL := exchange.CalculateProfitLoss("trader1")
	expectedNewPL := (10000.0 + (10 * 200.0)) - 10000.0
	if newPL != expectedNewPL {
		t.Errorf("Expected P&L %.2f after price change, got %.2f", expectedNewPL, newPL)
	}
}

// Test Pending Sell Orders Logic
func TestPendingSellOrders(t *testing.T) {
	exchange := createTestExchange()

	// Give trader some shares
	exchange.traders["trader1"].Holdings["1"] = 20

	// Place first sell order for 10 shares
	sellOrder1 := createTestOrder("sell1", "trader1", "1", models.Sell, 150.0, 10)
	err := exchange.PlaceOrder(sellOrder1)
	if err != nil {
		t.Fatalf("Failed to place first sell order: %v", err)
	}

	// Try to place another sell order for 15 shares (should fail - only 10 remaining)
	sellOrder2 := createTestOrder("sell2", "trader1", "1", models.Sell, 155.0, 15)
	err = exchange.PlaceOrder(sellOrder2)
	if err == nil {
		t.Error("Expected error for selling more shares than available (considering pending orders)")
	}

	// Should be able to sell exactly the remaining 10 shares
	sellOrder3 := createTestOrder("sell3", "trader1", "1", models.Sell, 155.0, 10)
	err = exchange.PlaceOrder(sellOrder3)
	if err != nil {
		t.Errorf("Failed to place sell order for remaining shares: %v", err)
	}
}
