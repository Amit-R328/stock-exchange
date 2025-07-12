package services

import (
	"stock-exchange/internal/models"
	"testing"
	"time"
)

func TestNewAlgorithmManager(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	if am == nil {
		t.Fatal("Expected algorithm manager to be created, got nil")
	}

	if am.exchange != exchange {
		t.Error("Expected algorithm manager to reference the provided exchange")
	}

	if len(am.traders) == 0 {
		t.Error("Expected algorithm manager to have default traders")
	}

	// Check that default traders are created
	foundMomentum := false
	foundContrarian := false

	for _, trader := range am.traders {
		if trader.Strategy == "momentum" && trader.Name == "Momentum Hunter 🚀" {
			foundMomentum = true
		}
		if trader.Strategy == "contrarian" && trader.Name == "Contrarian Trader 📉" {
			foundContrarian = true
		}
	}

	if !foundMomentum {
		t.Error("Expected to find Momentum Hunter bot")
	}

	if !foundContrarian {
		t.Error("Expected to find Contrarian Trader bot")
	}
}

func TestAlgorithmManager_Start_And_Stop(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	// Test starting the algorithm manager
	if am.running {
		t.Error("Expected algorithm manager to be stopped initially")
	}

	am.Start()

	if !am.running {
		t.Error("Expected algorithm manager to be running after start")
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Test stopping
	am.Stop()

	if am.running {
		t.Error("Expected algorithm manager to be stopped after stop")
	}
}

func TestAlgorithmManager_StartAlgorithm(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	// Initially bots should be inactive
	for _, trader := range am.traders {
		if trader.Active {
			t.Errorf("Expected trader %s to be inactive initially", trader.Name)
		}
	}

	// Activate a bot
	if len(am.traders) > 0 {
		botID := am.traders[0].ID
		err := am.StartAlgorithm(botID)

		if err != nil {
			t.Errorf("Expected no error starting algorithm, got: %v", err)
		}

		// Check that the bot is now active
		activated := false
		for _, trader := range am.traders {
			if trader.ID == botID && trader.Active {
				activated = true
				break
			}
		}

		if !activated {
			t.Error("Expected bot to be activated")
		}
	}
}

func TestAlgorithmManager_StopAlgorithm(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	// Activate a bot first
	if len(am.traders) > 0 {
		botID := am.traders[0].ID
		am.StartAlgorithm(botID)

		// Then deactivate it
		err := am.StopAlgorithm(botID)

		if err != nil {
			t.Errorf("Expected no error stopping algorithm, got: %v", err)
		}

		// Check that the bot is now inactive
		for _, trader := range am.traders {
			if trader.ID == botID && trader.Active {
				t.Error("Expected bot to be deactivated")
			}
		}
	}
}

func TestAlgorithmManager_ToggleTrader(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	if len(am.traders) > 0 {
		trader := am.traders[0]
		originalStatus := trader.Active

		// Toggle the trader
		err := am.ToggleTrader(trader.ID)
		if err != nil {
			t.Errorf("Expected no error toggling trader, got: %v", err)
		}

		// Status should be flipped
		if trader.Active == originalStatus {
			t.Error("Expected trader status to be toggled")
		}

		// Toggle back
		err = am.ToggleTrader(trader.ID)
		if err != nil {
			t.Errorf("Expected no error toggling trader back, got: %v", err)
		}

		// Should be back to original status
		if trader.Active != originalStatus {
			t.Error("Expected trader status to be toggled back")
		}
	}
}

func TestAlgorithmManager_GetAlgoTraders(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	traders := am.GetAlgoTraders()

	if len(traders) == 0 {
		t.Error("Expected to get algorithmic traders")
	}

	// Should return the same traders as internal list
	if len(traders) != len(am.traders) {
		t.Errorf("Expected %d traders, got %d", len(am.traders), len(traders))
	}
}

func TestAlgorithmManager_GetAlgorithm(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	if len(am.traders) > 0 {
		expectedTrader := am.traders[0]
		foundTrader := am.GetAlgorithm(expectedTrader.ID)

		if foundTrader == nil {
			t.Error("Expected to find trader by ID")
		}

		if foundTrader.ID != expectedTrader.ID {
			t.Errorf("Expected trader ID %s, got %s", expectedTrader.ID, foundTrader.ID)
		}

		// Test with non-existent ID
		notFound := am.GetAlgorithm("non-existent-id")
		if notFound != nil {
			t.Error("Expected nil for non-existent trader ID")
		}
	}
}

func TestAlgoTrader_Configuration(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	for _, trader := range am.traders {
		// Test that configurations are reasonable
		if trader.Config.MaxOrderValue <= 0 {
			t.Errorf("Trader %s has invalid MaxOrderValue: %f", trader.Name, trader.Config.MaxOrderValue)
		}

		if trader.Config.MinOrderValue <= 0 {
			t.Errorf("Trader %s has invalid MinOrderValue: %f", trader.Name, trader.Config.MinOrderValue)
		}

		if trader.Config.MaxOrderValue <= trader.Config.MinOrderValue {
			t.Errorf("Trader %s has MaxOrderValue <= MinOrderValue", trader.Name)
		}

		if trader.Config.RiskThreshold <= 0 || trader.Config.RiskThreshold > 1 {
			t.Errorf("Trader %s has invalid RiskThreshold: %f", trader.Name, trader.Config.RiskThreshold)
		}

		if trader.Config.CooldownSeconds <= 0 {
			t.Errorf("Trader %s has invalid CooldownSeconds: %d", trader.Name, trader.Config.CooldownSeconds)
		}

		if trader.InitialMoney <= 0 {
			t.Errorf("Trader %s has invalid InitialMoney: %f", trader.Name, trader.InitialMoney)
		}
	}
}

func TestAlgoTrader_GetRecentPrices(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	if len(am.traders) == 0 {
		t.Fatal("No traders found for testing")
	}

	trader := am.traders[0]

	// Test with existing stock
	prices := trader.getRecentPrices("1", 5)
	if len(prices) != 5 {
		t.Errorf("Expected 5 price points, got %d", len(prices))
	}

	// Test that prices are realistic
	basePrice := exchange.stocks["1"].CurrentPrice
	for i, price := range prices {
		if price <= 0 {
			t.Errorf("Price at index %d should be positive, got %f", i, price)
		}

		// Prices should be within reasonable range of base price
		if price < basePrice*0.8 || price > basePrice*1.2 {
			t.Errorf("Price at index %d (%f) is outside reasonable range of base price (%f)", i, price, basePrice)
		}
	}

	// Test with different count
	morePrices := trader.getRecentPrices("1", 10)
	if len(morePrices) != 10 {
		t.Errorf("Expected 10 price points, got %d", len(morePrices))
	}
}

func TestAlgoTrader_CalculateOrderQuantity(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	if len(am.traders) == 0 {
		t.Fatal("No traders found for testing")
	}

	trader := am.traders[0]
	stock := exchange.stocks["1"]

	// Test buy order quantity calculation
	quantity := trader.calculateOrderQuantity(stock, "buy")

	// Should return reasonable quantity (positive for buy orders)
	if quantity < 0 {
		t.Errorf("Expected non-negative quantity, got %d", quantity)
	}

	// Should respect risk threshold
	orderValue := float64(quantity) * stock.CurrentPrice
	maxExpectedValue := trader.InitialMoney * trader.Config.RiskThreshold

	if orderValue > maxExpectedValue*1.1 { // Allow 10% tolerance
		t.Errorf("Order value (%f) exceeds risk threshold (%f)", orderValue, maxExpectedValue)
	}
}

func TestAlgoTrader_Strategies(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	strategies := make(map[string]bool)

	for _, trader := range am.traders {
		strategies[trader.Strategy] = true

		// Test that strategy is one of the known types
		validStrategies := map[string]bool{
			"momentum":   true,
			"contrarian": true,
		}

		if !validStrategies[trader.Strategy] {
			t.Errorf("Unknown strategy '%s' for trader %s", trader.Strategy, trader.Name)
		}
	}

	// Should have multiple different strategies
	if len(strategies) < 2 {
		t.Error("Expected multiple different trading strategies")
	}
}

func TestAlgoTrader_TradingIntegration(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	// Register the algorithmic traders with the exchange
	for _, trader := range am.traders {
		// Create exchange trader for the algorithmic trader
		exchangeTrader := &models.Trader{
			ID:       trader.ID,
			Name:     trader.Name,
			Money:    trader.InitialMoney,
			Holdings: make(map[string]int),
		}
		exchange.traders[trader.ID] = exchangeTrader
	}

	// Count initial orders
	initialOrders := len(exchange.GetOpenOrders("1"))

	// Start the algorithm manager
	am.Start()

	// Activate first trader
	if len(am.traders) > 0 {
		am.StartAlgorithm(am.traders[0].ID)
	}

	// Let it run briefly
	time.Sleep(150 * time.Millisecond)

	// Stop the algorithm manager
	am.Stop()

	// Check final orders
	finalOrders := len(exchange.GetOpenOrders("1"))

	// Log results (may or may not have placed orders due to timing/conditions)
	t.Logf("Initial orders: %d, Final orders: %d", initialOrders, finalOrders)

	// Verify that the algorithmic trader is registered in exchange
	for _, trader := range am.traders {
		exchangeTrader, exists := exchange.GetTrader(trader.ID)
		if !exists {
			t.Errorf("Expected trader %s to be registered in exchange", trader.Name)
		} else {
			t.Logf("Trader %s found in exchange with balance $%.2f", exchangeTrader.Name, exchangeTrader.Money)
		}
	}
}

func TestAlgoTrader_PerformanceTracking(t *testing.T) {
	exchange := createTestExchange()
	am := NewAlgorithmManager(exchange)

	if len(am.traders) == 0 {
		t.Fatal("No traders found for testing")
	}

	trader := am.traders[0]

	// Test initial state
	if trader.OrdersPlaced != 0 {
		t.Errorf("Expected 0 initial orders placed, got %d", trader.OrdersPlaced)
	}

	if trader.ProfitLoss != 0 {
		t.Errorf("Expected 0 initial profit/loss, got %f", trader.ProfitLoss)
	}

	// Verify trader has reasonable initial money
	if trader.InitialMoney <= 0 {
		t.Errorf("Expected positive initial money, got %f", trader.InitialMoney)
	}

	// Verify LastAction is initially zero
	if !trader.LastAction.IsZero() {
		t.Error("Expected LastAction to be zero initially")
	}
}
