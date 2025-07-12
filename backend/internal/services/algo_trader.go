package services

import (
	"fmt"
	"log"
	"math"
	"time"

	"stock-exchange/internal/models"
)

// AlgoTrader represents an algorithmic trading bot
type AlgoTrader struct {
	ID       string
	Name     string
	Strategy string
	Active   bool
	Exchange *Exchange
	Config   AlgoConfig

	// Trading state
	LastAction   time.Time
	OrdersPlaced int
	ProfitLoss   float64
	InitialMoney float64
}

// AlgoConfig contains configuration for the algorithmic trader
type AlgoConfig struct {
	MaxOrderValue     float64 // Maximum value per order
	MinOrderValue     float64 // Minimum value per order
	RiskThreshold     float64 // Maximum risk as percentage of portfolio
	CooldownSeconds   int     // Seconds to wait between orders for same stock
	MomentumThreshold float64 // Momentum percentage threshold for momentum strategy
	ContrarianSpread  float64 // Price deviation percentage for contrarian strategy
}

// AlgorithmManager manages all algorithmic traders
type AlgorithmManager struct {
	traders  []*AlgoTrader
	exchange *Exchange
	running  bool
	ticker   *time.Ticker
}

// NewAlgorithmManager creates a new algorithm manager
func NewAlgorithmManager(exchange *Exchange) *AlgorithmManager {
	am := &AlgorithmManager{
		exchange: exchange,
		traders:  make([]*AlgoTrader, 0),
	}

	// Initialize some basic algorithmic traders
	am.traders = append(am.traders, &AlgoTrader{
		ID:       "momentum-bot-1",
		Name:     "Momentum Hunter 🚀",
		Strategy: "momentum",
		Active:   true, // Auto-start for demo
		Exchange: exchange,
		Config: AlgoConfig{
			MaxOrderValue:     5000,
			MinOrderValue:     100,
			RiskThreshold:     0.1, // 10% of portfolio
			CooldownSeconds:   30,
			MomentumThreshold: 0.025, // 2.5% momentum threshold (more aggressive)
			ContrarianSpread:  0.005, // Not used for momentum strategy
		},
		InitialMoney: 50000,
	})

	am.traders = append(am.traders, &AlgoTrader{
		ID:       "contrarian-bot-1",
		Name:     "Contrarian Trader 📉",
		Strategy: "contrarian",
		Active:   true, // Auto-start for demo
		Exchange: exchange,
		Config: AlgoConfig{
			MaxOrderValue:     3000,
			MinOrderValue:     200,
			RiskThreshold:     0.15, // 15% of portfolio
			CooldownSeconds:   45,
			MomentumThreshold: 0.02,  // Not used for contrarian strategy
			ContrarianSpread:  0.007, // 0.7% spread (more conservative than before)
		},
		InitialMoney: 40000,
	})

	return am
}

// Start begins the algorithmic trading
func (am *AlgorithmManager) Start() {
	if am.running {
		return
	}

	am.running = true
	am.ticker = time.NewTicker(5 * time.Second) // Execute every 5 seconds for faster testing

	log.Println("🤖 Algorithm Manager started - bots are now trading!")

	go func() {
		for {
			select {
			case <-am.ticker.C:
				if am.running {
					log.Printf("🕐 Algorithm Manager tick - executing strategies...")
					am.executeAllStrategies()
				}
			}
		}
	}()
}

// Stop halts algorithmic trading
func (am *AlgorithmManager) Stop() {
	am.running = false
	if am.ticker != nil {
		am.ticker.Stop()
	}
	log.Println("🛑 Algorithm Manager stopped")
}

// executeAllStrategies runs all active trading algorithms
func (am *AlgorithmManager) executeAllStrategies() {
	log.Printf("🔄 Executing strategies for %d traders", len(am.traders))
	for _, trader := range am.traders {
		if trader.Active {
			log.Printf("🤖 Running %s strategy for %s", trader.Strategy, trader.Name)
			switch trader.Strategy {
			case "momentum":
				trader.executeMomentumStrategy()
			case "contrarian":
				trader.executeContrarianStrategy()
			}
		}
	}
}

