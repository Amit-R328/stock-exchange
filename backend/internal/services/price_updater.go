package services

import (
	"log"
	"math/rand"
	"time"
)

type PriceUpdater struct {
	exchange *Exchange
	ticker   *time.Ticker
	done     chan bool
	rng      *rand.Rand
}

func NewPriceUpdater(exchange *Exchange, interval time.Duration) *PriceUpdater {
	// Create a new random generator with time-based seed
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	log.Printf("🎲 Random seed initialized with: %d", seed)

	return &PriceUpdater{
		exchange: exchange,
		ticker:   time.NewTicker(interval),
		done:     make(chan bool),
		rng:      rng,
	}
}

func (pu *PriceUpdater) Start() {
	go func() {
		for {
			select {
			case <-pu.ticker.C:
				pu.updatePrices()
			case <-pu.done:
				return
			}
		}
	}()
}

func (pu *PriceUpdater) Stop() {
	pu.ticker.Stop()
	pu.done <- true
}

func (pu *PriceUpdater) updatePrices() {
	pu.exchange.mu.RLock()
	stocks := pu.exchange.stocks
	pu.exchange.mu.RUnlock()

	log.Printf("🔄 Starting price update for %d stocks...", len(stocks))
	changedCount := 0

	for _, stock := range stocks {
		// Random price change between -2% and +2%
		change := (pu.rng.Float64() - 0.5) * 0.04
		currentPrice := stock.GetPrice()
		newPrice := currentPrice * (1 + change)

		// Round to 2 decimal places
		newPrice = float64(int(newPrice*100+0.5)) / 100

		// Only update if there's actually a change
		if newPrice != currentPrice {
			stock.SetPrice(newPrice)
			changedCount++
			log.Printf("💹 Price updated: %s %.2f -> %.2f (%.2f%%)", stock.ID, currentPrice, newPrice, change*100)
		}
	}

	log.Printf("✅ Price update completed. %d out of %d stocks changed", changedCount, len(stocks))
}