// executeMomentumStrategy implements a simple momentum trading strategy
func (at *AlgoTrader) executeMomentumStrategy() {
	stocks := at.Exchange.GetAllStocks()
	log.Printf("📈 %s analyzing %d stocks for momentum opportunities", at.Name, len(stocks))

	for _, stock := range stocks {
		// Get recent price history (simulated)
		priceHistory := at.getRecentPrices(stock.ID, 5)
		if len(priceHistory) < 2 {
			log.Printf("⚠️  %s: Not enough price history for %s", at.Name, stock.ID)
			continue
		}

		// Calculate momentum (price change percentage)
		momentum := (priceHistory[len(priceHistory)-1] - priceHistory[0]) / priceHistory[0]
		log.Printf("📊 %s: %s momentum = %.4f%% (price: $%.2f, threshold: %.2f%%)", at.Name, stock.ID, momentum*100, stock.CurrentPrice, at.Config.MomentumThreshold*100)

		// Strong upward momentum - BUY
		if momentum > at.Config.MomentumThreshold {
			quantity := at.calculateOrderQuantity(stock, "buy")
			if quantity > 0 {
				log.Printf("🚀 %s: BUYING %s! Strong upward momentum %.2f%% (above %.2f%% threshold)", at.Name, stock.ID, momentum*100, at.Config.MomentumThreshold*100)
				at.placeBuyOrder(stock.ID, quantity, stock.CurrentPrice*1.001) // Slightly above market
			} else {
				log.Printf("💰 %s: Would buy %s but insufficient funds", at.Name, stock.ID)
			}
		}

		// Strong downward momentum and I have holdings - SELL
		if momentum < -at.Config.MomentumThreshold {
			holdings := at.getCurrentHoldings(stock.ID)
			if holdings > 0 {
				sellQuantity := int(math.Min(float64(holdings), float64(holdings)/2)) // Sell half
				if sellQuantity > 0 {
					log.Printf("📉 %s: SELLING %s! Strong downward momentum %.2f%% (below -%.2f%% threshold)", at.Name, stock.ID, momentum*100, at.Config.MomentumThreshold*100)
					at.placeSellOrder(stock.ID, sellQuantity, stock.CurrentPrice*0.999) // Slightly below market
				}
			}
		}
	}
}

// executeContrarianStrategy implements a contrarian (buy low, sell high) strategy
func (at *AlgoTrader) executeContrarianStrategy() {
	stocks := at.Exchange.GetAllStocks()
	log.Printf("📉 %s analyzing %d stocks for contrarian opportunities", at.Name, len(stocks))

	for _, stock := range stocks {
		priceHistory := at.getRecentPrices(stock.ID, 10)
		if len(priceHistory) < 5 {
			log.Printf("⚠️  %s: Not enough price history for %s", at.Name, stock.ID)
			continue
		}

		// Calculate simple moving average
		avgPrice := 0.0
		for _, price := range priceHistory {
			avgPrice += price
		}
		avgPrice /= float64(len(priceHistory))

		currentPrice := stock.CurrentPrice
		priceVsAvg := (currentPrice - avgPrice) / avgPrice
		log.Printf("📊 %s: %s current=$%.2f vs avg=$%.2f (%.2f%%, threshold: ±%.2f%%)", at.Name, stock.ID, currentPrice, avgPrice, priceVsAvg*100, at.Config.ContrarianSpread*100)

		// Price is significantly below average - BUY (contrarian)
		buyThreshold := 1.0 - at.Config.ContrarianSpread
		if currentPrice < avgPrice*buyThreshold {
			quantity := at.calculateOrderQuantity(stock, "buy")
			if quantity > 0 {
				discount := (1 - currentPrice/avgPrice) * 100
				log.Printf("📉 %s: BUYING %s! Price below average (%.2f%% discount, threshold: %.2f%%)", at.Name, stock.ID, discount, at.Config.ContrarianSpread*100)
				at.placeBuyOrder(stock.ID, quantity, currentPrice*1.002)
			} else {
				log.Printf("💰 %s: Would buy %s but insufficient funds", at.Name, stock.ID)
			}
		}

		// Price is significantly above average - SELL (contrarian)
		sellThreshold := 1.0 + at.Config.ContrarianSpread
		if currentPrice > avgPrice*sellThreshold {
			holdings := at.getCurrentHoldings(stock.ID)
			if holdings > 0 {
				sellQuantity := int(math.Min(float64(holdings), float64(holdings)/3)) // Sell third
				if sellQuantity > 0 {
					premium := (currentPrice/avgPrice - 1) * 100
					log.Printf("📈 %s: SELLING %s! Price above average (%.2f%% premium, threshold: %.2f%%)", at.Name, stock.ID, premium, at.Config.ContrarianSpread*100)
					at.placeSellOrder(stock.ID, sellQuantity, currentPrice*0.998)
				}
			}
		}
	}
}

// Helper methods

func (at *AlgoTrader) getRecentPrices(stockID string, count int) []float64 {
	// In a real implementation, this would query historical data
	// For now, I will simulate with stock-specific variations around current price
	stock, exists := at.Exchange.GetStock(stockID)
	if !exists {
		return []float64{}
	}

	prices := make([]float64, count)
	basePrice := stock.CurrentPrice

	// Create stock-specific seed for variation
	// Each stock gets a unique "personality" by converting its ID to a number
	// Example: Stock "1" → ASCII 49, Stock "AAPL" → sum of A(65)+A(65)+P(80)+L(76) = 286
	stockSeed := 0
	for _, char := range stockID {
		stockSeed += int(char) // Sum ASCII values of characters in stock ID
	}

	// Add current time for dynamic changes
	timeSeed := time.Now().Unix() / 30 // Changes every 30 seconds

	for i := 0; i < count; i++ {
		// Create three different wave patterns that combine to simulate realistic price movements

		// Pattern 1: Basic stock-specific oscillation (±2% range)
		// Each stock has its own unique wave pattern based on its stockSeed
		// i+stockSeed ensures each stock starts its wave from a different point
		variation1 := math.Sin(float64(i+stockSeed)) * 0.02

		// Pattern 2: Time-evolving medium-frequency wave (±1.5% range)
		// i*stockSeed creates different wave speeds for different stocks
		// timeSeed shifts the entire pattern every 30 seconds for dynamic behavior
		variation2 := math.Cos(float64(i*stockSeed+int(timeSeed))) * 0.015

		// Pattern 3: Fast-changing micro-movements (±1% range)
		// i*2 creates higher frequency oscillations (changes twice as fast)
		// timeSeed makes these micro-movements also evolve over time
		variation3 := math.Sin(float64(i*2+int(timeSeed))) * 0.01

		// Combine all three wave patterns to create complex, realistic price movement
		totalVariation := variation1 + variation2 + variation3

		// Safety clamp: Prevent extreme price swings beyond ±5%
		// This ensures no stock can have unrealistic price jumps
		if totalVariation > 0.05 {
			totalVariation = 0.05
		} else if totalVariation < -0.05 {
			totalVariation = -0.05
		}

		// Apply the calculated variation to the base price
		// Example: $100 base price + 3% variation = $103
		prices[i] = basePrice * (1 + totalVariation)
	}

	// Return array of simulated historical prices
	// Index 0 = oldest price, Index (count-1) = newest price
	// This array is used by momentum/contrarian strategies for decision making
	return prices
}

func (at *AlgoTrader) calculateOrderQuantity(stock *models.Stock, orderType string) int {
	trader, exists := at.Exchange.GetTrader(at.ID)
	if !exists {
		log.Printf("❌ %s: Trader not found in exchange! Cannot calculate order quantity", at.Name)
		return 0
	}

	log.Printf("💰 %s: Found trader with $%.2f available", at.Name, trader.Money)

	if orderType == "buy" {
		// Use RiskThreshold percentage of available money
		availableMoney := trader.Money
		orderValue := math.Min(availableMoney*at.Config.RiskThreshold, at.Config.MaxOrderValue) // RiskThreshold% of money or max order value
		orderValue = math.Max(orderValue, at.Config.MinOrderValue)

		quantity := int(orderValue / stock.CurrentPrice)
		log.Printf("📊 %s: Calculated buy quantity for %s: %d shares (value: $%.2f)", at.Name, stock.ID, quantity, orderValue)
		return quantity
	}

	return 0
}

func (at *AlgoTrader) getCurrentHoldings(stockID string) int {
	trader, exists := at.Exchange.GetTrader(at.ID)
	if !exists {
		return 0
	}

	return trader.Holdings[stockID]
}

func (at *AlgoTrader) placeBuyOrder(stockID string, quantity int, price float64) {
	if quantity <= 0 {
		return
	}

	order := &models.Order{
		ID:        fmt.Sprintf("algo-%s-%d", at.ID, time.Now().UnixNano()),
		TraderID:  at.ID,
		StockID:   stockID,
		Type:      models.Buy,
		Price:     price,
		Quantity:  quantity,
		Status:    models.Open,
		CreatedAt: time.Now(),
	}

	err := at.Exchange.PlaceOrder(order)
	if err == nil {
		at.OrdersPlaced++
		at.LastAction = time.Now()
		log.Printf("🤖 %s placed BUY order: %d shares of %s at $%.2f",
			at.Name, quantity, stockID, price)
	} else {
		log.Printf("❌ %s failed to place BUY order: %s", at.Name, err.Error())
	}
}

func (at *AlgoTrader) placeSellOrder(stockID string, quantity int, price float64) {
	if quantity <= 0 {
		return
	}

	order := &models.Order{
		ID:        fmt.Sprintf("algo-%s-%d", at.ID, time.Now().UnixNano()),
		TraderID:  at.ID,
		StockID:   stockID,
		Type:      models.Sell,
		Price:     price,
		Quantity:  quantity,
		Status:    models.Open,
		CreatedAt: time.Now(),
	}

	err := at.Exchange.PlaceOrder(order)
	if err == nil {
		at.OrdersPlaced++
		at.LastAction = time.Now()
		log.Printf("🤖 %s placed SELL order: %d shares of %s at $%.2f",
			at.Name, quantity, stockID, price)
	} else {
		log.Printf("❌ %s failed to place SELL order: %s", at.Name, err.Error())
	}
}

// GetAlgoTraders returns all algorithmic traders
func (am *AlgorithmManager) GetAlgoTraders() []*AlgoTrader {
	return am.traders
}

// ToggleTrader activates/deactivates an algorithmic trader
func (am *AlgorithmManager) ToggleTrader(traderID string) error {
	for _, trader := range am.traders {
		if trader.ID == traderID {
			trader.Active = !trader.Active
			status := "STOPPED"
			if trader.Active {
				status = "STARTED"
			}
			log.Printf("🤖 Algorithmic trader %s %s", trader.Name, status)
			return nil
		}
	}
	return fmt.Errorf("algorithmic trader not found")
}

// StartAlgorithm activates a specific algorithmic trader
func (am *AlgorithmManager) StartAlgorithm(traderID string) error {
	for _, trader := range am.traders {
		if trader.ID == traderID {
			if trader.Active {
				return fmt.Errorf("algorithm is already running")
			}
			trader.Active = true
			log.Printf("🤖 Algorithmic trader %s STARTED", trader.Name)
			return nil
		}
	}
	return fmt.Errorf("algorithmic trader not found")
}

// StopAlgorithm deactivates a specific algorithmic trader
func (am *AlgorithmManager) StopAlgorithm(traderID string) error {
	for _, trader := range am.traders {
		if trader.ID == traderID {
			if !trader.Active {
				return fmt.Errorf("algorithm is not running")
			}
			trader.Active = false
			log.Printf("🤖 Algorithmic trader %s STOPPED", trader.Name)
			return nil
		}
	}
	return fmt.Errorf("algorithmic trader not found")
}

// GetAlgorithm returns a specific algorithmic trader
func (am *AlgorithmManager) GetAlgorithm(traderID string) *AlgoTrader {
	for _, trader := range am.traders {
		if trader.ID == traderID {
			return trader
		}
	}
	return nil
}
